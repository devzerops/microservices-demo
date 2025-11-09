#!/usr/bin/python
#
# Copyright 2024 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      https://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

import gzip
import logging
import os
import signal
import sys
import time
import threading
from collections import defaultdict
from io import BytesIO
from typing import Dict, List, Tuple

from google.cloud import secretmanager_v1
from urllib.parse import unquote, urlparse
from langchain_core.messages import HumanMessage
from langchain_google_genai import ChatGoogleGenerativeAI, GoogleGenerativeAIEmbeddings
from flask import Flask, request, jsonify

from langchain_google_alloydb_pg import AlloyDBEngine, AlloyDBVectorStore

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

PROJECT_ID = os.environ["PROJECT_ID"]
REGION = os.environ["REGION"]
ALLOYDB_DATABASE_NAME = os.environ["ALLOYDB_DATABASE_NAME"]
ALLOYDB_TABLE_NAME = os.environ["ALLOYDB_TABLE_NAME"]
ALLOYDB_CLUSTER_NAME = os.environ["ALLOYDB_CLUSTER_NAME"]
ALLOYDB_INSTANCE_NAME = os.environ["ALLOYDB_INSTANCE_NAME"]
ALLOYDB_SECRET_NAME = os.environ["ALLOYDB_SECRET_NAME"]

# LLM Model configuration with defaults
LLM_MODEL = os.environ.get("LLM_MODEL", "gemini-1.5-flash")
EMBEDDING_MODEL = os.environ.get("EMBEDDING_MODEL", "models/embedding-001")

# Input validation limits
MAX_MESSAGE_LENGTH = 1000
MAX_IMAGE_URL_LENGTH = 2048

# Rate limiting configuration
# LLM API calls are expensive, so use aggressive rate limiting
# Default: 5 requests per minute per IP (0.083 req/sec)
RATE_LIMIT_REQUESTS = int(os.environ.get('RATE_LIMIT_REQUESTS', '5'))
RATE_LIMIT_WINDOW = int(os.environ.get('RATE_LIMIT_WINDOW', '60'))  # seconds

class RateLimiter:
    """
    Simple in-memory rate limiter using sliding window algorithm.
    Tracks request timestamps per IP address.
    """

    def __init__(self, max_requests: int = RATE_LIMIT_REQUESTS, window_seconds: int = RATE_LIMIT_WINDOW):
        self.max_requests = max_requests
        self.window_seconds = window_seconds
        self.requests: Dict[str, List[float]] = defaultdict(list)
        self.lock = threading.Lock()

        # Start cleanup thread to remove old entries
        self.cleanup_thread = threading.Thread(target=self._cleanup_old_entries, daemon=True)
        self.cleanup_thread.start()

    def is_allowed(self, ip_address: str) -> Tuple[bool, int]:
        """
        Check if request from IP is allowed.
        Returns: (allowed: bool, remaining_requests: int)
        """
        with self.lock:
            now = time.time()
            cutoff = now - self.window_seconds

            # Remove old requests outside the window
            if ip_address in self.requests:
                self.requests[ip_address] = [
                    timestamp for timestamp in self.requests[ip_address]
                    if timestamp > cutoff
                ]

            current_count = len(self.requests[ip_address])

            if current_count < self.max_requests:
                self.requests[ip_address].append(now)
                remaining = self.max_requests - current_count - 1
                return True, remaining
            else:
                return False, 0

    def _cleanup_old_entries(self):
        """Periodically clean up old IP addresses that haven't made requests recently."""
        while True:
            time.sleep(300)  # Run every 5 minutes

            with self.lock:
                now = time.time()
                cutoff = now - (self.window_seconds * 2)  # Remove IPs inactive for 2x window

                ips_to_remove = []
                for ip, timestamps in self.requests.items():
                    # Remove old timestamps
                    timestamps = [t for t in timestamps if t > cutoff]

                    if not timestamps:
                        ips_to_remove.append(ip)
                    else:
                        self.requests[ip] = timestamps

                for ip in ips_to_remove:
                    del self.requests[ip]

                if ips_to_remove:
                    logger.debug(f"Rate limiter cleanup: removed {len(ips_to_remove)} inactive IPs")

# Global rate limiter instance
rate_limiter = RateLimiter()

def get_client_ip() -> str:
    """Extract real client IP from request headers (handles proxies)."""
    # Check X-Forwarded-For header (standard for proxies)
    if request.headers.get('X-Forwarded-For'):
        # X-Forwarded-For can contain multiple IPs, take the first one
        ips = request.headers.get('X-Forwarded-For').split(',')
        return ips[0].strip()

    # Check X-Real-IP header (used by some proxies)
    if request.headers.get('X-Real-IP'):
        return request.headers.get('X-Real-IP').strip()

    # Fallback to remote_addr
    return request.remote_addr or '0.0.0.0'

secret_manager_client = secretmanager_v1.SecretManagerServiceClient()
secret_name = secret_manager_client.secret_version_path(project=PROJECT_ID, secret=ALLOYDB_SECRET_NAME, secret_version="latest")
secret_request = secretmanager_v1.AccessSecretVersionRequest(name=secret_name)
secret_response = secret_manager_client.access_secret_version(request=secret_request)
PGPASSWORD = secret_response.payload.data.decode("UTF-8").strip()

engine = AlloyDBEngine.from_instance(
    project_id=PROJECT_ID,
    region=REGION,
    cluster=ALLOYDB_CLUSTER_NAME,
    instance=ALLOYDB_INSTANCE_NAME,
    database=ALLOYDB_DATABASE_NAME,
    user="postgres",
    password=PGPASSWORD
)

# Create a synchronous connection to our vectorstore
vectorstore = AlloyDBVectorStore.create_sync(
    engine=engine,
    table_name=ALLOYDB_TABLE_NAME,
    embedding_service=GoogleGenerativeAIEmbeddings(model=EMBEDDING_MODEL),
    id_column="id",
    content_column="description",
    embedding_column="product_embedding",
    metadata_columns=["id", "name", "categories"]
)

def create_app():
    app = Flask(__name__)

    # Rate limiting: Check before processing any request
    @app.before_request
    def check_rate_limit():
        """Apply rate limiting to all requests to prevent API abuse."""
        # Skip rate limiting if explicitly disabled
        if os.environ.get('DISABLE_RATE_LIMITING') == 'true':
            return None

        # Skip rate limiting for health checks and OPTIONS
        if request.path == '/_healthz' or request.method == 'OPTIONS':
            return None

        # Get client IP
        client_ip = get_client_ip()

        # Check rate limit
        allowed, remaining = rate_limiter.is_allowed(client_ip)

        if not allowed:
            # Log security event for rate limit exceeded
            logger.warning(
                f"Rate limit exceeded - IP: {client_ip}, "
                f"Path: {request.path}, Method: {request.method}, "
                f"security_event: rate_limit_exceeded"
            )

            # Return 429 Too Many Requests with headers
            response = jsonify({
                'error': 'Too Many Requests - Rate limit exceeded. Please try again later.',
                'retry_after': RATE_LIMIT_WINDOW
            })
            response.status_code = 429
            response.headers['X-RateLimit-Limit'] = str(RATE_LIMIT_REQUESTS)
            response.headers['X-RateLimit-Remaining'] = '0'
            response.headers['X-RateLimit-Reset'] = str(int(time.time()) + RATE_LIMIT_WINDOW)
            response.headers['Retry-After'] = str(RATE_LIMIT_WINDOW)
            return response

        # Add rate limit headers to successful requests (will be added in after_request)
        # Store in g object for access in after_request
        from flask import g
        g.rate_limit_remaining = remaining

        return None

    # Add security headers and CORS configuration to all responses
    @app.after_request
    def set_security_headers(response):
        # Prevent clickjacking attacks
        response.headers['X-Frame-Options'] = 'DENY'
        # Prevent MIME-type sniffing
        response.headers['X-Content-Type-Options'] = 'nosniff'
        # Enable HSTS for HTTPS
        response.headers['Strict-Transport-Security'] = 'max-age=31536000; includeSubDomains'
        # Content Security Policy
        response.headers['Content-Security-Policy'] = "default-src 'self'"
        # Referrer Policy
        response.headers['Referrer-Policy'] = 'strict-origin-when-cross-origin'
        # XSS Protection for older browsers
        response.headers['X-XSS-Protection'] = '1; mode=block'

        # Add CORS headers if ALLOWED_ORIGINS is configured
        allowed_origins_env = os.environ.get('ALLOWED_ORIGINS', '')
        origin = request.headers.get('Origin', '')

        if allowed_origins_env and origin:
            allowed_origins = [o.strip() for o in allowed_origins_env.split(',')]

            # Check if origin is allowed
            if '*' in allowed_origins or origin in allowed_origins:
                response.headers['Access-Control-Allow-Origin'] = origin
                response.headers['Access-Control-Allow-Credentials'] = 'true'
                response.headers['Access-Control-Allow-Methods'] = 'POST, OPTIONS'
                response.headers['Access-Control-Allow-Headers'] = 'Content-Type, Authorization'
                response.headers['Access-Control-Max-Age'] = '3600'
        elif allowed_origins_env == '*':
            # Allow all origins (not recommended for production)
            response.headers['Access-Control-Allow-Origin'] = '*'
            response.headers['Access-Control-Allow-Methods'] = 'POST, OPTIONS'
            response.headers['Access-Control-Allow-Headers'] = 'Content-Type, Authorization'
            response.headers['Access-Control-Max-Age'] = '3600'

        # Add rate limit headers for informational purposes
        if os.environ.get('DISABLE_RATE_LIMITING') != 'true':
            from flask import g
            if hasattr(g, 'rate_limit_remaining'):
                response.headers['X-RateLimit-Limit'] = str(RATE_LIMIT_REQUESTS)
                response.headers['X-RateLimit-Remaining'] = str(g.rate_limit_remaining)
                response.headers['X-RateLimit-Reset'] = str(int(time.time()) + RATE_LIMIT_WINDOW)

        # Apply gzip compression if enabled and client accepts it
        if os.environ.get('ENABLE_COMPRESSION') != 'false':
            accept_encoding = request.headers.get('Accept-Encoding', '')
            if 'gzip' in accept_encoding and response.status_code == 200:
                # Only compress JSON responses
                content_type = response.headers.get('Content-Type', '')
                if 'application/json' in content_type:
                    # Compress response data
                    gzip_buffer = BytesIO()
                    with gzip.GzipFile(mode='wb', fileobj=gzip_buffer, compresslevel=6) as gzip_file:
                        gzip_file.write(response.get_data())

                    # Set compressed data and headers
                    response.set_data(gzip_buffer.getvalue())
                    response.headers['Content-Encoding'] = 'gzip'
                    response.headers['Content-Length'] = len(response.get_data())

        return response

    # Handle preflight OPTIONS requests for CORS
    @app.route("/", methods=['OPTIONS'])
    def handle_options():
        return '', 200

    @app.route("/", methods=['POST'])
    def talkToGemini():
        logger.info("Beginning RAG call")

        # Validate that request has JSON content
        if not request.is_json:
            return jsonify({'error': 'Content-Type must be application/json'}), 400

        # Validate required fields are present
        if not request.json:
            return jsonify({'error': 'Request body must be valid JSON'}), 400

        if 'message' not in request.json:
            return jsonify({'error': 'Missing required field: message'}), 400

        if 'image' not in request.json:
            return jsonify({'error': 'Missing required field: image'}), 400

        prompt = request.json['message']

        # Validate message is not empty and within length limit
        if not prompt or not isinstance(prompt, str):
            return jsonify({'error': 'message must be a non-empty string'}), 400

        if len(prompt) > MAX_MESSAGE_LENGTH:
            return jsonify({'error': f'message too long (max {MAX_MESSAGE_LENGTH} characters)'}), 400

        prompt = unquote(prompt)

        # Validate image URL is not empty, within length limit, and is a valid URL
        image_url = request.json['image']
        if not image_url or not isinstance(image_url, str):
            return jsonify({'error': 'image must be a non-empty string (URL)'}), 400

        if len(image_url) > MAX_IMAGE_URL_LENGTH:
            return jsonify({'error': f'image URL too long (max {MAX_IMAGE_URL_LENGTH} characters)'}), 400

        # Validate URL format and scheme
        try:
            parsed_url = urlparse(image_url)
            if not all([parsed_url.scheme, parsed_url.netloc]):
                return jsonify({'error': 'Invalid image URL format'}), 400
            if parsed_url.scheme not in ['http', 'https']:
                return jsonify({'error': 'Image URL must use HTTP or HTTPS'}), 400
        except Exception as e:
            logger.error(f"URL validation failed: {str(e)}")
            return jsonify({'error': 'Invalid image URL'}), 400

        # Step 1 – Get a room description from Gemini-vision-pro
        try:
            llm_vision = ChatGoogleGenerativeAI(model=LLM_MODEL, timeout=30)
            message = HumanMessage(
                content=[
                    {
                        "type": "text",
                        "text": "You are a professional interior designer, give me a detailed description of the style of the room in this image",
                    },
                    {"type": "image_url", "image_url": image_url},
                ]
            )
            response = llm_vision.invoke([message])
            logger.info("Description step completed")
            logger.debug(f"LLM response: {response}")
            description_response = response.content
        except Exception as e:
            logger.error(f"LLM vision API failed: {str(e)}")
            return jsonify({'error': 'Failed to process image'}), 500

        # Step 2 – Similarity search with the description & user prompt
        vector_search_prompt = f""" This is the user's request: {prompt} Find the most relevant items for that prompt, while matching style of the room described here: {description_response} """
        logger.debug(f"Vector search prompt: {vector_search_prompt}")

        try:
            docs = vectorstore.similarity_search(vector_search_prompt)
            logger.info(f"Vector search completed for description")
            logger.info(f"Retrieved {len(docs)} documents")
        except Exception as e:
            logger.error(f"Vector search failed: {str(e)}")
            return jsonify({'error': 'Search temporarily unavailable'}), 503

        # Prepare relevant documents for inclusion in final prompt
        relevant_docs = ""
        for doc in docs:
            doc_details = doc.to_json()
            logger.debug(f"Adding relevant document to prompt context: {doc_details}")
            relevant_docs += str(doc_details) + ", "

        # Step 3 – Tie it all together by augmenting our call to Gemini-pro
        try:
            llm = ChatGoogleGenerativeAI(model=LLM_MODEL, timeout=30)
            design_prompt = (
                f" You are an interior designer that works for Online Boutique. You are tasked with providing recommendations to a customer on what they should add to a given room from our catalog. This is the description of the room: \n"
                f"{description_response} Here are a list of products that are relevant to it: {relevant_docs} Specifically, this is what the customer has asked for, see if you can accommodate it: {prompt} Start by repeating a brief description of the room's design to the customer, then provide your recommendations. Do your best to pick the most relevant item out of the list of products provided, but if none of them seem relevant, then say that instead of inventing a new product. At the end of the response, add a list of the IDs of the relevant products in the following format for the top 3 results: [<first product ID>], [<second product ID>], [<third product ID>] ")
            logger.info("Generating final design recommendations")
            logger.debug(f"Design prompt: {design_prompt}")
            design_response = llm.invoke(
                design_prompt
            )
        except Exception as e:
            logger.error(f"LLM generation API failed: {str(e)}")
            return jsonify({'error': 'Failed to generate recommendations'}), 500

        data = {'content': design_response.content}
        return data

    return app

def signal_handler(sig, frame):
    """Graceful shutdown handler"""
    logger.info(f"Received signal {sig}, initiating graceful shutdown...")
    try:
        # Close database connections
        engine.dispose()
        logger.info("Database connections closed")
    except Exception as e:
        logger.error(f"Error during shutdown: {str(e)}")
    logger.info("Shutdown complete")
    sys.exit(0)

if __name__ == "__main__":
    # Register signal handlers for graceful shutdown
    signal.signal(signal.SIGINT, signal_handler)
    signal.signal(signal.SIGTERM, signal_handler)

    # Create an instance of flask server when called directly
    app = create_app()
    port = int(os.environ.get('PORT', 8080))

    logger.info(f"Starting shopping assistant service on port {port}")
    logger.info(f"Using LLM model: {LLM_MODEL}")
    logger.info(f"Using embedding model: {EMBEDDING_MODEL}")

    # In production, use a WSGI server like gunicorn:
    # gunicorn --bind 0.0.0.0:8080 --workers 4 --timeout 60 --graceful-timeout 30 shoppingassistantservice:app
    app.run(host='0.0.0.0', port=port)
