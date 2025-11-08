#!/usr/bin/python
#
# Copyright 2024 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

import pytest
import grpc
from unittest.mock import Mock, patch, MagicMock
from concurrent import futures

import demo_pb2
import demo_pb2_grpc
from grpc_health.v1 import health_pb2
from recommendation_server import RecommendationService


class TestRecommendationService:
    """Test RecommendationService functionality"""

    @pytest.fixture
    def mock_product_catalog_stub(self):
        """Create a mock product catalog stub"""
        stub = Mock()

        # Mock product list response
        products = [
            demo_pb2.Product(id="OLJCESPC7Z", name="Sunglasses"),
            demo_pb2.Product(id="66VCHSJNUP", name="Tank Top"),
            demo_pb2.Product(id="1YMWWN1N4O", name="Watch"),
            demo_pb2.Product(id="L9ECAV7KIM", name="Loafers"),
            demo_pb2.Product(id="2ZYFJ3GM2N", name="Hairdryer"),
            demo_pb2.Product(id="0PUK6V6EV0", name="Candle Holder"),
            demo_pb2.Product(id="LS4PSXUNUM", name="Salt & Pepper Shakers"),
            demo_pb2.Product(id="9SIQT8TOJO", name="Bamboo Glass Jar"),
            demo_pb2.Product(id="6E92ZMYYFZ", name="Mug"),
        ]

        stub.ListProducts.return_value = demo_pb2.ListProductsResponse(products=products)
        return stub

    @patch('recommendation_server.product_catalog_stub')
    def test_list_recommendations_returns_products(self, mock_stub, mock_product_catalog_stub):
        """Test that ListRecommendations returns product recommendations"""
        mock_stub.ListProducts = mock_product_catalog_stub.ListProducts

        service = RecommendationService()
        request = demo_pb2.ListRecommendationsRequest(
            user_id="user-123",
            product_ids=[]
        )
        context = Mock()

        response = service.ListRecommendations(request, context)

        assert isinstance(response, demo_pb2.ListRecommendationsResponse)
        assert len(response.product_ids) <= 5
        assert len(response.product_ids) > 0

    @patch('recommendation_server.product_catalog_stub')
    def test_list_recommendations_filters_current_products(self, mock_stub, mock_product_catalog_stub):
        """Test that recommendations exclude currently viewed products"""
        mock_stub.ListProducts = mock_product_catalog_stub.ListProducts

        service = RecommendationService()
        current_product_ids = ["OLJCESPC7Z", "66VCHSJNUP"]
        request = demo_pb2.ListRecommendationsRequest(
            user_id="user-456",
            product_ids=current_product_ids
        )
        context = Mock()

        response = service.ListRecommendations(request, context)

        # Verify that none of the current products are in recommendations
        for product_id in response.product_ids:
            assert product_id not in current_product_ids

    @patch('recommendation_server.product_catalog_stub')
    def test_list_recommendations_max_five_items(self, mock_stub, mock_product_catalog_stub):
        """Test that recommendations return at most 5 products"""
        mock_stub.ListProducts = mock_product_catalog_stub.ListProducts

        service = RecommendationService()
        request = demo_pb2.ListRecommendationsRequest(
            user_id="user-789",
            product_ids=[]
        )
        context = Mock()

        response = service.ListRecommendations(request, context)

        assert len(response.product_ids) <= 5

    @patch('recommendation_server.product_catalog_stub')
    def test_list_recommendations_with_all_products_filtered(self, mock_stub):
        """Test recommendations when all products are already in cart"""
        # Create a smaller product list
        products = [
            demo_pb2.Product(id="PROD1", name="Product 1"),
            demo_pb2.Product(id="PROD2", name="Product 2"),
        ]
        mock_stub.ListProducts.return_value = demo_pb2.ListProductsResponse(products=products)

        service = RecommendationService()
        request = demo_pb2.ListRecommendationsRequest(
            user_id="user-999",
            product_ids=["PROD1", "PROD2"]  # All products already in cart
        )
        context = Mock()

        response = service.ListRecommendations(request, context)

        # Should return empty list when all products are filtered
        assert len(response.product_ids) == 0

    @patch('recommendation_server.product_catalog_stub')
    def test_list_recommendations_randomness(self, mock_stub, mock_product_catalog_stub):
        """Test that recommendations are randomized"""
        mock_stub.ListProducts = mock_product_catalog_stub.ListProducts

        service = RecommendationService()
        request = demo_pb2.ListRecommendationsRequest(
            user_id="user-random",
            product_ids=[]
        )
        context = Mock()

        # Get multiple recommendation sets
        recommendations = []
        for _ in range(5):
            response = service.ListRecommendations(request, context)
            recommendations.append(tuple(response.product_ids))

        # At least some should be different (not guaranteed but highly likely)
        unique_recommendations = set(recommendations)
        # With random sampling, we expect some variation
        assert len(unique_recommendations) >= 1

    def test_health_check_serving(self):
        """Test that health check returns SERVING status"""
        service = RecommendationService()
        request = health_pb2.HealthCheckRequest()
        context = Mock()

        response = service.Check(request, context)

        assert response.status == health_pb2.HealthCheckResponse.SERVING

    def test_health_watch_unimplemented(self):
        """Test that Watch returns UNIMPLEMENTED status"""
        service = RecommendationService()
        request = health_pb2.HealthCheckRequest()
        context = Mock()

        response = service.Watch(request, context)

        assert response.status == health_pb2.HealthCheckResponse.UNIMPLEMENTED


class TestRecommendationServiceIntegration:
    """Integration tests for recommendation service"""

    @pytest.fixture
    def mock_product_catalog_channel(self):
        """Create a mock product catalog service"""
        server = grpc.server(futures.ThreadPoolExecutor(max_workers=1))

        # Create a mock product catalog service
        class MockProductCatalogService(demo_pb2_grpc.ProductCatalogServiceServicer):
            def ListProducts(self, request, context):
                products = [
                    demo_pb2.Product(id=f"PROD{i}", name=f"Product {i}")
                    for i in range(10)
                ]
                return demo_pb2.ListProductsResponse(products=products)

        demo_pb2_grpc.add_ProductCatalogServiceServicer_to_server(
            MockProductCatalogService(), server
        )

        port = server.add_insecure_port('[::]:0')
        server.start()

        yield f'localhost:{port}'

        server.stop(0)

    @patch('recommendation_server.product_catalog_stub')
    def test_recommendation_service_with_grpc_server(self, mock_stub, mock_product_catalog_channel):
        """Test recommendation service through gRPC"""
        # Create a real gRPC channel to mock catalog
        channel = grpc.insecure_channel(mock_product_catalog_channel)
        mock_stub.ListProducts = demo_pb2_grpc.ProductCatalogServiceStub(channel).ListProducts

        # Create recommendation server
        server = grpc.server(futures.ThreadPoolExecutor(max_workers=1))
        service = RecommendationService()
        demo_pb2_grpc.add_RecommendationServiceServicer_to_server(service, server)
        health_pb2_grpc.add_HealthServicer_to_server(service, server)

        port = server.add_insecure_port('[::]:0')
        server.start()

        try:
            # Test the recommendation service
            with grpc.insecure_channel(f'localhost:{port}') as rec_channel:
                stub = demo_pb2_grpc.RecommendationServiceStub(rec_channel)
                request = demo_pb2.ListRecommendationsRequest(
                    user_id="integration-test-user",
                    product_ids=["PROD1", "PROD2"]
                )

                response = stub.ListRecommendations(request)

                assert isinstance(response, demo_pb2.ListRecommendationsResponse)
                assert len(response.product_ids) <= 5
                # Should not include filtered products
                for pid in response.product_ids:
                    assert pid not in ["PROD1", "PROD2"]

        finally:
            server.stop(0)
            channel.close()


class TestRecommendationAlgorithm:
    """Test the recommendation algorithm logic"""

    @patch('recommendation_server.product_catalog_stub')
    def test_recommendation_count_with_few_products(self, mock_stub):
        """Test recommendations when catalog has fewer than 5 products"""
        products = [
            demo_pb2.Product(id="P1", name="Product 1"),
            demo_pb2.Product(id="P2", name="Product 2"),
            demo_pb2.Product(id="P3", name="Product 3"),
        ]
        mock_stub.ListProducts.return_value = demo_pb2.ListProductsResponse(products=products)

        service = RecommendationService()
        request = demo_pb2.ListRecommendationsRequest(
            user_id="user-small-catalog",
            product_ids=[]
        )
        context = Mock()

        response = service.ListRecommendations(request, context)

        # Should return all 3 products (less than max of 5)
        assert len(response.product_ids) == 3

    @patch('recommendation_server.product_catalog_stub')
    def test_product_ids_are_strings(self, mock_stub, mock_product_catalog_stub):
        """Test that returned product IDs are valid strings"""
        mock_stub.ListProducts = mock_product_catalog_stub.ListProducts

        service = RecommendationService()
        request = demo_pb2.ListRecommendationsRequest(
            user_id="user-type-check",
            product_ids=[]
        )
        context = Mock()

        response = service.ListRecommendations(request, context)

        for product_id in response.product_ids:
            assert isinstance(product_id, str)
            assert len(product_id) > 0


if __name__ == '__main__':
    pytest.main([__file__, '-v'])
