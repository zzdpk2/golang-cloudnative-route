import { readFile, realpath, stat } from "node:fs/promises";
import path from "node:path";
import { allGates } from "$lib/journey";
import type { RequestHandler } from "./$types";

const repositoryRoot = path.resolve(process.cwd(), "..", "..");
const maxSourceBytes = 1_000_000;
const allowedSources = new Set([
  ...allGates.map((gate) => gate.source),
  "docs/LEARNING_PATH.md",
  "docs/go100/README.md",
]);

export const GET: RequestHandler = async ({ params }) => {
  const relativeSource = params.path;
  if (!allowedSources.has(relativeSource)) {
    return new Response("Source is not part of the learning journey.", { status: 403 });
  }
  const requested = path.resolve(repositoryRoot, ...relativeSource.split("/"));

  try {
    const [realRepositoryRoot, realRequested] = await Promise.all([
      realpath(repositoryRoot),
      realpath(requested),
    ]);
    const insideRepository =
      realRequested === realRepositoryRoot ||
      realRequested.startsWith(realRepositoryRoot + path.sep);
    if (!insideRepository) {
      return new Response("Source path is outside the repository.", { status: 403 });
    }
    const metadata = await stat(realRequested);
    if (!metadata.isFile() || metadata.size > maxSourceBytes) {
      return new Response("Source is not a readable text file.", { status: 400 });
    }
    const content = await readFile(realRequested, "utf8");
    return new Response(content, {
      headers: {
        "Content-Type": "text/plain; charset=utf-8",
        "Cache-Control": "no-store, max-age=0",
        "X-Content-Type-Options": "nosniff",
      },
    });
  } catch {
    return new Response("Source file not found.", { status: 404 });
  }
};
