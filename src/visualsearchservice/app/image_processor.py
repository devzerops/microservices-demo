"""
Image processing and feature extraction using deep learning
"""

import numpy as np
from PIL import Image
import tensorflow as tf
from tensorflow.keras.applications import MobileNetV2
from tensorflow.keras.applications.mobilenet_v2 import preprocess_input
from tensorflow.keras.preprocessing import image as keras_image
import logging
from typing import BinaryIO

logger = logging.getLogger(__name__)


class ImageProcessor:
    """Handles image processing and feature extraction"""

    def __init__(self, model_name: str = "MobileNetV2"):
        """
        Initialize image processor with a pre-trained model

        Args:
            model_name: Name of the model to use (default: MobileNetV2)
        """
        logger.info(f"Loading {model_name} model...")

        # Load pre-trained model
        base_model = MobileNetV2(
            weights='imagenet',
            include_top=False,
            pooling='avg',
            input_shape=(224, 224, 3)
        )

        # Use the model for feature extraction (no training)
        base_model.trainable = False

        self.model = base_model
        self.model_name = model_name
        self.input_shape = (224, 224)

        logger.info(f"{model_name} model loaded successfully")

    def extract_features(self, image_file: BinaryIO) -> np.ndarray:
        """
        Extract feature vector from an image

        Args:
            image_file: Binary image file (BytesIO or file object)

        Returns:
            Feature vector as numpy array
        """
        try:
            # Load and preprocess image
            img = Image.open(image_file)

            # Convert to RGB if needed
            if img.mode != 'RGB':
                img = img.convert('RGB')

            # Resize to model input size
            img = img.resize(self.input_shape)

            # Convert to array
            img_array = keras_image.img_to_array(img)
            img_array = np.expand_dims(img_array, axis=0)

            # Preprocess for MobileNetV2
            img_array = preprocess_input(img_array)

            # Extract features
            features = self.model.predict(img_array, verbose=0)

            # Normalize features
            features = features.flatten()
            features = features / np.linalg.norm(features)

            logger.debug(f"Extracted feature vector of shape: {features.shape}")

            return features

        except Exception as e:
            logger.error(f"Error extracting features: {str(e)}")
            raise

    def get_model_info(self) -> dict:
        """Get information about the loaded model"""
        return {
            "model_name": self.model_name,
            "input_shape": self.input_shape,
            "feature_dimension": self.model.output_shape[-1]
        }

    def preprocess_image_url(self, image_url: str) -> np.ndarray:
        """
        Download and process an image from URL

        Args:
            image_url: URL of the image

        Returns:
            Feature vector
        """
        import requests
        from io import BytesIO

        try:
            response = requests.get(image_url, timeout=10)
            response.raise_for_status()

            image_bytes = BytesIO(response.content)
            return self.extract_features(image_bytes)

        except Exception as e:
            logger.error(f"Error processing image from URL: {str(e)}")
            raise
