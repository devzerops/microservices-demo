"""
Vector similarity search using FAISS
"""

import numpy as np
import faiss
import pickle
import os
from typing import List, Tuple, Optional
import logging

from .models import SimilarProduct

logger = logging.getLogger(__name__)


class VectorStore:
    """Manages vector storage and similarity search using FAISS"""

    def __init__(self, dimension: int = 1280):
        """
        Initialize vector store

        Args:
            dimension: Dimension of feature vectors (MobileNetV2 outputs 1280)
        """
        self.dimension = dimension

        # Create FAISS index (L2 distance)
        self.index = faiss.IndexFlatL2(dimension)

        # Store product IDs and metadata
        self.product_ids = []
        self.product_metadata = {}

        self.index_file = "data/faiss_index.bin"
        self.metadata_file = "data/metadata.pkl"

        logger.info(f"Initialized vector store with dimension {dimension}")

    def add_product(self, product_id: str, features: np.ndarray, metadata: Optional[dict] = None):
        """
        Add a product to the index

        Args:
            product_id: Unique product identifier
            features: Feature vector
            metadata: Optional product metadata (name, price, etc.)
        """
        # Ensure features are the right shape
        if features.shape[0] != self.dimension:
            raise ValueError(f"Feature dimension {features.shape[0]} doesn't match index dimension {self.dimension}")

        # Add to FAISS index
        features_2d = features.reshape(1, -1).astype('float32')
        self.index.add(features_2d)

        # Store product ID
        self.product_ids.append(product_id)

        # Store metadata
        if metadata:
            self.product_metadata[product_id] = metadata

        logger.debug(f"Added product {product_id} to index")

    def search(self, query_features: np.ndarray, top_k: int = 5, threshold: float = 0.7) -> List[SimilarProduct]:
        """
        Search for similar products

        Args:
            query_features: Query feature vector
            top_k: Number of results to return
            threshold: Minimum similarity threshold (0-1)

        Returns:
            List of similar products with scores
        """
        if self.index.ntotal == 0:
            logger.warning("Index is empty")
            return []

        # Ensure features are the right shape
        query_features_2d = query_features.reshape(1, -1).astype('float32')

        # Search
        distances, indices = self.index.search(query_features_2d, min(top_k, self.index.ntotal))

        # Convert L2 distances to similarity scores (0-1)
        # Smaller distance = higher similarity
        # Using exponential decay: similarity = exp(-distance)
        similarities = np.exp(-distances[0])

        results = []
        for idx, similarity in zip(indices[0], similarities):
            if similarity < threshold:
                continue

            product_id = self.product_ids[idx]
            metadata = self.product_metadata.get(product_id, {})

            results.append(SimilarProduct(
                product_id=product_id,
                similarity_score=float(similarity),
                product_name=metadata.get('name'),
                product_image_url=metadata.get('image_url'),
                price=metadata.get('price')
            ))

        return results

    def save_index(self):
        """Save index and metadata to disk"""
        try:
            os.makedirs("data", exist_ok=True)

            # Save FAISS index
            faiss.write_index(self.index, self.index_file)

            # Save metadata
            with open(self.metadata_file, 'wb') as f:
                pickle.dump({
                    'product_ids': self.product_ids,
                    'product_metadata': self.product_metadata
                }, f)

            logger.info(f"Saved index with {self.index.ntotal} products")

        except Exception as e:
            logger.error(f"Error saving index: {str(e)}")
            raise

    def load_index(self):
        """Load index and metadata from disk"""
        try:
            if not os.path.exists(self.index_file):
                logger.warning(f"Index file not found: {self.index_file}")
                return

            # Load FAISS index
            self.index = faiss.read_index(self.index_file)

            # Load metadata
            with open(self.metadata_file, 'rb') as f:
                data = pickle.load(f)
                self.product_ids = data['product_ids']
                self.product_metadata = data['product_metadata']

            logger.info(f"Loaded index with {self.index.ntotal} products")

        except Exception as e:
            logger.error(f"Error loading index: {str(e)}")
            raise

    def get_index_size(self) -> int:
        """Get number of products in index"""
        return self.index.ntotal

    def create_sample_index(self) -> int:
        """
        Create sample index for demonstration

        This creates fake products with random features for demo purposes.
        In production, this would fetch real products from the catalog.
        """
        logger.info("Creating sample product index...")

        # Sample product data (matching the existing product catalog)
        sample_products = [
            {
                'id': 'OLJCESPC7Z',
                'name': 'Sunglasses',
                'price': 19.99,
                'image_url': '/static/img/products/sunglasses.jpg'
            },
            {
                'id': '66VCHSJNUP',
                'name': 'Tank Top',
                'price': 18.99,
                'image_url': '/static/img/products/tank-top.jpg'
            },
            {
                'id': '1YMWWN1N4O',
                'name': 'Watch',
                'price': 109.99,
                'image_url': '/static/img/products/watch.jpg'
            },
            {
                'id': 'L9ECAV7KIM',
                'name': 'Loafers',
                'price': 89.99,
                'image_url': '/static/img/products/loafers.jpg'
            },
            {
                'id': '2ZYFJ3GM2N',
                'name': 'Hairdryer',
                'price': 24.99,
                'image_url': '/static/img/products/hairdryer.jpg'
            },
            {
                'id': '0PUK6V6EV0',
                'name': 'Candle Holder',
                'price': 18.99,
                'image_url': '/static/img/products/candle-holder.jpg'
            },
            {
                'id': 'LS4PSXUNUM',
                'name': 'Salt & Pepper Shakers',
                'price': 18.49,
                'image_url': '/static/img/products/salt-pepper-shakers.jpg'
            },
            {
                'id': '9SIQT8TOJO',
                'name': 'Bamboo Glass Jar',
                'price': 5.49,
                'image_url': '/static/img/products/bamboo-glass-jar.jpg'
            },
            {
                'id': '6E92ZMYYFZ',
                'name': 'Mug',
                'price': 8.99,
                'image_url': '/static/img/products/mug.jpg'
            }
        ]

        # Generate random features for each product (in production, extract from real images)
        for product in sample_products:
            # Create semi-realistic random features
            features = np.random.randn(self.dimension).astype('float32')
            features = features / np.linalg.norm(features)  # Normalize

            metadata = {
                'name': product['name'],
                'price': product['price'],
                'image_url': product['image_url']
            }

            self.add_product(product['id'], features, metadata)

        # Save the index
        self.save_index()

        logger.info(f"Created sample index with {len(sample_products)} products")

        return len(sample_products)
