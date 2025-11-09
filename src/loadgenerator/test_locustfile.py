#!/usr/bin/python
"""
Unit tests for loadgenerator locustfile.
"""

import unittest
from unittest.mock import Mock, MagicMock, patch, call
import random
import datetime
from locustfile import (
    index,
    setCurrency,
    browseProduct,
    viewCart,
    addToCart,
    empty_cart,
    checkout,
    logout,
    products,
    UserBehavior,
    WebsiteUser
)


class TestLocustfileFunctions(unittest.TestCase):
    """Test individual task functions in locustfile."""

    def setUp(self):
        """Set up mock Locust user for each test."""
        self.mock_locust = Mock()
        self.mock_locust.client = Mock()

    def test_index(self):
        """Test index function makes GET request to homepage."""
        index(self.mock_locust)
        self.mock_locust.client.get.assert_called_once_with("/")

    def test_setCurrency(self):
        """Test setCurrency function posts valid currency code."""
        with patch('random.choice', return_value='USD'):
            setCurrency(self.mock_locust)
            self.mock_locust.client.post.assert_called_once_with(
                "/setCurrency",
                {'currency_code': 'USD'}
            )

    def test_setCurrency_all_currencies(self):
        """Test setCurrency uses valid currency codes."""
        currencies = ['EUR', 'USD', 'JPY', 'CAD', 'GBP', 'TRY']
        for currency in currencies:
            self.mock_locust.client.reset_mock()
            with patch('random.choice', return_value=currency):
                setCurrency(self.mock_locust)
                self.mock_locust.client.post.assert_called_once()
                call_args = self.mock_locust.client.post.call_args
                self.assertEqual(call_args[0][0], "/setCurrency")
                self.assertEqual(call_args[0][1]['currency_code'], currency)

    def test_browseProduct(self):
        """Test browseProduct function browses a valid product."""
        with patch('random.choice', return_value='0PUK6V6EV0'):
            browseProduct(self.mock_locust)
            self.mock_locust.client.get.assert_called_once_with("/product/0PUK6V6EV0")

    def test_browseProduct_valid_products(self):
        """Test browseProduct only uses products from the product list."""
        for product in products:
            self.mock_locust.client.reset_mock()
            with patch('random.choice', return_value=product):
                browseProduct(self.mock_locust)
                self.mock_locust.client.get.assert_called_once_with(f"/product/{product}")

    def test_viewCart(self):
        """Test viewCart function makes GET request to /cart."""
        viewCart(self.mock_locust)
        self.mock_locust.client.get.assert_called_once_with("/cart")

    def test_addToCart(self):
        """Test addToCart function adds product with quantity to cart."""
        with patch('random.choice', return_value='1YMWWN1N4O'), \
             patch('random.randint', return_value=5):
            addToCart(self.mock_locust)

            # Should first GET the product page
            self.mock_locust.client.get.assert_called_once_with("/product/1YMWWN1N4O")

            # Then POST to cart
            self.mock_locust.client.post.assert_called_once_with("/cart", {
                'product_id': '1YMWWN1N4O',
                'quantity': 5
            })

    def test_addToCart_quantity_range(self):
        """Test addToCart uses quantity between 1 and 10."""
        for qty in [1, 5, 10]:
            self.mock_locust.client.reset_mock()
            with patch('random.choice', return_value='2ZYFJ3GM2N'), \
                 patch('random.randint', return_value=qty):
                addToCart(self.mock_locust)

                call_args = self.mock_locust.client.post.call_args
                self.assertEqual(call_args[0][1]['quantity'], qty)
                self.assertGreaterEqual(qty, 1)
                self.assertLessEqual(qty, 10)

    def test_empty_cart(self):
        """Test empty_cart function makes POST request to /cart/empty."""
        empty_cart(self.mock_locust)
        self.mock_locust.client.post.assert_called_once_with('/cart/empty')

    @patch('locustfile.fake')
    def test_checkout(self, mock_faker):
        """Test checkout function submits order with all required fields."""
        # Mock faker data
        mock_faker.email.return_value = "test@example.com"
        mock_faker.street_address.return_value = "123 Main St"
        mock_faker.zipcode.return_value = "12345"
        mock_faker.city.return_value = "TestCity"
        mock_faker.state_abbr.return_value = "TS"
        mock_faker.country.return_value = "TestCountry"
        mock_faker.credit_card_number.return_value = "4111111111111111"

        with patch('random.choice', return_value='6E92ZMYYFZ'), \
             patch('random.randint', side_effect=[3, 1, 2026, 123]):
            checkout(self.mock_locust)

            # Should add to cart first
            self.assertEqual(self.mock_locust.client.get.call_count, 1)
            self.assertEqual(self.mock_locust.client.post.call_count, 2)

            # Check checkout POST call
            checkout_call = self.mock_locust.client.post.call_args_list[1]
            self.assertEqual(checkout_call[0][0], "/cart/checkout")

            checkout_data = checkout_call[0][1]
            self.assertEqual(checkout_data['email'], "test@example.com")
            self.assertEqual(checkout_data['street_address'], "123 Main St")
            self.assertEqual(checkout_data['zip_code'], "12345")
            self.assertEqual(checkout_data['city'], "TestCity")
            self.assertEqual(checkout_data['state'], "TS")
            self.assertEqual(checkout_data['country'], "TestCountry")
            self.assertEqual(checkout_data['credit_card_number'], "4111111111111111")

    def test_logout(self):
        """Test logout function makes GET request to /logout."""
        logout(self.mock_locust)
        self.mock_locust.client.get.assert_called_once_with('/logout')


class TestUserBehavior(unittest.TestCase):
    """Test UserBehavior TaskSet."""

    def test_on_start_calls_index(self):
        """Test that on_start calls index function."""
        mock_user = Mock()
        mock_user.client = Mock()

        behavior = UserBehavior(mock_user)
        behavior.on_start()

        # Should call GET on "/"
        mock_user.client.get.assert_called_with("/")

    def test_tasks_defined(self):
        """Test that all tasks are defined in UserBehavior."""
        self.assertIsNotNone(UserBehavior.tasks)
        self.assertIsInstance(UserBehavior.tasks, dict)

        # Verify all expected tasks are present
        task_functions = UserBehavior.tasks.keys()
        self.assertIn(index, task_functions)
        self.assertIn(setCurrency, task_functions)
        self.assertIn(browseProduct, task_functions)
        self.assertIn(addToCart, task_functions)
        self.assertIn(viewCart, task_functions)
        self.assertIn(checkout, task_functions)

    def test_task_weights(self):
        """Test that task weights are correctly set."""
        tasks = UserBehavior.tasks
        self.assertEqual(tasks[index], 1)
        self.assertEqual(tasks[setCurrency], 2)
        self.assertEqual(tasks[browseProduct], 10)
        self.assertEqual(tasks[addToCart], 2)
        self.assertEqual(tasks[viewCart], 3)
        self.assertEqual(tasks[checkout], 1)


class TestWebsiteUser(unittest.TestCase):
    """Test WebsiteUser class."""

    def test_tasks_assigned(self):
        """Test that WebsiteUser has tasks assigned."""
        self.assertIsNotNone(WebsiteUser.tasks)
        self.assertIn(UserBehavior, WebsiteUser.tasks)

    def test_wait_time_defined(self):
        """Test that wait_time is defined."""
        self.assertIsNotNone(WebsiteUser.wait_time)


class TestProductList(unittest.TestCase):
    """Test product list configuration."""

    def test_products_not_empty(self):
        """Test that products list is not empty."""
        self.assertGreater(len(products), 0)

    def test_products_are_strings(self):
        """Test that all products are strings."""
        for product in products:
            self.assertIsInstance(product, str)
            self.assertGreater(len(product), 0)

    def test_products_unique(self):
        """Test that all product IDs are unique."""
        self.assertEqual(len(products), len(set(products)))

    def test_expected_products(self):
        """Test that expected products are in the list."""
        expected_products = [
            '0PUK6V6EV0',
            '1YMWWN1N4O',
            '2ZYFJ3GM2N',
            '66VCHSJNUP',
            '6E92ZMYYFZ',
            '9SIQT8TOJO',
            'L9ECAV7KIM',
            'LS4PSXUNUM',
            'OLJCESPC7Z'
        ]
        for product in expected_products:
            self.assertIn(product, products)


if __name__ == '__main__':
    unittest.main()
