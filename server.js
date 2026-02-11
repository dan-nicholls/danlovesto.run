import { createServer } from "node:http";
import { readFile, stat } from "node:fs/promises";
import { extname, join } from "node:path";
import { fileURLToPath } from "node:url";

const __dirname = fileURLToPath(new URL(".", import.meta.url));

const MIME_TYPES = {
  ".html": "text/html",
  ".js": "text/javascript",
  ".css": "text/css",
  ".png": "image/png",
  ".svg": "image/svg+xml",
  ".ico": "image/x-icon",
  ".json": "application/json",
};

function parseEnv(content) {
  const env = {};
  const lines = content.split("\n");
  for (const line of lines) {
    const trimmed = line.trim();
    if (!trimmed || trimmed.startsWith("#")) continue;
    const index = trimmed.indexOf("=");
    if (index === -1) continue;
    const key = trimmed.slice(0, index).trim();
    const value = trimmed.slice(index + 1).trim().replace(/^"|"$/g, "");
    env[key] = value;
  }
  return env;
}

async function loadEnv() {
  try {
    const content = await readFile(join(__dirname, ".env"), "utf-8");
    return parseEnv(content);
  } catch {
    return {};
  }
}

const env = { ...process.env, ...(await loadEnv()) };
const PORT = Number(env.PORT || 5173);
const CLIENT_ID = env.STRAVA_CLIENT_ID || "";
const CLIENT_SECRET = env.STRAVA_CLIENT_SECRET || "";
const MAPBOX_PUBLIC_TOKEN = env.MAPBOX_PUBLIC_TOKEN || "";

function json(res, status, payload) {
  res.writeHead(status, { "Content-Type": "application/json" });
  res.end(JSON.stringify(payload));
}

async function handleConfig(res) {
  if (!CLIENT_ID || !CLIENT_SECRET) {
    json(res, 500, {
      error: "Missing STRAVA_CLIENT_ID or STRAVA_CLIENT_SECRET in .env",
    });
    return;
  }

  json(res, 200, {
    clientId: CLIENT_ID,
    tokenExchangeUrl: "/oauth/token",
    mapboxToken: MAPBOX_PUBLIC_TOKEN,
  });
}

async function handleTokenExchange(req, res) {
  if (!CLIENT_ID || !CLIENT_SECRET) {
    json(res, 500, {
      error: "Missing STRAVA_CLIENT_ID or STRAVA_CLIENT_SECRET in .env",
    });
    return;
  }

  let body = "";
  for await (const chunk of req) {
    body += chunk;
  }

  let payload;
  try {
    payload = JSON.parse(body || "{}");
  } catch {
    json(res, 400, { error: "Invalid JSON" });
    return;
  }

  const response = await fetch("https://www.strava.com/oauth/token", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      ...payload,
      client_id: CLIENT_ID,
      client_secret: CLIENT_SECRET,
    }),
  });

  const text = await response.text();
  res.writeHead(response.status, {
    "Content-Type": response.headers.get("content-type") || "application/json",
  });
  res.end(text);
}

async function serveStatic(req, res) {
  const url = new URL(req.url, `http://${req.headers.host}`);
  const pathname = url.pathname === "/" ? "/index.html" : url.pathname;
  const filePath = join(__dirname, pathname);

  try {
    const fileStat = await stat(filePath);
    if (!fileStat.isFile()) {
      res.writeHead(404);
      res.end("Not found");
      return;
    }

    const data = await readFile(filePath);
    const ext = extname(filePath);
    res.writeHead(200, { "Content-Type": MIME_TYPES[ext] || "application/octet-stream" });
    res.end(data);
  } catch {
    res.writeHead(404);
    res.end("Not found");
  }
}

const server = createServer(async (req, res) => {
  if (req.method === "GET" && req.url === "/config") {
    await handleConfig(res);
    return;
  }

  if (req.method === "POST" && req.url === "/oauth/token") {
    await handleTokenExchange(req, res);
    return;
  }

  if (req.method === "OPTIONS") {
    res.writeHead(204, {
      "Access-Control-Allow-Origin": "*",
      "Access-Control-Allow-Methods": "GET,POST,OPTIONS",
      "Access-Control-Allow-Headers": "Content-Type",
    });
    res.end();
    return;
  }

  await serveStatic(req, res);
});

server.listen(PORT, () => {
  console.log(`Server running at http://localhost:${PORT}`);
});
