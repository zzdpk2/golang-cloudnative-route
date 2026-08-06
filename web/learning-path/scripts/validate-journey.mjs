import { access, readFile } from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";

const appRoot = path.dirname(path.dirname(fileURLToPath(import.meta.url)));
const repositoryRoot = path.resolve(appRoot, "..", "..");

const [catalogSource, runnerSource, progressSource, markdownSource, htmlSource] = await Promise.all([
  readFile(path.join(appRoot, "src", "lib", "journey.ts"), "utf8"),
  readFile(path.join(repositoryRoot, "cmd", "exercise", "journey.go"), "utf8"),
  readFile(path.join(repositoryRoot, "docs", "learning-progress.js"), "utf8"),
  readFile(path.join(repositoryRoot, "docs", "LEARNING_PATH.md"), "utf8"),
  readFile(path.join(repositoryRoot, "docs", "learning-path.html"), "utf8"),
]);

const uiIds = [...catalogSource.matchAll(/\bid:\s*"([A-Z]\d+)"/g)].map(
  (match) => match[1],
);
const orderBlock = runnerSource.match(
  /var learningJourney = mustOrderChecks\(checkCatalog, \[\]string\{([\s\S]*?)\}\)/,
);
if (!orderBlock) throw new Error("cannot find the Go learning journey order");
const runnerIds = [...orderBlock[1].matchAll(/"([A-Z]\d+)"/g)].map(
  (match) => match[1],
);

if (JSON.stringify(uiIds) !== JSON.stringify(runnerIds)) {
  throw new Error(
    `SvelteKit gates do not match the Go runner order:\nUI     ${uiIds.join(" ")}\nRunner ${runnerIds.join(" ")}`,
  );
}
if (runnerIds.length !== 36 || uiIds.length !== 36) {
  throw new Error(`the required journey must contain exactly 36 gates, found ${runnerIds.length}`);
}
if (new Set(uiIds).size !== uiIds.length) {
  throw new Error("the SvelteKit gate catalog contains duplicate IDs");
}

const optionalCatalogBlock = catalogSource.match(
  /export const optionalGates: Gate\[\] = \[([\s\S]*?)\n\];/,
);
const optionalRunnerBlock = runnerSource.match(
  /var starredOptional = \[\]optionalGroup\{([\s\S]*?)\n\}/,
);
if (!optionalCatalogBlock || !optionalRunnerBlock) {
  throw new Error("cannot find the optional extension catalogs");
}
const optionalUIIds = [...optionalCatalogBlock[1].matchAll(/\bid:\s*"(OPT-[A-Z]+)"/g)].map(
  (match) => match[1],
);
const optionalRunnerIds = [...optionalRunnerBlock[1].matchAll(/\bid:\s*"(OPT-[A-Z]+)"/g)].map(
  (match) => match[1],
);
if (
  optionalUIIds.length !== 6 ||
  JSON.stringify(optionalUIIds) !== JSON.stringify(optionalRunnerIds)
) {
  throw new Error("SvelteKit optional branches do not match the Go runner order");
}
const optionalUIPhases = [
  ...optionalCatalogBlock[1].matchAll(/\bid:\s*"OPT-[A-Z]+",\s*\n\s*phase:\s*([1-4])/g),
].map((match) => Number(match[1]));
const optionalRunnerPhases = [
  ...optionalRunnerBlock[1].matchAll(/\bid:\s*"OPT-[A-Z]+",\s*\n\s*phase:\s*([1-4])/g),
].map((match) => Number(match[1]));
if (JSON.stringify(optionalUIPhases) !== JSON.stringify(optionalRunnerPhases)) {
  throw new Error("optional phase placement drifted between SvelteKit and the Go runner");
}
const optionalFlags = [...optionalCatalogBlock[1].matchAll(/\boptional:\s*true/g)];
if (optionalFlags.length !== optionalUIIds.length) {
  throw new Error("every optional SvelteKit branch must declare optional: true");
}
for (const field of ["title", "command", "source", "tag", "focus"]) {
  const expression = new RegExp(`\\b${field}:\\s*"([^"]+)"`, "g");
  const uiValues = [...optionalCatalogBlock[1].matchAll(expression)].map((match) => match[1]);
  const runnerValues = [...optionalRunnerBlock[1].matchAll(expression)].map((match) => match[1]);
  if (JSON.stringify(uiValues) !== JSON.stringify(runnerValues)) {
    throw new Error(`optional ${field} metadata drifted between SvelteKit and the Go runner`);
  }
}

const markdownIds = [...markdownSource.matchAll(/^\| ([A-Z]\d+) \|/gm)].map(
  (match) => match[1],
);
const htmlIds = [...htmlSource.matchAll(/data-gate="([A-Z]\d+)"/g)].map(
  (match) => match[1],
);
for (const [name, ids] of [
  ["Markdown tracker", markdownIds],
  ["HTML tracker", htmlIds],
]) {
  if (JSON.stringify(ids) !== JSON.stringify(runnerIds)) {
    throw new Error(`${name} does not match the Go runner order`);
  }
}

const markdownOptionalIds = [
  ...markdownSource.matchAll(/^\| `(OPT-[A-Z]+)`/gm),
].map((match) => match[1]);
const htmlOptionalIds = [
  ...htmlSource.matchAll(/data-optional="(OPT-[A-Z]+)"/g),
].map((match) => match[1]);
for (const [name, ids] of [
  ["Markdown optional tracker", markdownOptionalIds],
  ["HTML optional tracker", htmlOptionalIds],
]) {
  if (JSON.stringify(ids) !== JSON.stringify(optionalRunnerIds)) {
    throw new Error(`${name} does not match the Go optional runner order`);
  }
}

const expectedPhase = (id) => {
  if (id.startsWith("F")) return 1;
  if (id.startsWith("E")) return 2;
  if (id.startsWith("R")) return 4;
  return 3;
};
const uiPhases = [...catalogSource.matchAll(/\bid:\s*"([A-Z]\d+)",[\s\S]*?\bphase:\s*([1-4]),/g)];
if (uiPhases.length !== uiIds.length) {
  throw new Error("cannot validate every SvelteKit gate phase");
}
for (const [, id, phase] of uiPhases) {
  if (Number(phase) !== expectedPhase(id)) {
    throw new Error(`gate ${id} is assigned to the wrong phase`);
  }
}

const sourcePaths = [
  ...catalogSource.matchAll(/\bsource:\s*"([^"]+)"/g),
].map((match) => match[1]);
await Promise.all(
  sourcePaths.map((source) => access(path.resolve(repositoryRoot, source))),
);

const focusLine = progressSource
  .split(/\r?\n/)
  .find((line) => line.startsWith("window.__GO_MISTAKE_FOCUS__ = "));
if (!focusLine) throw new Error("generated progress has no Go Mistakes focus");
const focus = JSON.parse(
  focusLine
    .slice("window.__GO_MISTAKE_FOCUS__ = ".length)
    .replace(/;\s*$/, ""),
);
for (const id of uiIds) {
  if (typeof focus[id] !== "string" || focus[id].trim() === "") {
    throw new Error(`gate ${id} has no generated Go Mistakes focus`);
  }
}

console.log(
  `Validated ${uiIds.length} required gates and ${optionalUIIds.length} optional branches across the runner, SvelteKit, Markdown, HTML, sources, phases, and review prompts.`,
);
