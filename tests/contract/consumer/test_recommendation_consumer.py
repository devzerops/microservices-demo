"""
Consumer Contract Test for Recommendation Service

This test defines the contract from the Recommendation Service's perspective
as a consumer of the Product Catalog Service.
"""

import pytest
import atexit
import os
import sys
from pact import Consumer, Provider, Like, EachLike, Term

# Add proto path
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '../../../src/recommendationservice'))

# Pact configuration
pact = Consumer('RecommendationService').has_pact_with(
    Provider('ProductCatalogService'),
    pact_dir='../pacts'
)

# Clean up Pact mock service on exit
atexit.register(pact.stop)


class TestRecommendationServiceContract:
    """
    Consumer contract tests for RecommendationService consuming ProductCatalogService
    """

    def test_get_product_list_contract(self):
        """
        Test contract: RecommendationService expects ProductCatalogService
        to return a list of products
        """
        expected_response = {
            'products': EachLike({
                'id': Like('OLJCESPC7Z'),
                'name': Like('Sunglasses'),
                'description': Like('Add a modern touch to your wardrobe'),
                'picture': Like('/static/img/products/sunglasses.jpg'),
                'price_usd': {
                    'currency_code': Like('USD'),
                    'units': Like(19),
                    'nanos': Like(990000000)
                },
                'categories': EachLike('accessories')
            }, minimum=1)
        }

        (pact
         .given('products exist in catalog')
         .upon_receiving('a request for all products')
         .with_request('get', '/products')
         .will_respond_with(200, body=expected_response))

        with pact:
            # This is where you would make the actual HTTP request
            # For gRPC services, you'd use a gRPC-HTTP bridge or test the REST API
            # For demonstration, we're showing the contract definition
            result = self.get_products_from_catalog()

            assert 'products' in result
            assert len(result['products']) > 0
            assert 'id' in result['products'][0]
            assert 'name' in result['products'][0]
            assert 'price_usd' in result['products'][0]

    def test_get_specific_product_contract(self):
        """
        Test contract: RecommendationService expects to get a specific product by ID
        """
        product_id = 'OLJCESPC7Z'

        expected_response = {
            'id': Like(product_id),
            'name': Like('Sunglasses'),
            'description': Like('Add a modern touch to your wardrobe'),
            'price_usd': {
                'currency_code': Like('USD'),
                'units': Like(19),
                'nanos': Like(990000000)
            }
        }

        (pact
         .given(f'product {product_id} exists')
         .upon_receiving(f'a request for product {product_id}')
         .with_request('get', f'/products/{product_id}')
         .will_respond_with(200, body=expected_response))

        with pact:
            result = self.get_product_by_id(product_id)

            assert result['id'] == product_id
            assert 'name' in result
            assert 'price_usd' in result

    def test_product_not_found_contract(self):
        """
        Test contract: ProductCatalogService returns 404 for non-existent products
        """
        product_id = 'NONEXISTENT'

        (pact
         .given('product does not exist')
         .upon_receiving('a request for a non-existent product')
         .with_request('get', f'/products/{product_id}')
         .will_respond_with(404, body={'error': Like('Product not found')}))

        with pact:
            result = self.get_product_by_id(product_id)
            assert result is None or 'error' in result

    # Mock implementations for demonstration
    def get_products_from_catalog(self):
        """Mock implementation - in real code, this would call the actual service"""
        # This would be replaced with actual HTTP/gRPC call to mock service
        return {
            'products': [{
                'id': 'OLJCESPC7Z',
                'name': 'Sunglasses',
                'description': 'Add a modern touch',
                'price_usd': {
                    'currency_code': 'USD',
                    'units': 19,
                    'nanos': 990000000
                }
            }]
        }

    def get_product_by_id(self, product_id):
        """Mock implementation - in real code, this would call the actual service"""
        if product_id == 'NONEXISTENT':
            return None
        return {
            'id': product_id,
            'name': 'Sunglasses',
            'price_usd': {
                'currency_code': 'USD',
                'units': 19,
                'nanos': 990000000
            }
        }


class TestCheckoutServiceContract:
    """
    Consumer contract tests for CheckoutService consuming multiple services
    """

    def setup_method(self):
        """Set up Pact for CheckoutService"""
        self.pact = Consumer('CheckoutService').has_pact_with(
            Provider('PaymentService'),
            pact_dir='../pacts'
        )
        self.pact.start()

    def teardown_method(self):
        """Tear down Pact"""
        self.pact.stop()

    def test_charge_credit_card_contract(self):
        """
        Test contract: CheckoutService expects PaymentService to charge a credit card
        """
        payment_request = {
            'amount': {
                'currency_code': 'USD',
                'units': 50,
                'nanos': 990000000
            },
            'credit_card': {
                'credit_card_number': '4432-8015-6152-0454',
                'credit_card_cvv': 672,
                'credit_card_expiration_year': 2025,
                'credit_card_expiration_month': 1
            }
        }

        expected_response = {
            'transaction_id': Term(r'^[A-Z0-9\-]+$', 'TXN-12345-ABCDE')
        }

        (self.pact
         .given('payment service is available')
         .upon_receiving('a request to charge credit card')
         .with_request(
             'post',
             '/charge',
             body=payment_request,
             headers={'Content-Type': 'application/json'}
         )
         .will_respond_with(200, body=expected_response))

        with self.pact:
            result = self.charge_credit_card(payment_request)
            assert 'transaction_id' in result
            assert len(result['transaction_id']) > 0

    def test_payment_declined_contract(self):
        """
        Test contract: PaymentService returns error for declined payments
        """
        invalid_payment = {
            'amount': {
                'currency_code': 'USD',
                'units': 99999,  # Amount too large
                'nanos': 0
            },
            'credit_card': {
                'credit_card_number': '0000-0000-0000-0000',
                'credit_card_cvv': 000,
                'credit_card_expiration_year': 2020,  # Expired
                'credit_card_expiration_month': 1
            }
        }

        (self.pact
         .given('payment will be declined')
         .upon_receiving('a request with invalid credit card')
         .with_request(
             'post',
             '/charge',
             body=invalid_payment,
             headers={'Content-Type': 'application/json'}
         )
         .will_respond_with(
             402,
             body={'error': Like('Payment declined')}
         ))

        with self.pact:
            result = self.charge_credit_card(invalid_payment)
            assert result is None or 'error' in result

    def charge_credit_card(self, payment_data):
        """Mock implementation"""
        if payment_data['credit_card']['credit_card_number'] == '0000-0000-0000-0000':
            return {'error': 'Payment declined'}
        return {'transaction_id': 'TXN-12345-ABCDE'}


if __name__ == '__main__':
    pytest.main([__file__, '-v'])
