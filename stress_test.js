import http from 'k6/http';
import { sleep } from 'k6';

export const options = {
  //iterations: 10,
  vus: 500,
  duration: '3m',
  thresholds: {
    http_req_duration: ['p(99)<500'],  // 99% of requests must be < 500ms
    http_req_failed: ['rate<0.01'],    // Less than 1% of requests can fail
 }
};

// The default exported function is gonna be picked up by k6 as the entry point for the test script. It will be executed repeatedly in "iterations" for the whole duration of the test.
export default function () {
  http.get('https://localhost:3002/');
  // Sleep for 1 second to simulate real-world usage
  sleep(1);
}
