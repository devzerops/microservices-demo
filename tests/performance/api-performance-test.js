/**
 * API Performance Test
 *
 * Direct testing of backend gRPC/HTTP APIs for performance benchmarking
 */

import http from 'k6/http';
import { check, group } from 'k6';
import { Trend, Counter } from 'k6/metrics';

// Custom metrics for each API
const productListDuration = new Trend('api_product_list_duration');
const productGetDuration = new Trend('api_product_get_duration');
const recommendationDuration = new Trend('api_recommendation_duration');
const apiCalls = new Counter('api_calls_total');

export const options = {
  scenarios: {
    product_catalog_api: {
      executor: 'constant-arrival-rate',
      rate: 100, // 100 requests per second
      timeUnit: '1s',
      duration: '5m',
      preAllocatedVUs: 50,
      maxVUs: 200,
      exec: 'testProductCatalogAPI',
    },
    recommendation_api: {
      executor: 'constant-arrival-rate',
      rate: 50,
      timeUnit: '1s',
      duration: '5m',
      preAllocatedVUs: 25,
      maxVUs: 100,
      exec: 'testRecommendationAPI',
    },
  },

  thresholds: {
    'api_product_list_duration': ['p(95)<200', 'p(99)<500'],
    'api_product_get_duration': ['p(95)<100', 'p(99)<300'],
    'api_recommendation_duration': ['p(95)<300', 'p(99)<700'],
    'http_req_failed{api:productcatalog}': ['rate<0.01'],
    'http_req_failed{api:recommendation}': ['rate<0.01'],
  },
};

const PRODUCT_CATALOG_URL = __ENV.PRODUCT_CATALOG_URL || 'http://localhost:3550';
const RECOMMENDATION_URL = __ENV.RECOMMENDATION_URL || 'http://localhost:8080';

export function testProductCatalogAPI() {
  group('ProductCatalog API', () => {
    apiCalls.add(1);

    // Test ListProducts
    let start = Date.now();
    let res = http.get(`${PRODUCT_CATALOG_URL}/products`, {
      tags: { api: 'productcatalog', endpoint: 'list' },
    });
    productListDuration.add(Date.now() - start);

    check(res, {
      'list products returns 200': (r) => r.status === 200,
      'list products has data': (r) => r.json('products').length > 0,
    });

    // Test GetProduct
    start = Date.now();
    res = http.get(`${PRODUCT_CATALOG_URL}/products/OLJCESPC7Z`, {
      tags: { api: 'productcatalog', endpoint: 'get' },
    });
    productGetDuration.add(Date.now() - start);

    check(res, {
      'get product returns 200': (r) => r.status === 200,
      'get product has id': (r) => r.json('id') !== undefined,
    });
  });
}

export function testRecommendationAPI() {
  group('Recommendation API', () => {
    apiCalls.add(1);

    const start = Date.now();
    const res = http.post(
      `${RECOMMENDATION_URL}/recommendations`,
      JSON.stringify({
        user_id: `user-${__VU}`,
        product_ids: ['OLJCESPC7Z'],
      }),
      {
        headers: { 'Content-Type': 'application/json' },
        tags: { api: 'recommendation' },
      }
    );
    recommendationDuration.add(Date.now() - start);

    check(res, {
      'recommendations return 200': (r) => r.status === 200,
      'recommendations not empty': (r) => r.json('product_ids').length > 0,
      'recommendations max 5 items': (r) => r.json('product_ids').length <= 5,
    });
  });
}

export function handleSummary(data) {
  return {
    'results/api-performance-summary.json': JSON.stringify(data, null, 2),
    'stdout': textSummary(data, { indent: ' ', enableColors: true }),
  };
}

function textSummary(data, options) {
  const indent = options.indent || '';
  let summary = '\n';

  summary += `${indent}API Performance Test Results:\n`;
  summary += `${indent}==============================\n\n`;

  // Product Catalog API
  summary += `${indent}Product Catalog API:\n`;
  summary += `${indent}  List Products p95: ${data.metrics.api_product_list_duration?.values['p(95)']?.toFixed(2) || 'N/A'}ms\n`;
  summary += `${indent}  Get Product p95: ${data.metrics.api_product_get_duration?.values['p(95)']?.toFixed(2) || 'N/A'}ms\n`;

  // Recommendation API
  summary += `${indent}\nRecommendation API:\n`;
  summary += `${indent}  p95: ${data.metrics.api_recommendation_duration?.values['p(95)']?.toFixed(2) || 'N/A'}ms\n`;

  // Total calls
  summary += `${indent}\nTotal API Calls: ${data.metrics.api_calls_total?.values.count || 0}\n`;

  return summary;
}
