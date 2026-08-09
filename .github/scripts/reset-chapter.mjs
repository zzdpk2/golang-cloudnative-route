import { execFileSync } from "node:child_process";
import { existsSync, mkdirSync, readFileSync, writeFileSync } from "node:fs";
import { dirname, join, resolve, sep } from "node:path";
import { fileURLToPath } from "node:url";

const starterRef = "43d1e4568b6c0adc285c28c5672731bc0eecec26";
const repositoryRoot = resolve(import.meta.dirname, "..", "..");
const expectedPathCounts = new Map([[1, 25], [2, 34], [3, 52], [4, 19]]);

function git(...args) {
  return execFileSync("git", args, {
    cwd: repositoryRoot,
    encoding: "utf8",
    maxBuffer: 10 * 1024 * 1024,
  });
}

export function isLearnerFile(path, phase) {
  const goImplementation = path.endsWith(".go") && !path.endsWith("_test.go");
  if (phase === 1) {
    if (path === "internal/foundation/testing/testcraft_test.go") return true;
    if (!path.startsWith("internal/foundation/")) return false;
    if (path.startsWith("internal/foundation/concurrency/")) return path.endsWith("/advanced.go");
    return goImplementation &&
      path !== "internal/foundation/testing/harness.go" &&
      path !== "internal/foundation/testing/subject.go";
  }
  if (phase === 2) {
    return path.startsWith("internal/commerce/") &&
      path !== "internal/commerce/testkit/subject.go" &&
      goImplementation;
  }
  if (phase === 3) {
    if (path.startsWith("deploy/inference-lab/")) return true;
    if (path.startsWith("test/acceptance/inference/")) return false;
    if (path.startsWith("test/acceptance/")) return path.endsWith(".robot") || path.endsWith(".resource");
    if (path.startsWith("internal/foundation/concurrency/")) return goImplementation && !path.endsWith("/advanced.go");
    return (path.startsWith("internal/inference/prefixcache/") ||
      path.startsWith("internal/inference/observability/")) && goImplementation;
  }
  if (path === "cmd/inference-lab/main.go") return true;
  if (path.startsWith("test/acceptance/inference/")) return path.endsWith(".robot") || path.endsWith(".resource");
  return path.startsWith("internal/inference/") &&
    !path.startsWith("internal/inference/prefixcache/") &&
    !path.startsWith("internal/inference/observability/") &&
    goImplementation;
}

export function safeTarget(path, root = repositoryRoot) {
  const target = resolve(root, path);
  if (!target.startsWith(`${root}${sep}`)) throw new Error(`unsafe reset path: ${path}`);
  return target;
}

function treePaths(ref) {
  return git("ls-tree", "-r", "--name-only", ref).split(/\r?\n/).filter(Boolean);
}

function overridePath(path) {
  if (path === "internal/foundation/language/data.go") {
    return join(repositoryRoot, ".exercise-starters", "data.go");
  }
  return join(repositoryRoot, ".exercise-starters", "files", path);
}

export function resetChapter(phase, { dryRun = false } = {}) {
  if (!expectedPathCounts.has(phase)) throw new Error("phase must be one of 1, 2, 3, or 4");

  git("cat-file", "-e", `${starterRef}^{commit}`);
  const baselinePaths = new Set(treePaths(starterRef).filter((path) => isLearnerFile(path, phase)));
  const currentPaths = treePaths("HEAD").filter((path) => isLearnerFile(path, phase));
  const paths = [...new Set([...baselinePaths, ...currentPaths])].sort();

  const expectedCount = expectedPathCounts.get(phase);
  if (paths.length !== expectedCount) {
    throw new Error(
      `phase ${phase} reset scope changed: found ${paths.length} files, expected ${expectedCount}; review the chapter manifest`,
    );
  }

  for (const path of paths) {
    const target = safeTarget(path);
    const override = overridePath(path);
    let content;
    if (existsSync(override)) {
      content = readFileSync(override);
    } else if (baselinePaths.has(path)) {
      content = execFileSync("git", ["show", `${starterRef}:${path}`], {
        cwd: repositoryRoot,
        maxBuffer: 10 * 1024 * 1024,
      });
    } else {
      throw new Error(`new exercise has no versioned starter: ${path}`);
    }

    if (!dryRun) {
      mkdirSync(dirname(target), { recursive: true });
      writeFileSync(target, content);
    }
  }

  console.log(`${dryRun ? "Validated" : "Reset"} phase ${phase}: ${paths.length} learner files`);
  return paths;
}

if (process.argv[1] && resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
  const phase = Number(process.argv[2]);
  try {
    resetChapter(phase, { dryRun: process.argv.includes("--dry-run") });
  } catch (error) {
    console.error(error instanceof Error ? error.message : error);
    process.exit(1);
  }
}
