#!/usr/bin/env node
/**
 * Secret hygiene checks for multi-project workspace.
 * - Detect tracked sensitive files.
 * - Check local .env values for obvious placeholders.
 * - Compare *.env.example with sibling .env for missing keys.
 *
 * Usage:
 *   node scripts/doctor-secrets.mjs
 *   node scripts/doctor-secrets.mjs --strict   # exit 1 if any warning
 *   node scripts/doctor-secrets.mjs --strict --ci
 */
import fs from "node:fs";
import path from "node:path";
import { execSync } from "node:child_process";
import { fileURLToPath } from "node:url";

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const strict = process.argv.includes("--strict");
const ciMode = process.argv.includes("--ci");
let warnCount = 0;

const EXAMPLE_FILES = [
  ".env.example",
  "cloud/env.example",
  "cloud/web/.env.example",
  "tortoise/.env.example",
];

const PLACEHOLDER_PATTERNS = [
  /placeholder/i,
  /change[-_ ]?me/i,
  /^your[_-]/i,
  /^replace[-_]/i,
  /^xxx+$/i,
  /^todo$/i,
];

const TRACKED_SENSITIVE_PATTERNS = [
  /(^|\/)\.env(\..+)?$/i,
  /(^|\/)credentials\.json$/i,
  /(^|\/)secrets?\.(json|ya?ml)$/i,
  /\.pem$/i,
  /\.key$/i,
  /\.p12$/i,
  /\.pfx$/i,
];

const TRACKED_ALLOWLIST = [
  /(^|\/)\.env\.example$/i,
  /(^|\/)env\.example$/i,
  /(^|\/)openclaw-main\/.*$/i,
];

function warning(msg) {
  warnCount += 1;
  console.warn(`warn ${msg}`);
}

function info(msg) {
  console.log(`ok ${msg}`);
}

function parseEnvFile(absPath) {
  const result = new Map();
  const text = fs.readFileSync(absPath, "utf8");
  const lines = text.split(/\r?\n/);
  for (const line of lines) {
    const trimmed = line.trim();
    if (!trimmed || trimmed.startsWith("#")) continue;
    const idx = trimmed.indexOf("=");
    if (idx <= 0) continue;
    const key = trimmed.slice(0, idx).trim();
    const value = trimmed.slice(idx + 1).trim();
    if (!key) continue;
    result.set(key, value);
  }
  return result;
}

function isPlaceholderValue(value) {
  if (!value) return false;
  return PLACEHOLDER_PATTERNS.some((re) => re.test(value));
}

function getTrackedFiles() {
  const out = execSync("git ls-files", { cwd: root, encoding: "utf8" });
  return out
    .split(/\r?\n/)
    .map((x) => x.trim())
    .filter(Boolean);
}

function checkTrackedSensitiveFiles() {
  const tracked = getTrackedFiles();
  const hits = tracked.filter((rel) => {
    const matched = TRACKED_SENSITIVE_PATTERNS.some((re) => re.test(rel));
    if (!matched) return false;
    return !TRACKED_ALLOWLIST.some((re) => re.test(rel));
  });
  if (hits.length === 0) {
    info("git tracked files: no obvious secret files");
    return;
  }
  for (const rel of hits) {
    warning(`tracked sensitive file detected: ${rel}`);
  }
}

function checkExampleCompletenessAndPlaceholders(exampleRel) {
  const exampleAbs = path.join(root, exampleRel);
  if (!fs.existsSync(exampleAbs)) {
    warning(`missing example file: ${exampleRel}`);
    return;
  }
  info(`found example: ${exampleRel}`);

  const localRel = path.join(path.dirname(exampleRel), ".env").replaceAll("\\", "/");
  const localAbs = path.join(root, localRel);
  if (!fs.existsSync(localAbs)) {
    if (ciMode) {
      info(`ci mode: local env skipped: ${localRel}`);
    } else {
      warning(`local env not found: ${localRel} (copy from ${exampleRel})`);
    }
    return;
  }

  const exampleKV = parseEnvFile(exampleAbs);
  const localKV = parseEnvFile(localAbs);
  const missingKeys = [];
  for (const key of exampleKV.keys()) {
    if (!localKV.has(key)) missingKeys.push(key);
  }
  if (missingKeys.length > 0) {
    warning(`${localRel} missing ${missingKeys.length} keys from ${exampleRel}`);
  } else {
    info(`${localRel} key coverage matches ${exampleRel}`);
  }

  let placeholderCount = 0;
  for (const [key, value] of localKV.entries()) {
    if (!/_KEY|TOKEN|SECRET|PASSWORD|DSN|DATABASE_URL/i.test(key)) continue;
    if (isPlaceholderValue(value)) {
      placeholderCount += 1;
      warning(`${localRel} has placeholder-like value: ${key}`);
    }
  }
  if (placeholderCount === 0) {
    info(`${localRel} has no obvious placeholder values for secret-like keys`);
  }
}

checkTrackedSensitiveFiles();
for (const rel of EXAMPLE_FILES) {
  checkExampleCompletenessAndPlaceholders(rel);
}

if (warnCount > 0) {
  console.warn(`\nSecret doctor finished with ${warnCount} warning(s).`);
  if (strict) process.exit(1);
} else {
  console.log("\nSecret doctor finished clean.");
}
