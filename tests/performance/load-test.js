/**
 * K6 Load Test - Standard Load Testing
 *
 * This test simulates normal traffic patterns to establish baseline performance.
 * It gradually ramps up to the target load, maintains it, then ramps down.
 */

import http from 'k6/http';
import { check, group, sleep } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');
const productViewDuration = new Trend('product_view_duration');
const checkoutDuration = new Trend('checkout_duration');
const requestCounter = new Counter('requests_total');

// Test configuration
export const options = {
  stages: [
    { duration: '2m', target: 50 },   // Ramp up to 50 users over 2 minutes
    { duration: '5m', target: 50 },   // Stay at 50 users for 5 minutes
    { duration: '2m', target: 100 },  // Ramp up to 100 users
    { duration: '5m', target: 100 },  // Stay at 100 users for 5 minutes
    { duration: '2m', target: 0 },    // Ramp down to 0 users
  ],
  thresholds: {
    http_req_duration: ['p(95)<500', 'p(99)<1000'], // 95% of requests under 500ms, 99% under 1s
    http_req_failed: ['rate<0.05'],                  // Error rate under 5%
    errors: ['rate<0.1'],                            // Custom error rate under 10%
  },
};

// Base URL - can be overridden with -e BASE_URL=http://...
const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

// Sample product IDs (from product catalog)
const PRODUCT_IDS = [
  'OLJCESPC7Z', // Sunglasses
  '66VCHSJNUP', // Tank Top
  '1YMWWN1N4O', // Watch
  'L9ECAV7KIM', // Loafers
  '2ZYFJ3GM2N', // Hairdryer
  '0PUK6V6EV0', // Candle Holder
  'LS4PSXUNUM', // Salt & Pepper Shakers
  '9SIQT8TOJO', // Bamboo Glass Jar
  '6E92ZMYYFZ', // Mug
];

/**
 * Setup function - runs once before all VUs
 */
export function setup() {
  console.log(`Starting load test against ${BASE_URL}`);

  // Verify that the service is accessible
  const res = http.get(BASE_URL);
  check(res, {
    'homepage is accessible': (r) => r.status === 200,
  });

  return { startTime: new Date() };
}

/**
 * Main test function - runs for each VU iteration
 */
export default function (data) {
  requestCounter.add(1);

  // Simulate realistic user behavior
  group('Browse Products', () => {
    browseHomepage();
    sleep(1);

    const productId = PRODUCT_IDS[Math.floor(Math.random() * PRODUCT_IDS.length)];
    viewProductDetail(productId);
    sleep(2);
  });

  // 30% of users add items to cart
  if (Math.random() < 0.3) {
    group('Shopping Cart', () => {
      const productId = PRODUCT_IDS[Math.floor(Math.random() * PRODUCT_IDS.length)];
      addToCart(productId);
      sleep(1);

      viewCart();
      sleep(2);
    });
  }

  // 10% of users complete checkout
  if (Math.random() < 0.1) {
    group('Checkout', () => {
      checkout();
      sleep(1);
    });
  }

  sleep(Math.random() * 3 + 1); // Random think time between 1-4 seconds
}

/**
 * Browse homepage
 */
function browseHomepage() {
  const res = http.get(`${BASE_URL}/`);

  const success = check(res, {
    'homepage status is 200': (r) => r.status === 200,
    'homepage loads in <500ms': (r) => r.timings.duration < 500,
  });

  errorRate.add(!success);
}

/**
 * View product detail page
 */
function viewProductDetail(productId) {
  const start = new Date();
  const res = http.get(`${BASE_URL}/product/${productId}`);
  const duration = new Date() - start;

  productViewDuration.add(duration);

  const success = check(res, {
    'product page status is 200': (r) => r.status === 200,
    'product page has content': (r) => r.body.length > 0,
    'product page loads in <1s': (r) => r.timings.duration < 1000,
  });

  errorRate.add(!success);
}

/**
 * Add item to cart
 */
function addToCart(productId) {
  const payload = {
    product_id: productId,
    quantity: Math.floor(Math.random() * 3) + 1, // 1-3 items
  };

  const params = {
    headers: {
      'Content-Type': 'application/x-www-form-urlencoded',
    },
  };

  const res = http.post(
    `${BASE_URL}/cart`,
    `product_id=${payload.product_id}&quantity=${payload.quantity}`,
    params
  );

  const success = check(res, {
    'add to cart successful': (r) => r.status === 200 || r.status === 303,
  });

  errorRate.add(!success);
}

/**
 * View shopping cart
 */
function viewCart() {
  const res = http.get(`${BASE_URL}/cart`);

  const success = check(res, {
    'cart page status is 200': (r) => r.status === 200,
  });

  errorRate.add(!success);
}

/**
 * Complete checkout
 */
function checkout() {
  const start = new Date();

  const payload = {
    email: `testuser${Math.floor(Math.random() * 10000)}@test.com`,
    street_address: '123 Test Street',
    zip_code: '12345',
    city: 'Test City',
    state: 'TS',
    country: 'Test Country',
    credit_card_number: '4432-8015-6152-0454',
    credit_card_expiration_month: '1',
    credit_card_expiration_year: '2025',
    credit_card_cvv: '672',
  };

  const params = {
    headers: {
      'Content-Type': 'application/x-www-form-urlencoded',
    },
  };

  const formData = Object.keys(payload)
    .map(key => `${encodeURIComponent(key)}=${encodeURIComponent(payload[key])}`)
    .join('&');

  const res = http.post(`${BASE_URL}/cart/checkout`, formData, params);
  const duration = new Date() - start;

  checkoutDuration.add(duration);

  const success = check(res, {
    'checkout successful': (r) => r.status === 200 || r.status === 303,
    'checkout completes in <2s': (r) => r.timings.duration < 2000,
  });

  errorRate.add(!success);
}

/**
 * Teardown function - runs once after all VUs complete
 */
export function teardown(data) {
  const endTime = new Date();
  const duration = (endTime - data.startTime) / 1000;
  console.log(`Load test completed in ${duration} seconds`);
}
