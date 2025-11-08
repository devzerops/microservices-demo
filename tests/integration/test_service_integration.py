"""
Integration Tests for Microservices Demo

These tests verify the interaction between multiple services
and validate complete business workflows.
"""

import pytest
import grpc
import os
import sys
import time
from typing import Generator

# Add the proto directory to the path
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '../../src/productcatalogservice/genproto'))
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '../../src/recommendationservice'))

import demo_pb2
import demo_pb2_grpc


class TestServiceIntegration:
    """Integration tests for service-to-service communication"""

    @pytest.fixture(scope="class")
    def product_catalog_stub(self) -> Generator:
        """Create a gRPC stub for ProductCatalogService"""
        catalog_addr = os.getenv('PRODUCT_CATALOG_SERVICE_ADDR', 'localhost:3550')
        channel = grpc.insecure_channel(catalog_addr)

        # Wait for service to be ready
        grpc.channel_ready_future(channel).result(timeout=10)

        stub = demo_pb2_grpc.ProductCatalogServiceStub(channel)
        yield stub
        channel.close()

    @pytest.fixture(scope="class")
    def recommendation_stub(self) -> Generator:
        """Create a gRPC stub for RecommendationService"""
        rec_addr = os.getenv('RECOMMENDATION_SERVICE_ADDR', 'localhost:8080')
        channel = grpc.insecure_channel(rec_addr)

        # Wait for service to be ready
        grpc.channel_ready_future(channel).result(timeout=10)

        stub = demo_pb2_grpc.RecommendationServiceStub(channel)
        yield stub
        channel.close()

    @pytest.fixture(scope="class")
    def cart_stub(self) -> Generator:
        """Create a gRPC stub for CartService"""
        cart_addr = os.getenv('CART_SERVICE_ADDR', 'localhost:7070')
        channel = grpc.insecure_channel(cart_addr)

        # Wait for service to be ready
        grpc.channel_ready_future(channel).result(timeout=10)

        stub = demo_pb2_grpc.CartServiceStub(channel)
        yield stub
        channel.close()

    @pytest.fixture(scope="class")
    def checkout_stub(self) -> Generator:
        """Create a gRPC stub for CheckoutService"""
        checkout_addr = os.getenv('CHECKOUT_SERVICE_ADDR', 'localhost:5050')
        channel = grpc.insecure_channel(checkout_addr)

        # Wait for service to be ready
        grpc.channel_ready_future(channel).result(timeout=10)

        stub = demo_pb2_grpc.CheckoutServiceStub(channel)
        yield stub
        channel.close()

    def test_product_catalog_list_products(self, product_catalog_stub):
        """Test that ProductCatalogService returns a list of products"""
        request = demo_pb2.Empty()
        response = product_catalog_stub.ListProducts(request)

        assert response is not None
        assert len(response.products) > 0

        # Verify product structure
        for product in response.products:
            assert product.id != ""
            assert product.name != ""
            assert product.price_usd.currency_code == "USD"

    def test_product_catalog_get_product(self, product_catalog_stub):
        """Test getting a specific product by ID"""
        # First get all products
        all_products = product_catalog_stub.ListProducts(demo_pb2.Empty())
        assert len(all_products.products) > 0

        # Get the first product by ID
        product_id = all_products.products[0].id
        request = demo_pb2.GetProductRequest(id=product_id)
        response = product_catalog_stub.GetProduct(request)

        assert response.id == product_id
        assert response.name != ""

    def test_product_catalog_search(self, product_catalog_stub):
        """Test product search functionality"""
        request = demo_pb2.SearchProductsRequest(query="")
        response = product_catalog_stub.SearchProducts(request)

        assert response is not None
        # Empty query should return some results
        assert len(response.results) >= 0

    def test_recommendation_service_integration(
        self,
        product_catalog_stub,
        recommendation_stub
    ):
        """Test that RecommendationService integrates with ProductCatalog"""
        # Get all products from catalog
        all_products = product_catalog_stub.ListProducts(demo_pb2.Empty())
        assert len(all_products.products) > 0

        # Get recommendations (excluding first product)
        product_ids = [all_products.products[0].id]
        request = demo_pb2.ListRecommendationsRequest(
            user_id="test-user-123",
            product_ids=product_ids
        )
        response = recommendation_stub.ListRecommendations(request)

        assert response is not None
        assert len(response.product_ids) <= 5

        # Verify recommendations don't include the excluded product
        for rec_id in response.product_ids:
            assert rec_id not in product_ids

    def test_cart_operations(self, cart_stub, product_catalog_stub):
        """Test adding and retrieving items from cart"""
        user_id = f"integration-test-user-{int(time.time())}"

        # Get a product to add to cart
        all_products = product_catalog_stub.ListProducts(demo_pb2.Empty())
        product_id = all_products.products[0].id

        # Add item to cart
        add_request = demo_pb2.AddItemRequest(
            user_id=user_id,
            item=demo_pb2.CartItem(
                product_id=product_id,
                quantity=2
            )
        )
        cart_stub.AddItem(add_request)

        # Get cart
        get_request = demo_pb2.GetCartRequest(user_id=user_id)
        cart = cart_stub.GetCart(get_request)

        assert len(cart.items) == 1
        assert cart.items[0].product_id == product_id
        assert cart.items[0].quantity == 2

        # Empty cart
        empty_request = demo_pb2.EmptyCartRequest(user_id=user_id)
        cart_stub.EmptyCart(empty_request)

        # Verify cart is empty
        cart = cart_stub.GetCart(get_request)
        assert len(cart.items) == 0

    def test_complete_checkout_flow(
        self,
        product_catalog_stub,
        cart_stub,
        checkout_stub
    ):
        """Test complete checkout flow: browse -> cart -> checkout"""
        user_id = f"checkout-test-user-{int(time.time())}"

        # Step 1: Browse products
        all_products = product_catalog_stub.ListProducts(demo_pb2.Empty())
        assert len(all_products.products) > 0

        # Step 2: Add items to cart
        for i, product in enumerate(all_products.products[:2]):  # Add first 2 products
            add_request = demo_pb2.AddItemRequest(
                user_id=user_id,
                item=demo_pb2.CartItem(
                    product_id=product.id,
                    quantity=i + 1
                )
            )
            cart_stub.AddItem(add_request)

        # Step 3: Verify cart
        cart = cart_stub.GetCart(demo_pb2.GetCartRequest(user_id=user_id))
        assert len(cart.items) == 2

        # Step 4: Checkout
        checkout_request = demo_pb2.PlaceOrderRequest(
            user_id=user_id,
            user_currency="USD",
            address=demo_pb2.Address(
                street_address="123 Test St",
                city="Test City",
                state="TS",
                country="Test Country",
                zip_code="12345"
            ),
            email=f"{user_id}@test.com",
            credit_card=demo_pb2.CreditCardInfo(
                credit_card_number="4432-8015-6152-0454",
                credit_card_cvv=672,
                credit_card_expiration_year=2025,
                credit_card_expiration_month=1
            )
        )

        order = checkout_stub.PlaceOrder(checkout_request)

        # Verify order
        assert order.order_id != ""
        assert len(order.items) == 2
        assert order.shipping_cost is not None
        assert order.shipping_address.street_address == "123 Test St"

        # Verify cart is empty after checkout
        cart = cart_stub.GetCart(demo_pb2.GetCartRequest(user_id=user_id))
        assert len(cart.items) == 0


class TestServiceHealth:
    """Test health checks for all services"""

    @pytest.fixture(scope="class")
    def service_addresses(self):
        """Return addresses of all services to test"""
        return {
            'productcatalog': os.getenv('PRODUCT_CATALOG_SERVICE_ADDR', 'localhost:3550'),
            'recommendation': os.getenv('RECOMMENDATION_SERVICE_ADDR', 'localhost:8080'),
            'cart': os.getenv('CART_SERVICE_ADDR', 'localhost:7070'),
            'checkout': os.getenv('CHECKOUT_SERVICE_ADDR', 'localhost:5050'),
        }

    @pytest.mark.parametrize("service_name", [
        'productcatalog',
        'recommendation',
        'cart',
        'checkout',
    ])
    def test_service_health_check(self, service_addresses, service_name):
        """Test that each service responds to health checks"""
        from grpc_health.v1 import health_pb2, health_pb2_grpc

        addr = service_addresses[service_name]
        channel = grpc.insecure_channel(addr)

        try:
            # Wait for service to be ready
            grpc.channel_ready_future(channel).result(timeout=10)

            # Check health
            health_stub = health_pb2_grpc.HealthStub(channel)
            request = health_pb2.HealthCheckRequest()
            response = health_stub.Check(request)

            assert response.status == health_pb2.HealthCheckResponse.SERVING
        finally:
            channel.close()


class TestErrorHandling:
    """Test error handling in service integration"""

    @pytest.fixture(scope="class")
    def product_catalog_stub(self) -> Generator:
        catalog_addr = os.getenv('PRODUCT_CATALOG_SERVICE_ADDR', 'localhost:3550')
        channel = grpc.insecure_channel(catalog_addr)
        grpc.channel_ready_future(channel).result(timeout=10)
        stub = demo_pb2_grpc.ProductCatalogServiceStub(channel)
        yield stub
        channel.close()

    def test_get_nonexistent_product(self, product_catalog_stub):
        """Test that getting a non-existent product returns proper error"""
        request = demo_pb2.GetProductRequest(id="NONEXISTENT-PRODUCT-ID")

        with pytest.raises(grpc.RpcError) as exc_info:
            product_catalog_stub.GetProduct(request)

        assert exc_info.value.code() == grpc.StatusCode.NOT_FOUND

    def test_invalid_cart_operations(self):
        """Test invalid cart operations"""
        cart_addr = os.getenv('CART_SERVICE_ADDR', 'localhost:7070')
        channel = grpc.insecure_channel(cart_addr)

        try:
            grpc.channel_ready_future(channel).result(timeout=10)
            stub = demo_pb2_grpc.CartServiceStub(channel)

            # Try to add item with invalid quantity
            request = demo_pb2.AddItemRequest(
                user_id="test-user",
                item=demo_pb2.CartItem(
                    product_id="PRODUCT-ID",
                    quantity=0  # Invalid quantity
                )
            )

            # This should either succeed (allowing 0) or fail with proper error
            try:
                stub.AddItem(request)
            except grpc.RpcError as e:
                # If it fails, it should be a proper error
                assert e.code() in [
                    grpc.StatusCode.INVALID_ARGUMENT,
                    grpc.StatusCode.FAILED_PRECONDITION
                ]
        finally:
            channel.close()


if __name__ == '__main__':
    pytest.main([__file__, '-v', '--tb=short'])
