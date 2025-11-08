"""
Tests for Visual Search Service data models
"""

import pytest
from datetime import datetime
from app.models import (
    SimilarProduct,
    VisualSearchResponse,
    HealthCheckResponse,
    IndexProductRequest
)


class TestSimilarProduct:
    """Test SimilarProduct model"""

    def test_create_similar_product(self):
        """Test creating a valid SimilarProduct"""
        product = SimilarProduct(
            product_id="TEST001",
            similarity_score=0.95,
            product_name="Test Product",
            product_image_url="https://example.com/image.jpg",
            price=99.99
        )

        assert product.product_id == "TEST001"
        assert product.similarity_score == 0.95
        assert product.product_name == "Test Product"
        assert product.price == 99.99

    def test_similarity_score_validation(self):
        """Test that similarity score must be between 0 and 1"""
        # Valid scores
        product = SimilarProduct(
            product_id="TEST001",
            similarity_score=0.5
        )
        assert product.similarity_score == 0.5

        # Invalid score > 1
        with pytest.raises(ValueError):
            SimilarProduct(
                product_id="TEST001",
                similarity_score=1.5
            )

        # Invalid score < 0
        with pytest.raises(ValueError):
            SimilarProduct(
                product_id="TEST001",
                similarity_score=-0.1
            )

    def test_optional_fields(self):
        """Test that optional fields can be None"""
        product = SimilarProduct(
            product_id="TEST001",
            similarity_score=0.8
        )

        assert product.product_name is None
        assert product.product_image_url is None
        assert product.price is None


class TestVisualSearchResponse:
    """Test VisualSearchResponse model"""

    def test_create_visual_search_response(self):
        """Test creating a valid VisualSearchResponse"""
        products = [
            SimilarProduct(
                product_id="PROD1",
                similarity_score=0.95
            ),
            SimilarProduct(
                product_id="PROD2",
                similarity_score=0.85
            )
        ]

        response = VisualSearchResponse(
            query_image="test_image.jpg",
            results=products,
            total_results=2,
            timestamp=datetime.now()
        )

        assert response.query_image == "test_image.jpg"
        assert len(response.results) == 2
        assert response.total_results == 2
        assert isinstance(response.timestamp, datetime)

    def test_empty_results(self):
        """Test response with no results"""
        response = VisualSearchResponse(
            query_image="test_image.jpg",
            results=[],
            total_results=0,
            timestamp=datetime.now()
        )

        assert len(response.results) == 0
        assert response.total_results == 0


class TestHealthCheckResponse:
    """Test HealthCheckResponse model"""

    def test_create_health_check_response(self):
        """Test creating a valid HealthCheckResponse"""
        response = HealthCheckResponse(
            status="healthy",
            service="visualsearch",
            timestamp=datetime.now(),
            index_size=100
        )

        assert response.status == "healthy"
        assert response.service == "visualsearch"
        assert isinstance(response.timestamp, datetime)
        assert response.index_size == 100

    def test_optional_index_size(self):
        """Test that index_size is optional"""
        response = HealthCheckResponse(
            status="healthy",
            service="visualsearch",
            timestamp=datetime.now()
        )

        assert response.index_size is None


class TestIndexProductRequest:
    """Test IndexProductRequest model"""

    def test_create_index_product_request(self):
        """Test creating a valid IndexProductRequest"""
        request = IndexProductRequest(
            product_id="PROD001",
            product_name="Test Product",
            image_url="https://example.com/product.jpg"
        )

        assert request.product_id == "PROD001"
        assert request.product_name == "Test Product"
        assert request.image_url == "https://example.com/product.jpg"

    def test_minimal_index_request(self):
        """Test creating request with only required field"""
        request = IndexProductRequest(
            product_id="PROD001"
        )

        assert request.product_id == "PROD001"
        assert request.product_name is None
        assert request.image_url is None

    def test_missing_product_id(self):
        """Test that product_id is required"""
        with pytest.raises(ValueError):
            IndexProductRequest()
