#!/usr/bin/python
#
# Copyright 2018 Google LLC
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
import sys
import os

# Add parent directory to path to import modules
sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

import demo_pb2
import demo_pb2_grpc
from grpc_health.v1 import health_pb2
from email_server import (
    BaseEmailService,
    DummyEmailService,
    HealthCheck,
)


class TestBaseEmailService:
    """Test the BaseEmailService class"""

    def test_check_returns_serving(self):
        """Test that Check method returns SERVING status"""
        service = BaseEmailService()
        request = health_pb2.HealthCheckRequest()
        context = Mock()

        response = service.Check(request, context)

        assert response.status == health_pb2.HealthCheckResponse.SERVING

    def test_watch_returns_unimplemented(self):
        """Test that Watch method returns UNIMPLEMENTED status"""
        service = BaseEmailService()
        request = health_pb2.HealthCheckRequest()
        context = Mock()

        response = service.Watch(request, context)

        assert response.status == health_pb2.HealthCheckResponse.UNIMPLEMENTED


class TestDummyEmailService:
    """Test the DummyEmailService class"""

    def test_send_order_confirmation_success(self):
        """Test successful order confirmation email"""
        service = DummyEmailService()

        # Create a mock order
        order = demo_pb2.OrderResult(
            order_id="test-order-123",
            shipping_tracking_id="track-456",
            shipping_cost=demo_pb2.Money(currency_code="USD", units=5, nanos=0),
            shipping_address=demo_pb2.Address(
                street_address="123 Test St",
                city="Test City",
                state="TC",
                country="Test Country",
                zip_code=12345
            ),
            items=[]
        )

        request = demo_pb2.SendOrderConfirmationRequest(
            email="test@example.com",
            order=order
        )
        context = Mock()

        response = service.SendOrderConfirmation(request, context)

        # Should return Empty message
        assert isinstance(response, demo_pb2.Empty)

    def test_send_order_confirmation_with_items(self):
        """Test order confirmation with order items"""
        service = DummyEmailService()

        # Create order with items
        item = demo_pb2.OrderItem(
            item=demo_pb2.CartItem(
                product_id="PROD-001",
                quantity=2
            ),
            cost=demo_pb2.Money(currency_code="USD", units=50, nanos=0)
        )

        order = demo_pb2.OrderResult(
            order_id="test-order-456",
            shipping_tracking_id="track-789",
            shipping_cost=demo_pb2.Money(currency_code="USD", units=10, nanos=0),
            shipping_address=demo_pb2.Address(
                street_address="456 Test Ave",
                city="Test Town",
                state="TT",
                country="Testland",
                zip_code=54321
            ),
            items=[item]
        )

        request = demo_pb2.SendOrderConfirmationRequest(
            email="customer@example.com",
            order=order
        )
        context = Mock()

        response = service.SendOrderConfirmation(request, context)

        assert isinstance(response, demo_pb2.Empty)

    def test_send_order_confirmation_multiple_items(self):
        """Test order confirmation with multiple items"""
        service = DummyEmailService()

        items = [
            demo_pb2.OrderItem(
                item=demo_pb2.CartItem(product_id=f"PROD-{i}", quantity=i+1),
                cost=demo_pb2.Money(currency_code="USD", units=10*(i+1), nanos=0)
            )
            for i in range(3)
        ]

        order = demo_pb2.OrderResult(
            order_id="test-multi-order",
            shipping_tracking_id="track-multi",
            shipping_cost=demo_pb2.Money(currency_code="USD", units=15, nanos=500000000),
            shipping_address=demo_pb2.Address(
                street_address="789 Multi St",
                city="Multi City",
                state="MC",
                country="Multiland",
                zip_code=99999
            ),
            items=items
        )

        request = demo_pb2.SendOrderConfirmationRequest(
            email="multi@example.com",
            order=order
        )
        context = Mock()

        response = service.SendOrderConfirmation(request, context)

        assert isinstance(response, demo_pb2.Empty)

    def test_send_order_confirmation_different_currencies(self):
        """Test order confirmation with different currencies"""
        service = DummyEmailService()

        for currency in ["USD", "EUR", "GBP", "JPY"]:
            order = demo_pb2.OrderResult(
                order_id=f"order-{currency}",
                shipping_tracking_id=f"track-{currency}",
                shipping_cost=demo_pb2.Money(currency_code=currency, units=10, nanos=0),
                shipping_address=demo_pb2.Address(
                    street_address="Currency St",
                    city="Currency City",
                    state="CC",
                    country="Currency Country",
                    zip_code=11111
                ),
                items=[]
            )

            request = demo_pb2.SendOrderConfirmationRequest(
                email=f"{currency.lower()}@example.com",
                order=order
            )
            context = Mock()

            response = service.SendOrderConfirmation(request, context)
            assert isinstance(response, demo_pb2.Empty)

    def test_inherits_from_base_email_service(self):
        """Test that DummyEmailService inherits from BaseEmailService"""
        service = DummyEmailService()

        assert isinstance(service, BaseEmailService)

        # Test inherited Check method
        request = health_pb2.HealthCheckRequest()
        context = Mock()
        response = service.Check(request, context)
        assert response.status == health_pb2.HealthCheckResponse.SERVING

        # Test inherited Watch method
        response = service.Watch(request, context)
        assert response.status == health_pb2.HealthCheckResponse.UNIMPLEMENTED


class TestHealthCheck:
    """Test the HealthCheck class"""

    def test_health_check_returns_serving(self):
        """Test that HealthCheck returns SERVING status"""
        health = HealthCheck()
        request = health_pb2.HealthCheckRequest()
        context = Mock()

        response = health.Check(request, context)

        assert response.status == health_pb2.HealthCheckResponse.SERVING


class TestTemplateRendering:
    """Test template rendering functionality"""

    def test_template_exists(self):
        """Test that the confirmation template exists"""
        from jinja2 import Environment, FileSystemLoader, select_autoescape

        env = Environment(
            loader=FileSystemLoader('templates'),
            autoescape=select_autoescape(['html', 'xml'])
        )
        template = env.get_template('confirmation.html')

        assert template is not None

    def test_template_renders_with_order(self):
        """Test that template can render with an order"""
        from jinja2 import Environment, FileSystemLoader, select_autoescape

        env = Environment(
            loader=FileSystemLoader('templates'),
            autoescape=select_autoescape(['html', 'xml'])
        )
        template = env.get_template('confirmation.html')

        order = demo_pb2.OrderResult(
            order_id="TEST-123",
            shipping_tracking_id="TRACK-456",
            shipping_cost=demo_pb2.Money(currency_code="USD", units=5, nanos=990000000),
            shipping_address=demo_pb2.Address(
                street_address="123 Main St",
                city="Anytown",
                state="CA",
                country="USA",
                zip_code=12345
            ),
            items=[]
        )

        # Render template
        html = template.render(order=order)

        assert html is not None
        assert len(html) > 0
        assert "TEST-123" in html  # Order ID should be in rendered HTML

    def test_template_renders_with_items(self):
        """Test that template renders order items correctly"""
        from jinja2 import Environment, FileSystemLoader, select_autoescape

        env = Environment(
            loader=FileSystemLoader('templates'),
            autoescape=select_autoescape(['html', 'xml'])
        )
        template = env.get_template('confirmation.html')

        items = [
            demo_pb2.OrderItem(
                item=demo_pb2.CartItem(product_id="PROD-001", quantity=2),
                cost=demo_pb2.Money(currency_code="USD", units=25, nanos=500000000)
            ),
            demo_pb2.OrderItem(
                item=demo_pb2.CartItem(product_id="PROD-002", quantity=1),
                cost=demo_pb2.Money(currency_code="USD", units=15, nanos=0)
            )
        ]

        order = demo_pb2.OrderResult(
            order_id="MULTI-ITEM",
            shipping_tracking_id="TRACK-999",
            shipping_cost=demo_pb2.Money(currency_code="USD", units=7, nanos=500000000),
            shipping_address=demo_pb2.Address(
                street_address="789 Oak Ave",
                city="Somewhere",
                state="NY",
                country="USA",
                zip_code=54321
            ),
            items=items
        )

        html = template.render(order=order)

        assert html is not None
        assert "MULTI-ITEM" in html
        assert "PROD-001" in html
        assert "PROD-002" in html
