/**
 * Legends of Future Past — Minimal Bot Example (TypeScript)
 *
 * Prerequisites: npm install ws @types/ws tsx
 * Usage: LOFP_API_KEY=lofp_... npx tsx bot_example.ts
 */

import WebSocket from "ws";

const API_KEY = process.env.LOFP_API_KEY;
const SERVER = process.env.LOFP_SERVER || "wss://lofp.metavert.io/ws/game";

if (!API_KEY) {
  console.error("Set LOFP_API_KEY environment variable");
  process.exit(1);
}

interface GameMessage {
  type: string;
  data: {
    success?: boolean;
    error?: string;
    character?: string;
    messages?: string[];
    roomName?: string;
    roomDesc?: string;
    exits?: string[];
    items?: string[];
    promptIndicators?: string;
    playerState?: {
      firstName: string;
      bodyPoints: number;
      maxBodyPoints: number;
      mana: number;
      maxMana: number;
      level: number;
      experience: number;
    };
  };
}

const ws = new WebSocket(SERVER);

function send(type: string, data: Record<string, unknown>): void {
  ws.send(JSON.stringify({ type, data }));
}

ws.on("open", () => {
  console.log(`[WS] Connected to ${SERVER}`);
  send("auth_apikey", { key: API_KEY });
});

ws.on("message", (raw: WebSocket.Data) => {
  const msg: GameMessage = JSON.parse(raw.toString());

  switch (msg.type) {
    case "auth_result":
      if (msg.data.success) {
        console.log(`[AUTH] Logged in as ${msg.data.character}`);
        send("command", { input: "look" });
      } else {
        console.error(`[AUTH] Failed: ${msg.data.error}`);
        ws.close();
      }
      break;

    case "result":
      for (const line of msg.data.messages ?? []) {
        console.log(line);
      }
      // Access player state if needed
      if (msg.data.playerState) {
        const ps = msg.data.playerState;
        console.log(`[STATE] BP: ${ps.bodyPoints}/${ps.maxBodyPoints}, Mana: ${ps.mana}/${ps.maxMana}`);
      }
      break;

    case "broadcast":
      for (const line of msg.data.messages ?? []) {
        console.log(`  >> ${line}`);
      }
      break;
  }
});

ws.on("error", (err: Error) => console.error(`[WS] Error: ${err.message}`));
ws.on("close", () => console.log("[WS] Disconnected"));
