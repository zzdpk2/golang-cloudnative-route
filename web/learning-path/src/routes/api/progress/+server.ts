import { readFile, stat } from "node:fs/promises";
import path from "node:path";
import { json } from "@sveltejs/kit";
import { gates, optionalGates } from "$lib/journey";
import type { RequestHandler } from "./$types";

const progressPath = path.resolve(process.cwd(), "..", "..", "docs", "learning-progress.js");

function parseAssignment(source: string, name: string): unknown {
  const line = source
    .split(/\r?\n/)
    .find((candidate) => candidate.startsWith(`window.${name} = `));
  if (!line) throw new Error(`missing ${name}`);
  const encoded = line.slice(`window.${name} = `.length).replace(/;\s*$/, "");
  return JSON.parse(encoded) as unknown;
}

function parseOptionalAssignment(source: string, name: string): unknown {
  return source.includes(`window.${name} = `) ? parseAssignment(source, name) : [];
}

export const GET: RequestHandler = async () => {
  try {
    const [source, metadata] = await Promise.all([
      readFile(progressPath, "utf8"),
      stat(progressPath),
    ]);
    const rawVerified = parseAssignment(source, "__LEARNING_VERIFIED_GATES__");
    const rawOptionalVerified = parseOptionalAssignment(
      source,
      "__LEARNING_OPTIONAL_VERIFIED_GATES__",
    );
    const rawFocus = parseAssignment(source, "__GO_MISTAKE_FOCUS__");
    if (!Array.isArray(rawVerified) || rawVerified.some((id) => typeof id !== "string")) {
      throw new Error("invalid verified gate data");
    }
    if (
      !Array.isArray(rawOptionalVerified) ||
      rawOptionalVerified.some((id) => typeof id !== "string")
    ) {
      throw new Error("invalid optional verified gate data");
    }
    if (typeof rawFocus !== "object" || rawFocus === null || Array.isArray(rawFocus)) {
      throw new Error("invalid review focus data");
    }

    const knownIds = new Set(gates.map((gate) => gate.id));
    const generatedVerified = new Set(rawVerified);
    if (
      generatedVerified.size !== rawVerified.length ||
      rawVerified.some((id) => !knownIds.has(id))
    ) {
      throw new Error("invalid verified gate IDs");
    }
    const verified = gates
      .filter((gate) => generatedVerified.has(gate.id))
      .map((gate) => gate.id);

    const knownOptionalIds = new Set(optionalGates.map((gate) => gate.id));
    const generatedOptionalVerified = new Set(rawOptionalVerified);
    if (
      generatedOptionalVerified.size !== rawOptionalVerified.length ||
      rawOptionalVerified.some((id) => !knownOptionalIds.has(id))
    ) {
      throw new Error("invalid optional verified gate IDs");
    }
    const optionalVerified = optionalGates
      .filter((gate) => generatedOptionalVerified.has(gate.id))
      .map((gate) => gate.id);

    const generatedFocus = rawFocus as Record<string, unknown>;
    if (
      Object.keys(generatedFocus).length !== gates.length ||
      Object.keys(generatedFocus).some((id) => !knownIds.has(id))
    ) {
      throw new Error("invalid review focus IDs");
    }
    const focus: Record<string, string> = {};
    for (const gate of gates) {
      const prompt = generatedFocus[gate.id];
      if (typeof prompt !== "string" || prompt.trim() === "") {
        throw new Error(`invalid review focus for ${gate.id}`);
      }
      focus[gate.id] = prompt;
    }

    return json(
      { verified, optionalVerified, focus, updatedAt: metadata.mtime.toISOString() },
      { headers: { "Cache-Control": "no-store, max-age=0" } },
    );
  } catch {
    return json(
      {
        error: "Verified progress is temporarily unavailable.",
        verified: [],
        optionalVerified: [],
        focus: {},
      },
      { status: 503, headers: { "Cache-Control": "no-store, max-age=0" } },
    );
  }
};
