/**
 * Black Friday Scenario
 *
 * Simulates Black Friday traffic pattern:
 * - High concurrent users
 * - Heavy cart operations
 * - Many checkouts
 */

import http from 'k6/http';
import { check, group, sleep } from 'k6';
import { SharedArray } from 'k6/data';

// Shared product list
const products = new SharedArray('products', function () {
  return [
    'OLJCESPC7Z', '66VCHSJNUP', '1YMWWN1N4O',
    'L9ECAV7KIM', '2ZYFJ3GM2N', '0PUK6V6EV0',
  ];
});

export const options = {
  scenarios: {
    // Early morning rush (6 AM - 8 AM)
    early_rush: {
      executor: 'ramping-vus',
      startTime: '0s',
      stages: [
        { duration: '10m', target: 500 },
        { duration: '20m', target: 500 },
        { duration: '10m', target: 200 },
      ],
      gracefulRampDown: '5m',
    },

    // Peak hours (12 PM - 2 PM)
    peak_hours: {
      executor: 'ramping-vus',
      startTime: '40m',
      stages: [
        { duration: '5m', target: 1000 },
        { duration: '30m', target: 1000 },
        { duration: '5m', target: 300 },
      ],
      gracefulRampDown: '10m',
    },

    // Sustained load throughout the day
    background_traffic: {
      executor: 'constant-vus',
      vus: 200,
      duration: '2h',
    },
  },

  thresholds: {
    http_req_duration: ['p(95)<1000', 'p(99)<3000'],
    http_req_failed: ['rate<0.05'],
    'group_duration{group:::Checkout}': ['p(95)<5000'],
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

export default function () {
  // 60% browse only
  if (Math.random() < 0.6) {
    browseProducts();
  }
  // 30% add to cart
  else if (Math.random() < 0.9) {
    addMultipleToCart();
  }
  // 10% complete purchase
  else {
    completePurchase();
  }

  sleep(Math.random() * 2);
}

function browseProducts() {
  group('Browse', () => {
    http.get(`${BASE_URL}/`);
    sleep(1);

    const productId = products[Math.floor(Math.random() * products.length)];
    http.get(`${BASE_URL}/product/${productId}`);
    sleep(2);
  });
}

function addMultipleToCart() {
  group('Add to Cart', () => {
    // Add 2-4 items
    const itemCount = Math.floor(Math.random() * 3) + 2;

    for (let i = 0; i < itemCount; i++) {
      const productId = products[Math.floor(Math.random() * products.length)];
      const quantity = Math.floor(Math.random() * 3) + 1;

      http.post(
        `${BASE_URL}/cart`,
        `product_id=${productId}&quantity=${quantity}`,
        { headers: { 'Content-Type': 'application/x-www-form-urlencoded' } }
      );

      sleep(0.5);
    }
  });
}

function completePurchase() {
  group('Checkout', () => {
    addMultipleToCart();
    sleep(1);

    const res = http.post(
      `${BASE_URL}/cart/checkout`,
      `email=customer${__VU}@test.com&street_address=123 Main St&city=NYC&state=NY&zip_code=10001&country=USA&credit_card_number=4432-8015-6152-0454&credit_card_expiration_month=12&credit_card_expiration_year=2025&credit_card_cvv=123`,
      { headers: { 'Content-Type': 'application/x-www-form-urlencoded' } }
    );

    check(res, {
      'checkout successful': (r) => r.status === 200 || r.status === 303,
    });
  });
}
