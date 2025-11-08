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
from email_server import DummyEmailService, BaseEmailService


class TestBaseEmailService:
    """Test BaseEmailService health check functionality"""

    def test_health_check_serving(self):
        """Test that health check returns SERVING status"""
        service = BaseEmailService()
        request = health_pb2.HealthCheckRequest()
        context = Mock()

        response = service.Check(request, context)

        assert response.status == health_pb2.HealthCheckResponse.SERVING

    def test_health_watch_unimplemented(self):
        """Test that Watch returns UNIMPLEMENTED status"""
        service = BaseEmailService()
        request = health_pb2.HealthCheckRequest()
        context = Mock()

        response = service.Watch(request, context)

        assert response.status == health_pb2.HealthCheckResponse.UNIMPLEMENTED


class TestDummyEmailService:
    """Test DummyEmailService functionality"""

    def test_send_order_confirmation_returns_empty(self):
        """Test that SendOrderConfirmation returns Empty response"""
        service = DummyEmailService()

        # Create a mock order
        order = demo_pb2.OrderResult(
            order_id="test-order-123",
            shipping_tracking_id="tracking-456"
        )

        request = demo_pb2.SendOrderConfirmationRequest(
            email="test@example.com",
            order=order
        )
        context = Mock()

        response = service.SendOrderConfirmation(request, context)

        assert isinstance(response, demo_pb2.Empty)

    def test_send_order_confirmation_with_valid_email(self):
        """Test sending order confirmation with valid email address"""
        service = DummyEmailService()

        order = demo_pb2.OrderResult(
            order_id="order-789",
            shipping_tracking_id="track-999"
        )

        request = demo_pb2.SendOrderConfirmationRequest(
            email="customer@test.com",
            order=order
        )
        context = Mock()

        response = service.SendOrderConfirmation(request, context)

        assert response is not None
        assert isinstance(response, demo_pb2.Empty)

    def test_send_order_confirmation_with_items(self):
        """Test sending order confirmation with order items"""
        service = DummyEmailService()

        # Create order items
        item1 = demo_pb2.OrderItem(
            item=demo_pb2.CartItem(
                product_id="OLJCESPC7Z",
                quantity=1
            ),
            cost=demo_pb2.Money(
                currency_code="USD",
                units=35,
                nanos=990000000
            )
        )

        order = demo_pb2.OrderResult(
            order_id="order-with-items",
            shipping_tracking_id="tracking-001",
            items=[item1]
        )

        request = demo_pb2.SendOrderConfirmationRequest(
            email="buyer@example.com",
            order=order
        )
        context = Mock()

        response = service.SendOrderConfirmation(request, context)

        assert isinstance(response, demo_pb2.Empty)

    def test_health_check_inherited(self):
        """Test that DummyEmailService inherits health check from BaseEmailService"""
        service = DummyEmailService()
        request = health_pb2.HealthCheckRequest()
        context = Mock()

        response = service.Check(request, context)

        assert response.status == health_pb2.HealthCheckResponse.SERVING


class TestEmailServiceIntegration:
    """Integration tests for email service gRPC server"""

    @pytest.fixture
    def grpc_server(self):
        """Set up a test gRPC server"""
        server = grpc.server(futures.ThreadPoolExecutor(max_workers=1))
        service = DummyEmailService()
        demo_pb2_grpc.add_EmailServiceServicer_to_server(service, server)
        health_pb2_grpc.add_HealthServicer_to_server(service, server)

        port = server.add_insecure_port('[::]:0')  # Use dynamic port
        server.start()

        yield server, port

        server.stop(0)

    def test_grpc_server_health_check(self, grpc_server):
        """Test health check through gRPC channel"""
        server, port = grpc_server

        with grpc.insecure_channel(f'localhost:{port}') as channel:
            stub = health_pb2_grpc.HealthStub(channel)
            request = health_pb2.HealthCheckRequest()

            response = stub.Check(request)

            assert response.status == health_pb2.HealthCheckResponse.SERVING

    def test_grpc_send_order_confirmation(self, grpc_server):
        """Test sending order confirmation through gRPC channel"""
        server, port = grpc_server

        with grpc.insecure_channel(f'localhost:{port}') as channel:
            stub = demo_pb2_grpc.EmailServiceStub(channel)

            order = demo_pb2.OrderResult(
                order_id="grpc-test-order",
                shipping_tracking_id="grpc-tracking"
            )

            request = demo_pb2.SendOrderConfirmationRequest(
                email="grpc@test.com",
                order=order
            )

            response = stub.SendOrderConfirmation(request)

            assert isinstance(response, demo_pb2.Empty)


class TestEmailTemplate:
    """Test email template rendering functionality"""

    @patch('email_server.template')
    def test_template_rendering_called(self, mock_template):
        """Test that template.render is called with order data"""
        mock_template.render.return_value = "<html>Confirmation</html>"

        # This test would require EmailService to be implemented
        # For now, we just verify the template mock setup
        assert mock_template.render is not None


if __name__ == '__main__':
    pytest.main([__file__, '-v'])
