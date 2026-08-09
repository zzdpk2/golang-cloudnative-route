import { timingSafeEqual } from "node:crypto";
import { gateIds, pathsForGate, sharedGatesFor } from "../reset/catalog.mjs";

function allowedOrigins() {
  return (process.env.ALLOWED_ORIGINS ?? "")
    .split(",")
    .map((origin) => origin.trim())
    .filter(Boolean);
}

function responseHeaders(origin) {
  const headers = new Headers({
    "cache-control": "no-store",
    "content-type": "application/json; charset=utf-8",
    "x-content-type-options": "nosniff",
  });
  if (origin && allowedOrigins().includes(origin)) {
    headers.set("access-control-allow-origin", origin);
    headers.set("access-control-allow-headers", "authorization, content-type");
    headers.set("access-control-allow-methods", "GET, POST, OPTIONS");
    headers.set("vary", "Origin");
  }
  return headers;
}

function json(body, status, origin) {
  return new Response(JSON.stringify(body), {
    status,
    headers: responseHeaders(origin),
  });
}

function secretMatches(candidate, expected) {
  const left = Buffer.from(candidate ?? "", "utf8");
  const right = Buffer.from(expected ?? "", "utf8");
  return left.length === right.length && left.length >= 32 && timingSafeEqual(left, right);
}

function configuration() {
  const repository = process.env.GITHUB_REPOSITORY ?? "zzdpk2/golang-cloudnative-route";
  const [owner, name, ...extra] = repository.split("/");
  const resetKey = process.env.RESET_KEY ?? "";
  const dispatchToken = process.env.GITHUB_DISPATCH_TOKEN ?? "";
  const ready = Boolean(owner && name && extra.length === 0 && resetKey.length >= 32 && dispatchToken);
  return { ready, repository, owner, resetKey, dispatchToken };
}

function githubHeaders(token) {
  return {
    accept: "application/vnd.github+json",
    authorization: `Bearer ${token}`,
    "x-github-api-version": "2022-11-28",
    "user-agent": "route-learning-reset-control-plane",
  };
}

export default {
  async fetch(request) {
    const origin = request.headers.get("origin") ?? "";
    const originAllowed = allowedOrigins().includes(origin);

    if (request.method === "OPTIONS") {
      return new Response(null, { status: originAllowed ? 204 : 403, headers: responseHeaders(origin) });
    }
    if (!originAllowed) return json({ error: "origin is not allowed" }, 403, origin);
    const config = configuration();
    if (request.method === "GET") {
      return config.ready
        ? json({
            available: true,
            gates: Object.fromEntries(gateIds.map((gate) => [gate, {
              affectedFiles: pathsForGate(gate).length,
              sharedWith: sharedGatesFor(gate),
            }])),
          }, 200, origin)
        : json({ available: false }, 503, origin);
    }
    if (request.method !== "POST") return json({ error: "method is not allowed" }, 405, origin);
    if (!config.ready) return json({ error: "dispatch is not configured" }, 503, origin);

    const authorization = request.headers.get("authorization") ?? "";
    const suppliedKey = authorization.startsWith("Bearer ") ? authorization.slice(7) : "";
    if (!secretMatches(suppliedKey, config.resetKey)) {
      return json({ error: "invalid reset key" }, 401, origin);
    }

    let body;
    try {
      body = await request.json();
    } catch {
      return json({ error: "invalid JSON body" }, 400, origin);
    }
    if (typeof body !== "object" || body === null || Array.isArray(body)) {
      return json({ error: "a valid, confirmed gate is required" }, 400, origin);
    }
    const gate = String(body.gate ?? "").toUpperCase();
    if (body.confirmed !== true || !gateIds.includes(gate)) {
      return json({ error: "a valid, confirmed gate is required" }, 400, origin);
    }

    const identityResponse = await fetch("https://api.github.com/user", {
      headers: githubHeaders(config.dispatchToken),
    });
    if (!identityResponse.ok) {
      console.error("GitHub token identity check failed", identityResponse.status);
      return json({ error: "GitHub dispatch identity is invalid" }, 502, origin);
    }
    const identity = await identityResponse.json();
    if (String(identity.login).toLowerCase() !== config.owner.toLowerCase()) {
      console.error("GitHub dispatch token is not owned by the repository owner");
      return json({ error: "GitHub dispatch identity is not authorized" }, 503, origin);
    }

    const githubResponse = await fetch(
      `https://api.github.com/repos/${config.repository}/actions/workflows/reset-gate.yml/dispatches`,
      {
        method: "POST",
        headers: {
          ...githubHeaders(config.dispatchToken),
          "content-type": "application/json",
        },
        body: JSON.stringify({
          ref: "main",
          inputs: { gate, confirmation: "RESET" },
        }),
      },
    );

    if (!githubResponse.ok) {
      console.error("GitHub workflow dispatch failed", githubResponse.status);
      return json({ error: `GitHub dispatch failed (${githubResponse.status})` }, 502, origin);
    }

    return json({
      queued: true,
      gate,
      affectedFiles: pathsForGate(gate).length,
      actionsUrl: `https://github.com/${config.repository}/actions/workflows/reset-gate.yml`,
      syncCommand: "git pull --ff-only",
    }, 202, origin);
  },
};
