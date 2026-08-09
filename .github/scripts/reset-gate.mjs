import { execFileSync } from "node:child_process";
import { existsSync, lstatSync, mkdirSync, readFileSync, writeFileSync } from "node:fs";
import { dirname, join, resolve, sep } from "node:path";
import { fileURLToPath } from "node:url";
import { gateIds, pathsForGate } from "../../reset/catalog.mjs";

const starterRef = "43d1e4568b6c0adc285c28c5672731bc0eecec26";
const repositoryRoot = resolve(import.meta.dirname, "..", "..");

function git(...args) {
  return execFileSync("git", args, {
    cwd: repositoryRoot,
    encoding: "utf8",
    maxBuffer: 10 * 1024 * 1024,
  });
}

export function safeTarget(path, root = repositoryRoot) {
  const resolvedRoot = resolve(root);
  const target = resolve(resolvedRoot, path);
  if (!target.startsWith(`${resolvedRoot}${sep}`)) throw new Error(`unsafe reset path: ${path}`);
  return target;
}

export function safeResetTarget(path, root = repositoryRoot) {
  const resolvedRoot = resolve(root);
  const target = safeTarget(path, resolvedRoot);
  let current = resolvedRoot;
  for (const segment of target.slice(resolvedRoot.length + 1).split(sep)) {
    current = join(current, segment);
    const stat = lstatSync(current, { throwIfNoEntry: false });
    if (stat?.isSymbolicLink()) throw new Error(`reset path contains a symlink: ${path}`);
  }
  return target;
}

function overridePath(path) {
  if (path === "internal/foundation/language/data.go") {
    return join(repositoryRoot, ".exercise-starters", "data.go");
  }
  return join(repositoryRoot, ".exercise-starters", "files", path);
}

function starterContent(path) {
  const override = overridePath(path);
  if (existsSync(override)) return readFileSync(override);
  try {
    return execFileSync("git", ["show", `${starterRef}:${path}`], {
      cwd: repositoryRoot,
      maxBuffer: 10 * 1024 * 1024,
    });
  } catch {
    throw new Error(`gate file has no versioned starter: ${path}`);
  }
}

export function resetGate(gateId, { dryRun = false } = {}) {
  const normalized = String(gateId ?? "").toUpperCase();
  const configuredPaths = pathsForGate(normalized);
  if (!configuredPaths) {
    throw new Error(`gate must be one of: ${gateIds.join(", ")}`);
  }

  git("cat-file", "-e", `${starterRef}^{commit}`);
  const paths = [...new Set(configuredPaths)].sort();
  if (paths.length !== configuredPaths.length || paths.length === 0) {
    throw new Error(`gate ${normalized} reset manifest must contain unique files`);
  }

  for (const path of paths) {
    const target = safeResetTarget(path);
    const content = starterContent(path);
    if (!dryRun) {
      mkdirSync(dirname(target), { recursive: true });
      writeFileSync(target, content);
    }
  }

  console.log(`${dryRun ? "Validated" : "Reset"} gate ${normalized}: ${paths.length} learner files`);
  return paths;
}

function changedRepositoryPaths() {
  const commands = [
    ["diff", "--name-only", "--no-renames", "-z", "--"],
    ["diff", "--cached", "--name-only", "--no-renames", "-z", "--"],
    ["ls-files", "--others", "--exclude-standard", "-z"],
  ];
  return [...new Set(commands.flatMap((args) => git(...args).split("\0").filter(Boolean)))].sort();
}

export function unexpectedResetPaths(gateId, changedPaths) {
  const allowed = pathsForGate(gateId);
  if (!allowed) throw new Error(`unknown gate: ${gateId}`);
  const allowlist = new Set(allowed);
  return [...new Set(changedPaths)].filter((path) => !allowlist.has(path)).sort();
}

export function stageGateChanges(gateId) {
  const normalized = String(gateId ?? "").toUpperCase();
  const changedPaths = changedRepositoryPaths();
  const unexpected = unexpectedResetPaths(normalized, changedPaths);
  if (unexpected.length > 0) {
    throw new Error(`gate ${normalized} changed files outside its allowlist: ${unexpected.join(", ")}`);
  }
  if (changedPaths.length > 0) git("add", "--", ...changedPaths);
  console.log(`Staged gate ${normalized}: ${changedPaths.length} changed learner files`);
  return changedPaths;
}

if (process.argv[1] && resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
  try {
    if (process.argv.includes("--all-dry-run")) {
      for (const gateId of gateIds) resetGate(gateId, { dryRun: true });
    } else if (process.argv.includes("--stage")) {
      stageGateChanges(process.argv[2]);
    } else {
      resetGate(process.argv[2], { dryRun: process.argv.includes("--dry-run") });
    }
  } catch (error) {
    console.error(error instanceof Error ? error.message : error);
    process.exit(1);
  }
}
