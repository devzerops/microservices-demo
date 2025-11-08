/**
 * K6 Spike Test - Traffic Spike Testing
 *
 * This test simulates sudden traffic spikes to test system resilience
 * and auto-scaling capabilities.
 */

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

const errorRate = new Rate('errors');

export const options = {
  stages: [
    { duration: '1m', target: 50 },     // Normal load
    { duration: '30s', target: 500 },   // Sudden spike!
    { duration: '2m', target: 500 },    // Maintain spike
    { duration: '30s', target: 50 },    // Drop back
    { duration: '1m', target: 50 },     // Normal load
    { duration: '30s', target: 1000 },  // Even bigger spike!
    { duration: '1m', target: 1000 },   // Maintain
    { duration: '1m', target: 0 },      // Ramp down
  ],
  thresholds: {
    http_req_duration: ['p(99)<3000'],  // More lenient during spike
    http_req_failed: ['rate<0.1'],       // Allow 10% error rate during spike
    errors: ['rate<0.15'],
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

export default function () {
  const endpoints = [
    '/',
    '/product/OLJCESPC7Z',
    '/product/66VCHSJNUP',
    '/cart',
  ];

  const endpoint = endpoints[Math.floor(Math.random() * endpoints.length)];
  const res = http.get(`${BASE_URL}${endpoint}`);

  const success = check(res, {
    'status is 200 or 503': (r) => r.status === 200 || r.status === 503, // 503 OK during spike
    'response time acceptable': (r) => r.timings.duration < 5000,
  });

  errorRate.add(!success);

  sleep(0.5); // Short sleep to maintain high RPS
}
