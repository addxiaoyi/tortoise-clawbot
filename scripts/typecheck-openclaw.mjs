#!/usr/bin/env node
/**
 * Optional: typecheck the vendored `openclaw-main` subtree (upstream OpenClaw).
 * Skips with exit 0 when `openclaw-main/node_modules` is missing.
 * Install: `cd openclaw-main && pnpm install`, then from repo root: `npm run typecheck:openclaw`
 */
import { existsSync } from "node:fs";
import { spawnSync } from "node:child_process";
import path from "node:path";
import { fileURLToPath } from "node:url";

const root = path.join(path.dirname(fileURLToPath(import.meta.url)), "..");
const oc = path.join(root, "openclaw-main");
const nm = path.join(oc, "node_modules");

if (!existsSync(nm)) {
  console.log(
    "[typecheck:openclaw] SKIP — openclaw-main/node_modules missing (cd openclaw-main && pnpm install)",
  );
  process.exit(0);
}

const r = spawnSync("pnpm", ["exec", "tsc", "--noEmit", "-p", "tsconfig.json"], {
  cwd: oc,
  stdio: "inherit",
  shell: true,
  env: process.env,
});

if (r.status !== 0) {
  console.log(
    "[typecheck:openclaw] FAIL — bare tsc may not match upstream workflow; see openclaw-main/package.json `check` (pnpm tsgo, oxlint, …)",
  );
}

process.exit(r.status ?? 1);
