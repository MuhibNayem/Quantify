import http from "k6/http";
import { check, sleep } from "k6";
import { Rate, Trend } from "k6/metrics";

const BASE_URL = __ENV.BASE_URL || "https://api.staging.quantify.dev";

const errorRate = new Rate("errors");
const loginDuration = new Trend("login_duration", true);

export const options = {
  stages: [
    { duration: "1m", target: 10 },
    { duration: "3m", target: 50 },
    { duration: "1m", target: 100 },
    { duration: "3m", target: 100 },
    { duration: "1m", target: 0 },
  ],
  thresholds: {
    http_req_duration: ["p(95)<500", "p(99)<1000"],
    errors: ["rate<0.01"],
    login_duration: ["p(95)<800"],
  },
};

export default function () {
  const healthRes = http.get(`${BASE_URL}/health`);
  check(healthRes, { "health is 200": (r) => r.status === 200 });
  errorRate.add(healthRes.status !== 200);

  const readyRes = http.get(`${BASE_URL}/ready`);
  check(readyRes, { "ready is 200": (r) => r.status === 200 });
  errorRate.add(readyRes.status !== 200);

  const productsRes = http.get(`${BASE_URL}/products`);
  check(productsRes, {
    "products status < 500": (r) => r.status < 500,
  });
  errorRate.add(productsRes.status >= 500);

  sleep(1);
}
