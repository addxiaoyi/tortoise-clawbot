import "dotenv/config";
import http from "node:http";

const port = Number.parseInt(process.env.HERMES_RUNTIME_PORT ?? "3000", 10);
const host = process.env.HERMES_RUNTIME_HOST ?? "127.0.0.1";
const startedAt = new Date().toISOString();

function json(res, status, payload) {
  res.writeHead(status, {
    "Content-Type": "application/json; charset=utf-8",
    "Cache-Control": "no-store",
    "X-Content-Type-Options": "nosniff",
  });
  res.end(JSON.stringify(payload));
}

const server = http.createServer((req, res) => {
  if (req.method === "GET" && req.url === "/health") {
    return json(res, 200, {
      status: "ok",
      runtime: "hermes-agent",
      mode: "desktop-local",
      startedAt,
      pid: process.pid,
    });
  }

  if (req.method === "GET" && req.url === "/ready") {
    return json(res, 200, {
      ready: true,
      runtime: "hermes-agent",
      pid: process.pid,
    });
  }

  return json(res, 404, {
    status: "not_found",
    path: req.url,
  });
});

server.listen(port, host, () => {
  console.error(`[hermes-runtime] listening on http://${host}:${port}`);
});

function shutdown(signal) {
  console.error(`[hermes-runtime] shutting down due to ${signal}`);
  server.close(() => {
    process.exit(0);
  });
}

process.on("SIGINT", () => shutdown("SIGINT"));
process.on("SIGTERM", () => shutdown("SIGTERM"));
