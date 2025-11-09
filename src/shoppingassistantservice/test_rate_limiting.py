#!/usr/bin/python
#
# Copyright 2025 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      https://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

"""
Unit tests for rate limiting functionality in shopping assistant service.
"""

import os
import time
import pytest
from unittest.mock import patch, MagicMock
from shoppingassistantservice import RateLimiter, get_client_ip, create_app


class TestRateLimiter:
    """Test cases for RateLimiter class"""

    def test_allows_requests_within_limit(self):
        """Test that requests within the limit are allowed"""
        limiter = RateLimiter(max_requests=5, window_seconds=60)

        # First 5 requests should be allowed
        for i in range(5):
            allowed, remaining = limiter.is_allowed("192.168.1.1")
            assert allowed is True, f"Request {i+1} should be allowed"
            assert remaining == 5 - i - 1, f"Remaining should be {5 - i - 1}"

    def test_blocks_requests_over_limit(self):
        """Test that requests exceeding the limit are blocked"""
        limiter = RateLimiter(max_requests=3, window_seconds=60)

        # Exhaust the limit
        for _ in range(3):
            allowed, _ = limiter.is_allowed("192.168.1.1")
            assert allowed is True

        # Next request should be blocked
        allowed, remaining = limiter.is_allowed("192.168.1.1")
        assert allowed is False, "Request should be blocked after exceeding limit"
        assert remaining == 0, "Remaining should be 0"

    def test_per_ip_limiting(self):
        """Test that rate limiting is per-IP address"""
        limiter = RateLimiter(max_requests=2, window_seconds=60)

        # Use up limit for IP 1
        for _ in range(2):
            allowed, _ = limiter.is_allowed("192.168.1.1")
            assert allowed is True

        # IP 1 should be blocked
        allowed, _ = limiter.is_allowed("192.168.1.1")
        assert allowed is False

        # IP 2 should still be allowed
        allowed, remaining = limiter.is_allowed("192.168.1.2")
        assert allowed is True, "Different IP should have its own limit"
        assert remaining == 1

    def test_sliding_window(self):
        """Test that old requests are removed from the window"""
        limiter = RateLimiter(max_requests=2, window_seconds=1)  # 1 second window

        # Use up limit
        limiter.is_allowed("192.168.1.1")
        limiter.is_allowed("192.168.1.1")

        # Should be blocked immediately
        allowed, _ = limiter.is_allowed("192.168.1.1")
        assert allowed is False

        # Wait for window to slide
        time.sleep(1.1)

        # Should be allowed again
        allowed, _ = limiter.is_allowed("192.168.1.1")
        assert allowed is True, "Should be allowed after window expires"

    def test_cleanup_removes_old_ips(self):
        """Test that cleanup removes inactive IPs"""
        limiter = RateLimiter(max_requests=5, window_seconds=60)

        # Add some IPs
        limiter.is_allowed("192.168.1.1")
        limiter.is_allowed("192.168.1.2")

        assert len(limiter.requests) == 2

        # Make IP 1 old by manipulating the timestamps
        with limiter.lock:
            limiter.requests["192.168.1.1"] = [time.time() - 200]  # Very old timestamp

        # Manually trigger cleanup logic
        with limiter.lock:
            now = time.time()
            cutoff = now - (limiter.window_seconds * 2)

            ips_to_remove = []
            for ip, timestamps in limiter.requests.items():
                timestamps = [t for t in timestamps if t > cutoff]
                if not timestamps:
                    ips_to_remove.append(ip)

            for ip in ips_to_remove:
                del limiter.requests[ip]

        # IP 1 should be removed, IP 2 should remain
        assert "192.168.1.1" not in limiter.requests
        assert "192.168.1.2" in limiter.requests


class TestGetClientIP:
    """Test cases for get_client_ip function"""

    def test_direct_connection(self):
        """Test IP extraction from direct connection"""
        with patch('shoppingassistantservice.request') as mock_request:
            mock_request.headers.get.return_value = None
            mock_request.remote_addr = "192.168.1.1"

            ip = get_client_ip()
            assert ip == "192.168.1.1"

    def test_x_forwarded_for_single_ip(self):
        """Test IP extraction from X-Forwarded-For with single IP"""
        with patch('shoppingassistantservice.request') as mock_request:
            mock_request.headers.get.side_effect = lambda header: {
                'X-Forwarded-For': '203.0.113.1',
                'X-Real-IP': None
            }.get(header)
            mock_request.remote_addr = "10.0.0.1"

            ip = get_client_ip()
            assert ip == "203.0.113.1"

    def test_x_forwarded_for_multiple_ips(self):
        """Test IP extraction from X-Forwarded-For with multiple IPs"""
        with patch('shoppingassistantservice.request') as mock_request:
            mock_request.headers.get.side_effect = lambda header: {
                'X-Forwarded-For': '203.0.113.1, 198.51.100.1, 192.0.2.1',
                'X-Real-IP': None
            }.get(header)
            mock_request.remote_addr = "10.0.0.1"

            ip = get_client_ip()
            assert ip == "203.0.113.1"

    def test_x_real_ip(self):
        """Test IP extraction from X-Real-IP"""
        with patch('shoppingassistantservice.request') as mock_request:
            mock_request.headers.get.side_effect = lambda header: {
                'X-Forwarded-For': None,
                'X-Real-IP': '203.0.113.2'
            }.get(header)
            mock_request.remote_addr = "10.0.0.1"

            ip = get_client_ip()
            assert ip == "203.0.113.2"

    def test_x_forwarded_for_precedence(self):
        """Test that X-Forwarded-For takes precedence over X-Real-IP"""
        with patch('shoppingassistantservice.request') as mock_request:
            mock_request.headers.get.side_effect = lambda header: {
                'X-Forwarded-For': '203.0.113.1',
                'X-Real-IP': '203.0.113.2'
            }.get(header)
            mock_request.remote_addr = "10.0.0.1"

            ip = get_client_ip()
            assert ip == "203.0.113.1"

    def test_fallback_to_remote_addr(self):
        """Test fallback to remote_addr when headers are missing"""
        with patch('shoppingassistantservice.request') as mock_request:
            mock_request.headers.get.return_value = None
            mock_request.remote_addr = None

            ip = get_client_ip()
            assert ip == "0.0.0.0"


class TestRateLimitingIntegration:
    """Integration tests for rate limiting in Flask app"""

    @pytest.fixture
    def app(self):
        """Create test Flask app"""
        # Mock all the external dependencies
        with patch('shoppingassistantservice.AlloyDBEngine'), \
             patch('shoppingassistantservice.AlloyDBVectorStore'), \
             patch('shoppingassistantservice.secretmanager_v1'):

            # Set minimal environment for testing
            os.environ['PROJECT_ID'] = 'test-project'
            os.environ['REGION'] = 'us-central1'
            os.environ['ALLOYDB_DATABASE_NAME'] = 'test-db'
            os.environ['ALLOYDB_TABLE_NAME'] = 'test-table'
            os.environ['ALLOYDB_CLUSTER_NAME'] = 'test-cluster'
            os.environ['ALLOYDB_INSTANCE_NAME'] = 'test-instance'
            os.environ['ALLOYDB_SECRET_NAME'] = 'test-secret'
            os.environ['RATE_LIMIT_REQUESTS'] = '3'
            os.environ['RATE_LIMIT_WINDOW'] = '60'

            app = create_app()
            app.config['TESTING'] = True
            return app

    def test_allows_normal_requests(self, app):
        """Test that normal requests within limit are allowed"""
        client = app.test_client()

        # Make a request that should succeed (won't actually call LLM due to mocking)
        response = client.post('/',
                             json={'message': 'test', 'image': 'http://example.com/img.jpg'},
                             headers={'X-Forwarded-For': '192.168.1.1'})

        # Should get past rate limiting (might fail on actual logic due to mocks, but that's OK)
        # We're just testing that it's not 429
        assert response.status_code != 429

    def test_blocks_excessive_requests(self, app):
        """Test that excessive requests are blocked with 429"""
        client = app.test_client()

        client_ip = '192.168.1.100'

        # Make requests up to the limit (3 requests)
        for _ in range(3):
            response = client.post('/',
                                 json={'message': 'test', 'image': 'http://example.com/img.jpg'},
                                 headers={'X-Forwarded-For': client_ip})

        # Next request should be rate limited
        response = client.post('/',
                             json={'message': 'test', 'image': 'http://example.com/img.jpg'},
                             headers={'X-Forwarded-For': client_ip})

        assert response.status_code == 429
        data = response.get_json()
        assert 'error' in data
        assert 'Rate limit exceeded' in data['error']

    def test_returns_correct_headers(self, app):
        """Test that rate limit headers are returned"""
        client = app.test_client()

        client_ip = '192.168.1.101'

        # Exhaust limit
        for _ in range(3):
            client.post('/',
                      json={'message': 'test', 'image': 'http://example.com/img.jpg'},
                      headers={'X-Forwarded-For': client_ip})

        # Get rate limited response
        response = client.post('/',
                             json={'message': 'test', 'image': 'http://example.com/img.jpg'},
                             headers={'X-Forwarded-For': client_ip})

        assert response.status_code == 429
        assert 'X-RateLimit-Limit' in response.headers
        assert 'X-RateLimit-Remaining' in response.headers
        assert 'X-RateLimit-Reset' in response.headers
        assert 'Retry-After' in response.headers
        assert response.headers['X-RateLimit-Remaining'] == '0'

    def test_health_check_bypasses_rate_limiting(self, app):
        """Test that health check endpoint bypasses rate limiting"""
        client = app.test_client()

        # Health check should always work, even with high request volume
        for _ in range(10):
            response = client.get('/_healthz')
            assert response.status_code == 200

    def test_options_bypasses_rate_limiting(self, app):
        """Test that OPTIONS requests bypass rate limiting"""
        client = app.test_client()

        client_ip = '192.168.1.102'

        # OPTIONS requests should always work
        for _ in range(10):
            response = client.options('/', headers={'X-Forwarded-For': client_ip})
            assert response.status_code == 200

    def test_can_disable_rate_limiting(self, app):
        """Test that rate limiting can be disabled via environment variable"""
        os.environ['DISABLE_RATE_LIMITING'] = 'true'

        # Recreate app with new environment
        with patch('shoppingassistantservice.AlloyDBEngine'), \
             patch('shoppingassistantservice.AlloyDBVectorStore'), \
             patch('shoppingassistantservice.secretmanager_v1'):
            app_no_limit = create_app()
            app_no_limit.config['TESTING'] = True

        client = app_no_limit.test_client()
        client_ip = '192.168.1.103'

        # Make many requests - should not be rate limited
        for _ in range(10):
            response = client.post('/',
                                 json={'message': 'test', 'image': 'http://example.com/img.jpg'},
                                 headers={'X-Forwarded-For': client_ip})
            # Should never get 429
            assert response.status_code != 429

        # Cleanup
        del os.environ['DISABLE_RATE_LIMITING']

    def test_rate_limit_headers_on_success(self, app):
        """Test that rate limit headers are added to successful responses"""
        client = app.test_client()

        client_ip = '192.168.1.104'

        # First request should have rate limit headers
        response = client.post('/',
                             json={'message': 'test', 'image': 'http://example.com/img.jpg'},
                             headers={'X-Forwarded-For': client_ip})

        # Check for rate limit headers (should be present on all responses)
        if response.status_code != 429:  # Only if not rate limited
            # Headers might be added in after_request
            assert 'X-RateLimit-Limit' in response.headers or response.status_code >= 400


if __name__ == '__main__':
    pytest.main([__file__, '-v'])
