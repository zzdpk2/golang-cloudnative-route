import { mkdir, readFile, writeFile } from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";

const scriptDirectory = path.dirname(fileURLToPath(import.meta.url));
const frontendRoot = path.resolve(scriptDirectory, "..");
const repositoryRoot = path.resolve(frontendRoot, "..", "..");

const sourcePath = path.join(
  repositoryRoot,
  "docs",
  "learning-progress.js",
);

const outputPath = path.join(
  frontendRoot,
  "static",
  "progress.json",
);

function readAssignment(source, name, fallback) {
  const prefix = `window.${name} = `;

  const line = source
    .split(/\r?\n/)
    .find((candidate) => candidate.startsWith(prefix));

  if (!line) {
    if (fallback !== undefined) {
      return fallback;
    }

    throw new Error(`missing ${name}`);
  }

  const json = line
    .slice(prefix.length)
    .replace(/;\s*$/, "");

  return JSON.parse(json);
}

const source = await readFile(sourcePath, "utf8");

const progress = {
  verified: readAssignment(
    source,
    "__LEARNING_VERIFIED_GATES__",
  ),
  optionalVerified: readAssignment(
    source,
    "__LEARNING_OPTIONAL_VERIFIED_GATES__",
    [],
  ),
  focus: readAssignment(
    source,
    "__GO_MISTAKE_FOCUS__",
  ),
  updatedAt: new Date().toISOString(),
  revision: process.env.GITHUB_SHA ?? "local",
};

await mkdir(path.dirname(outputPath), {
  recursive: true,
});

await writeFile(
  outputPath,
  `${JSON.stringify(progress, null, 2)}\n`,
  "utf8",
);

console.log(
  `Generated ${path.relative(repositoryRoot, outputPath)}`,
);