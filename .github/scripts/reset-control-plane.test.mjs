import assert from "node:assert/strict";
import test, { afterEach } from "node:test";
import resetHandler from "../../api/reset.mjs";
import { isLearnerFile, safeTarget } from "./reset-chapter.mjs";

const originalFetch = globalThis.fetch;
const origin = "https://zzdpk2.github.io";
const resetKey = "0123456789abcdef0123456789abcdef";

afterEach(() => {
  globalThis.fetch = originalFetch;
  delete process.env.ALLOWED_ORIGINS;
  delete process.env.RESET_KEY;
  delete process.env.GITHUB_DISPATCH_TOKEN;
  delete process.env.GITHUB_REPOSITORY;
});

function configure() {
  process.env.ALLOWED_ORIGINS = origin;
  process.env.RESET_KEY = resetKey;
  process.env.GITHUB_DISPATCH_TOKEN = "test-dispatch-token";
  process.env.GITHUB_REPOSITORY = "zzdpk2/golang-cloudnative-route";
}

test("health stays unavailable until every server secret is configured", async () => {
  process.env.ALLOWED_ORIGINS = origin;
  const response = await resetHandler.fetch(new Request("https://control.example/api/reset", {
    headers: { origin },
  }));
  assert.equal(response.status, 503);
  assert.deepEqual(await response.json(), { available: false });
});

test("origin and Reset Key guard the dispatch endpoint", async () => {
  configure();
  const hostile = await resetHandler.fetch(new Request("https://control.example/api/reset", {
    method: "POST",
    headers: { origin: "https://evil.example", "content-type": "application/json" },
    body: JSON.stringify({ phase: 1, confirmed: true }),
  }));
  assert.equal(hostile.status, 403);

  const denied = await resetHandler.fetch(new Request("https://control.example/api/reset", {
    method: "POST",
    headers: { origin, authorization: "Bearer wrong", "content-type": "application/json" },
    body: JSON.stringify({ phase: 1, confirmed: true }),
  }));
  assert.equal(denied.status, 401);
});

test("only an owner-owned token dispatches the allowlisted workflow", async () => {
  configure();
  const calls = [];
  globalThis.fetch = async (url, options = {}) => {
    calls.push({ url: String(url), options });
    if (String(url).endsWith("/user")) {
      return Response.json({ login: "zzdpk2" });
    }
    return new Response(null, { status: 204 });
  };

  const response = await resetHandler.fetch(new Request("https://control.example/api/reset", {
    method: "POST",
    headers: {
      origin,
      authorization: `Bearer ${resetKey}`,
      "content-type": "application/json",
    },
    body: JSON.stringify({ phase: 3, confirmed: true }),
  }));
  assert.equal(response.status, 202);
  assert.equal(calls.length, 2);
  assert.match(calls[1].url, /reset-chapter\.yml\/dispatches$/);
  assert.deepEqual(JSON.parse(calls[1].options.body), {
    ref: "main",
    inputs: { phase: "3", confirmation: "RESET" },
  });
});

test("a collaborator-owned token cannot produce a false queued response", async () => {
  configure();
  globalThis.fetch = async () => Response.json({ login: "someone-else" });
  const response = await resetHandler.fetch(new Request("https://control.example/api/reset", {
    method: "POST",
    headers: {
      origin,
      authorization: `Bearer ${resetKey}`,
      "content-type": "application/json",
    },
    body: JSON.stringify({ phase: 2, confirmed: true }),
  }));
  assert.equal(response.status, 503);
  assert.equal((await response.json()).error, "GitHub dispatch identity is not authorized");
});

test("phase ownership excludes contracts and prevents path traversal", () => {
  assert.equal(isLearnerFile("internal/foundation/language/data.go", 1), true);
  assert.equal(isLearnerFile("internal/foundation/language/data_test.go", 1), false);
  assert.equal(isLearnerFile("internal/foundation/concurrency/advanced.go", 1), true);
  assert.equal(isLearnerFile("internal/foundation/concurrency/concurrency.go", 3), true);
  assert.equal(isLearnerFile("internal/inference/prefixcache/prefix.go", 3), true);
  assert.equal(isLearnerFile("internal/inference/routing/scheduler.go", 4), true);
  assert.equal(isLearnerFile("test/e2e/milestones_test.go", 2), false);
  assert.throws(() => safeTarget("../outside", "C:\\repo"), /unsafe reset path/);
});
