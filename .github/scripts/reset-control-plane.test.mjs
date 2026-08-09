import assert from "node:assert/strict";
import { mkdirSync, mkdtempSync, readFileSync, rmSync, symlinkSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import test, { afterEach } from "node:test";
import resetHandler from "../../api/reset.mjs";
import { gateIds, pathsForGate, sharedGatesFor } from "../../reset/catalog.mjs";
import { reconcilePendingResets, validResetMetadata } from "../../web/learning-path/src/lib/reset-state.js";
import { safeResetTarget, safeTarget, unexpectedResetPaths } from "./reset-gate.mjs";

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
    body: JSON.stringify({ gate: "F0", confirmed: true }),
  }));
  assert.equal(hostile.status, 403);

  const denied = await resetHandler.fetch(new Request("https://control.example/api/reset", {
    method: "POST",
    headers: { origin, authorization: "Bearer wrong", "content-type": "application/json" },
    body: JSON.stringify({ gate: "F0", confirmed: true }),
  }));
  assert.equal(denied.status, 401);
});

test("health publishes per-gate reset scope without exposing secrets", async () => {
  configure();
  const response = await resetHandler.fetch(new Request("https://control.example/api/reset", {
    headers: { origin },
  }));
  assert.equal(response.status, 200);
  const payload = await response.json();
  assert.equal(payload.available, true);
  assert.deepEqual(payload.gates.F0, { affectedFiles: 1, sharedWith: [] });
  assert.deepEqual(payload.gates.C1, {
    affectedFiles: 1,
    sharedWith: ["C0", "C2", "OPT-CONCURRENCY"],
  });
  assert.equal(JSON.stringify(payload).includes(resetKey), false);
  assert.equal(JSON.stringify(payload).includes("test-dispatch-token"), false);
  assert.equal(validResetMetadata(payload.gates, gateIds), true);
  assert.equal(validResetMetadata({ F0: payload.gates.F0 }, gateIds), false);
});

test("pending notes clear only after a newer published revision proves the gate open", () => {
  const pending = [{ gate: "F0", publishedRevision: "before-reset" }];
  const known = new Set(gateIds);
  const optional = new Set(["OPT-SYNTAX", "OPT-CONCURRENCY", "OPT-PERFORMANCE", "OPT-FP", "OPT-WORKFLOW", "OPT-COMMERCE"]);

  assert.deepEqual(reconcilePendingResets(pending, {
    revision: "before-reset",
    verified: [],
  }, known, optional), { remaining: pending, cleared: [] });

  assert.deepEqual(reconcilePendingResets(pending, {
    revision: "after-reset",
    verified: ["F0"],
  }, known, optional), { remaining: pending, cleared: [] });

  assert.deepEqual(reconcilePendingResets(pending, {
    revision: "after-reset",
    verified: [],
  }, known, optional), { remaining: [], cleared: ["F0"] });
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
    body: JSON.stringify({ gate: "K0", confirmed: true }),
  }));
  assert.equal(response.status, 202);
  assert.equal(calls.length, 2);
  assert.match(calls[1].url, /reset-gate\.yml\/dispatches$/);
  assert.deepEqual(JSON.parse(calls[1].options.body), {
    ref: "main",
    inputs: { gate: "K0", confirmation: "RESET" },
  });
  assert.deepEqual(await response.json(), {
    queued: true,
    gate: "K0",
    affectedFiles: 1,
    actionsUrl: "https://github.com/zzdpk2/golang-cloudnative-route/actions/workflows/reset-gate.yml",
    syncCommand: "git pull --ff-only",
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
    body: JSON.stringify({ gate: "E2", confirmed: true }),
  }));
  assert.equal(response.status, 503);
  assert.equal((await response.json()).error, "GitHub dispatch identity is not authorized");
});

test("gate manifest keeps resets narrow and prevents path traversal", () => {
  assert.equal(gateIds.length, 42);
  assert.deepEqual(pathsForGate("F0"), ["internal/foundation/language/data.go"]);
  assert.deepEqual(pathsForGate("E2"), ["internal/commerce/application/order_service.go"]);
  assert.deepEqual(pathsForGate("R4"), ["internal/inference/epp/request_handler.go"]);
  assert.deepEqual(sharedGatesFor("F0"), []);
  assert.deepEqual(sharedGatesFor("E6"), ["E5"]);
  assert.equal(pathsForGate("UNKNOWN"), undefined);
  for (const gate of gateIds) {
    const paths = pathsForGate(gate);
    assert.ok(paths.length > 0, `${gate} must own at least one reset file`);
    assert.equal(new Set(paths).size, paths.length, `${gate} contains duplicate reset files`);
    for (const path of paths) {
      assert.doesNotThrow(() => safeTarget(path, "C:\\repo"));
    }
  }
  assert.throws(() => safeTarget("../outside", "C:\\repo"), /unsafe reset path/);
  assert.deepEqual(unexpectedResetPaths("F0", ["internal/foundation/language/data.go"]), []);
  assert.deepEqual(unexpectedResetPaths("F0", [
    "internal/foundation/language/data.go",
    "internal/foundation/language/errors.go",
  ]), ["internal/foundation/language/errors.go"]);
});

test("reset targets reject symlink escapes", (t) => {
  const temporary = mkdtempSync(join(tmpdir(), "gate-reset-"));
  const root = join(temporary, "repo");
  const targetDirectory = join(root, "internal", "foundation", "language");
  const outside = join(temporary, "outside.go");
  mkdirSync(targetDirectory, { recursive: true });
  writeFileSync(outside, "package outside\n", "utf8");
  try {
    symlinkSync(outside, join(targetDirectory, "data.go"), "file");
  } catch (error) {
    rmSync(temporary, { recursive: true, force: true });
    if (error?.code === "EPERM") {
      t.skip("creating symlinks requires additional privileges on this Windows host");
      return;
    }
    throw error;
  }
  try {
    assert.throws(
      () => safeResetTarget("internal/foundation/language/data.go", root),
      /contains a symlink/,
    );
  } finally {
    rmSync(temporary, { recursive: true, force: true });
  }
});

test("UI catalog and workflow stay aligned with the reset allowlist", () => {
  const journeySource = readFileSync("web/learning-path/src/lib/journey.ts", "utf8");
  const uiGateIds = [...journeySource.matchAll(/\bid:\s*"([A-Z0-9-]+)"/g)]
    .map((match) => match[1]);
  assert.deepEqual(new Set(uiGateIds), new Set(gateIds));

  const workflow = readFileSync(".github/workflows/reset-gate.yml", "utf8");
  const inputLines = workflow.split(/\r?\n/).filter((line) => line.includes("inputs.gate"));
  assert.equal(inputLines.length, 3);
  assert.ok(inputLines.every((line) => line.trimStart().startsWith("RESET_GATE:")));
  assert.doesNotMatch(workflow, /git add -A/);
  assert.match(workflow, /reset-gate\.mjs "\$RESET_GATE" --stage/);
  assert.match(workflow, /git commit --allow-empty/);
});

test("invalid or unconfirmed gates never reach GitHub", async () => {
  configure();
  globalThis.fetch = async () => {
    throw new Error("GitHub must not be called");
  };
  for (const body of [
    null,
    [],
    { gate: "UNKNOWN", confirmed: true },
    { gate: "F0", confirmed: false },
    { gate: "F0", confirmed: "true" },
    { phase: 1, confirmed: true },
  ]) {
    const response = await resetHandler.fetch(new Request("https://control.example/api/reset", {
      method: "POST",
      headers: {
        origin,
        authorization: `Bearer ${resetKey}`,
        "content-type": "application/json",
      },
      body: JSON.stringify(body),
    }));
    assert.equal(response.status, 400);
  }
});
