"""
Visual Search Service

This service allows users to search for products using images.
It extracts features from uploaded images and finds similar products.
"""

from fastapi import FastAPI, File, UploadFile, HTTPException, Query
from fastapi.responses import JSONResponse
from fastapi.middleware.cors import CORSMiddleware
from typing import List, Optional
import io
import logging
from datetime import datetime

from .image_processor import ImageProcessor
from .vector_store import VectorStore
from .models import (
    VisualSearchResponse,
    SimilarProduct,
    HealthCheckResponse
)

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

app = FastAPI(
    title="Visual Search Service",
    description="Image-based product search using deep learning",
    version="1.0.0"
)

# CORS
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Initialize components
image_processor = ImageProcessor()
vector_store = VectorStore()


@app.on_event("startup")
async def startup_event():
    """Initialize service on startup"""
    logger.info("Visual Search Service starting up...")

    # Load pre-indexed products
    try:
        vector_store.load_index()
        logger.info(f"Loaded {vector_store.get_index_size()} products from index")
    except Exception as e:
        logger.warning(f"Could not load existing index: {e}")
        logger.info("Starting with empty index")


@app.get("/", response_model=HealthCheckResponse)
async def root():
    """Root endpoint - health check"""
    return HealthCheckResponse(
        status="healthy",
        service="visual-search-service",
        timestamp=datetime.utcnow()
    )


@app.get("/health", response_model=HealthCheckResponse)
async def health_check():
    """Health check endpoint"""
    return HealthCheckResponse(
        status="healthy",
        service="visual-search-service",
        timestamp=datetime.utcnow(),
        index_size=vector_store.get_index_size()
    )


@app.post("/search", response_model=VisualSearchResponse)
async def visual_search(
    image: UploadFile = File(...),
    top_k: int = Query(5, ge=1, le=20, description="Number of results to return"),
    threshold: float = Query(0.7, ge=0.0, le=1.0, description="Similarity threshold")
):
    """
    Search for similar products using an uploaded image

    Args:
        image: Uploaded image file (JPEG, PNG)
        top_k: Number of similar products to return (1-20)
        threshold: Minimum similarity score (0.0-1.0)

    Returns:
        List of similar products with similarity scores
    """
    try:
        # Validate file type
        if image.content_type not in ["image/jpeg", "image/png", "image/jpg"]:
            raise HTTPException(
                status_code=400,
                detail=f"Invalid file type: {image.content_type}. Only JPEG and PNG are supported."
            )

        # Read image
        contents = await image.read()
        image_bytes = io.BytesIO(contents)

        logger.info(f"Processing search request for image: {image.filename}")

        # Extract features
        features = image_processor.extract_features(image_bytes)

        # Search for similar products
        similar_products = vector_store.search(features, top_k=top_k, threshold=threshold)

        logger.info(f"Found {len(similar_products)} similar products")

        return VisualSearchResponse(
            query_image=image.filename,
            results=similar_products,
            total_results=len(similar_products),
            timestamp=datetime.utcnow()
        )

    except Exception as e:
        logger.error(f"Error processing image search: {str(e)}", exc_info=True)
        raise HTTPException(status_code=500, detail=f"Error processing image: {str(e)}")


@app.post("/index/product")
async def index_product(
    product_id: str,
    image: UploadFile = File(...)
):
    """
    Index a product image for future searches

    Args:
        product_id: Unique product identifier
        image: Product image to index

    Returns:
        Success message
    """
    try:
        # Read image
        contents = await image.read()
        image_bytes = io.BytesIO(contents)

        logger.info(f"Indexing product: {product_id}")

        # Extract features
        features = image_processor.extract_features(image_bytes)

        # Add to vector store
        vector_store.add_product(product_id, features)

        return JSONResponse(
            content={
                "status": "success",
                "product_id": product_id,
                "message": "Product indexed successfully"
            }
        )

    except Exception as e:
        logger.error(f"Error indexing product: {str(e)}", exc_info=True)
        raise HTTPException(status_code=500, detail=f"Error indexing product: {str(e)}")


@app.post("/index/batch")
async def batch_index_products():
    """
    Batch index all products from product catalog service

    This endpoint fetches all products from the product catalog
    and indexes their images for visual search.
    """
    try:
        logger.info("Starting batch indexing of products...")

        # This would call the product catalog service to get all products
        # and index their images
        # For now, we'll create some sample data

        indexed_count = vector_store.create_sample_index()

        logger.info(f"Batch indexing complete. Indexed {indexed_count} products.")

        return JSONResponse(
            content={
                "status": "success",
                "indexed_count": indexed_count,
                "message": "Batch indexing completed"
            }
        )

    except Exception as e:
        logger.error(f"Error in batch indexing: {str(e)}", exc_info=True)
        raise HTTPException(status_code=500, detail=f"Error in batch indexing: {str(e)}")


@app.get("/stats")
async def get_stats():
    """Get service statistics"""
    return {
        "index_size": vector_store.get_index_size(),
        "model_info": image_processor.get_model_info(),
        "uptime": "running",
        "version": "1.0.0"
    }


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(
        "app.main:app",
        host="0.0.0.0",
        port=8090,
        reload=True
    )
