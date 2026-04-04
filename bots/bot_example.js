/**
 * Legends of Future Past — Minimal Bot Example (Node.js)
 *
 * Prerequisites: npm install ws
 * Usage: LOFP_API_KEY=lofp_... node bot_example.js
 */

const WebSocket = require("ws");

const API_KEY = process.env.LOFP_API_KEY;
const SERVER = process.env.LOFP_SERVER || "wss://lofp.metavert.io/ws/game";

if (!API_KEY) {
  console.error("Set LOFP_API_KEY environment variable");
  process.exit(1);
}

const ws = new WebSocket(SERVER);

function send(type, data) {
  ws.send(JSON.stringify({ type, data }));
}

ws.on("open", () => {
  console.log(`[WS] Connected to ${SERVER}`);
  send("auth_apikey", { key: API_KEY });
});

ws.on("message", (raw) => {
  const msg = JSON.parse(raw);
  const { type, data } = msg;

  if (type === "auth_result") {
    if (data.success) {
      console.log(`[AUTH] Logged in as ${data.character}`);
      send("command", { input: "look" });
    } else {
      console.error(`[AUTH] Failed: ${data.error}`);
      ws.close();
    }
  } else if (type === "result") {
    for (const line of data.messages || []) {
      console.log(line);
    }
    // React to game state here
  } else if (type === "broadcast") {
    for (const line of data.messages || []) {
      console.log(`  >> ${line}`);
    }
  }
});

ws.on("error", (err) => console.error(`[WS] Error: ${err.message}`));
ws.on("close", () => console.log("[WS] Disconnected"));
