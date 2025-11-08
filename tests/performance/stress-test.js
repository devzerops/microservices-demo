/**
 * K6 Stress Test - Find Breaking Point
 *
 * This test gradually increases load beyond normal capacity to find
 * the system's breaking point and observe degradation patterns.
 */

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';

const errorRate = new Rate('errors');
const responseTimes = new Trend('response_times');

export const options = {
  stages: [
    { duration: '2m', target: 100 },   // Below normal load
    { duration: '3m', target: 200 },   // Normal load
    { duration: '3m', target: 400 },   // Above normal load
    { duration: '3m', target: 600 },   // Stress level
    { duration: '3m', target: 800 },   // Beyond stress
    { duration: '3m', target: 1000 },  // Breaking point
    { duration: '5m', target: 0 },     // Recovery
  ],
  thresholds: {
    // We expect these to fail - that's the point!
    // We're finding WHERE they fail
    http_req_duration: ['p(95)<2000'],
    http_req_failed: ['rate<0.2'],
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

export default function () {
  const res = http.get(`${BASE_URL}/`);

  responseTimes.add(res.timings.duration);

  const success = check(res, {
    'status is 200': (r) => r.status === 200,
  });

  errorRate.add(!success);

  // Track when degradation starts
  if (res.timings.duration > 1000) {
    console.warn(`Slow response: ${res.timings.duration}ms at ${__VU} VUs`);
  }

  if (!success) {
    console.error(`Failed request at ${__VU} VUs: ${res.status}`);
  }

  sleep(1);
}

export function handleSummary(data) {
  console.log('Stress Test Summary:');
  console.log(`Max VUs: ${Math.max(...data.metrics.vus.values.map(v => v.value))}`);
  console.log(`Error Rate: ${(data.metrics.errors.values.rate * 100).toFixed(2)}%`);
  console.log(`P95 Response Time: ${data.metrics.http_req_duration.values['p(95)']}ms`);

  return {
    'results/stress-test-summary.json': JSON.stringify(data, null, 2),
  };
}
