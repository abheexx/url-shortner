import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');

// Test configuration
export const options = {
  stages: [
    // Ramp up to 100 users over 2 minutes
    { duration: '2m', target: 100 },
    // Stay at 100 users for 5 minutes
    { duration: '5m', target: 100 },
    // Ramp up to 500 users over 3 minutes
    { duration: '3m', target: 500 },
    // Stay at 500 users for 5 minutes
    { duration: '5m', target: 500 },
    // Ramp down to 0 users over 2 minutes
    { duration: '2m', target: 0 },
  ],
  thresholds: {
    http_req_duration: ['p(95)<100'], // 95% of requests must complete below 100ms
    http_req_failed: ['rate<0.01'],   // Error rate must be below 1%
    errors: ['rate<0.01'],            // Custom error rate must be below 1%
  },
};

// Test data
const baseURL = __ENV.BASE_URL || 'http://localhost:8080';
const testURLs = [
  'https://www.google.com',
  'https://www.github.com',
  'https://www.stackoverflow.com',
  'https://www.reddit.com',
  'https://www.wikipedia.org',
];

// Shared state
let createdURLs = [];

// Setup function - runs once before the test
export function setup() {
  console.log('Setting up test data...');
  
  // Create some test URLs
  for (let i = 0; i < 10; i++) {
    const testURL = testURLs[i % testURLs.length];
    const payload = JSON.stringify({
      url: testURL,
      custom_alias: `test-${i}`,
    });
    
    const response = http.post(`${baseURL}/api/v1/shorten`, payload, {
      headers: { 'Content-Type': 'application/json' },
    });
    
    if (response.status === 201) {
      const data = JSON.parse(response.body);
      createdURLs.push(data.code);
    }
  }
  
  console.log(`Created ${createdURLs.length} test URLs`);
  return { createdURLs };
}

// Main test function
export default function(data) {
  const { createdURLs } = data;
  
  // Random sleep between requests (0.1 to 0.5 seconds)
  sleep(Math.random() * 0.4 + 0.1);
  
  // 95% of requests are GET (redirects) - simulating real usage
  if (Math.random() < 0.95) {
    // GET request - redirect to long URL
    if (createdURLs.length > 0) {
      const randomCode = createdURLs[Math.floor(Math.random() * createdURLs.length)];
      const response = http.get(`${baseURL}/${randomCode}`, {
        redirects: 0, // Don't follow redirects to measure redirect performance
      });
      
      check(response, {
        'redirect status is 301': (r) => r.status === 301,
        'has location header': (r) => r.headers.Location !== undefined,
        'response time < 100ms': (r) => r.timings.duration < 100,
      });
      
      if (response.status !== 301) {
        errorRate.add(1);
      }
    }
  } else {
    // 5% of requests are POST (create new URLs)
    const testURL = testURLs[Math.floor(Math.random() * testURLs.length)];
    const payload = JSON.stringify({
      url: testURL,
    });
    
    const response = http.post(`${baseURL}/api/v1/shorten`, payload, {
      headers: { 'Content-Type': 'application/json' },
    });
    
    check(response, {
      'create status is 201': (r) => r.status === 201,
      'response has code': (r) => JSON.parse(r.body).code !== undefined,
      'response time < 200ms': (r) => r.timings.duration < 200,
    });
    
    if (response.status !== 201) {
      errorRate.add(1);
    }
  }
  
  // Occasionally check URL metadata
  if (Math.random() < 0.1 && createdURLs.length > 0) {
    const randomCode = createdURLs[Math.floor(Math.random() * createdURLs.length)];
    const response = http.get(`${baseURL}/api/v1/urls/${randomCode}`);
    
    check(response, {
      'metadata status is 200': (r) => r.status === 200,
      'response time < 150ms': (r) => r.timings.duration < 150,
    });
  }
}

// Teardown function - runs once after the test
export function teardown(data) {
  console.log('Cleaning up test data...');
  
  // Clean up created URLs (optional - you might want to keep them for analysis)
  const { createdURLs } = data;
  for (const code of createdURLs) {
    http.del(`${baseURL}/api/v1/urls/${code}`);
  }
  
  console.log('Cleanup completed');
}

// Handle errors
export function handleSummary(data) {
  return {
    'load-test-results.json': JSON.stringify(data, null, 2),
    stdout: textSummary(data, { indent: ' ', enableColors: true }),
  };
}

// Helper function for summary
function textSummary(data, options) {
  const { metrics, root_group } = data;
  const { http_req_duration, http_req_failed, http_reqs } = metrics;
  
  return `
Load Test Results
=================

HTTP Requests:
  Total: ${http_reqs.count}
  Failed: ${http_req_failed.rate * 100}%
  Duration (p95): ${http_req_duration.values['p(95)']}ms

Thresholds:
  Duration p95 < 100ms: ${http_req_duration.values['p(95)'] < 100 ? 'PASS' : 'FAIL'}
  Error rate < 1%: ${http_req_failed.rate < 0.01 ? 'PASS' : 'FAIL'}

Test completed successfully!
`;
}
