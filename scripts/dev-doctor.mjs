#!/usr/bin/env node
/**
 * Verifies a minimal Hermes Agent dev layout before starting local runtimes.
 * OpenClaw checks are optional and can be enabled with --compat-openclaw.
 */
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const compatOpenClaw = process.argv.includes("--compat-openclaw");

const checks = [
  ["package.json", "根 package.json"],
  ["src/hermes/runtime/types.ts", "Hermes Runtime 类型定义"],
  ["src/hermes/runtime/skill-runner.ts", "Hermes Skill Runner"],
  ["src/hermes/runtime/tools.ts", "Hermes 工具执行层"],
  ["src/hermes/adapters/mcp.ts", "Hermes MCP 适配层"],
  ["src/mcp/tohelp-mcp-server.ts", "Hermes MCP stdio server"],
  ["src/plugins/new-core/gateway/plugin.ts", "Hermes 本地 GatewayPlugin"],
];

const openClawChecks = [
  [".openclaw-dev/openclaw.json", "开发用 OpenClaw 配置"],
  ["extensions/tohelp-openclaw/openclaw.plugin.json", "Tohelp OpenClaw 兼容扩展清单"],
  ["extensions/tohelp-openclaw/index.ts", "Tohelp OpenClaw 兼容扩展入口"],
  ["openclaw-main/package.json", "OpenClaw 源码目录 openclaw-main（可选兼容）"],
  ["openclaw-main/scripts/run-node.mjs", "OpenClaw 启动脚本 run-node.mjs（可选兼容）"],
];

let failed = false;
for (const [rel, hint] of checks) {
  const abs = path.join(root, rel);
  if (fs.existsSync(abs)) {
    console.log("ok", rel);
  } else {
    console.error("missing", rel);
    console.error("      →", hint);
    failed = true;
  }
}

if (compatOpenClaw) {
  for (const [rel, hint] of openClawChecks) {
    const abs = path.join(root, rel);
    if (fs.existsSync(abs)) {
      console.log("ok compat", rel);
    } else {
      console.error("missing compat", rel);
      console.error("      →", hint);
      failed = true;
    }
  }
}

const major = Number.parseInt(process.versions.node.split(".")[0] ?? "0", 10);
if (major < 20) {
  console.error("missing node>=20");
  console.error("      → Hermes Agent requires Node.js 20+; Node 22+ is recommended.");
  failed = true;
} else {
  console.log("ok node", process.version);
}

if (failed) {
  console.error("\n修复后重试: npm run doctor");
  process.exit(1);
}

if (!compatOpenClaw) {
  console.log("tip openclaw-main: optional compatibility layer; run npm run doctor:openclaw to verify it.");
} else {
  const openclawRoot = path.join(root, "openclaw-main");
  const openclawGit = path.join(openclawRoot, ".git");
  if (fs.existsSync(openclawRoot)) {
    if (fs.existsSync(openclawGit)) {
      try {
        const st = fs.lstatSync(openclawGit);
        if (st.isFile()) {
          console.log(
            "tip openclaw-main: git submodule (use docs/setup-openclaw-submodule.md to update upstream)",
          );
        } else {
          console.log(
            "tip openclaw-main: plain clone/dir — optional submodule: docs/setup-openclaw-submodule.md",
          );
        }
      } catch {
        console.log(
          "tip openclaw-main: see docs/setup-openclaw-submodule.md for submodule workflow",
        );
      }
    } else {
      console.log(
        "tip openclaw-main: no .git (vendor copy?) — clone OpenClaw or use submodule: docs/setup-openclaw-submodule.md",
      );
    }
  }
}

console.log("\nHermes Agent dev layout looks good.");
