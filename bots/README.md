# Legends of Future Past — Bot API

> **Note:** Bots were not part of the original 1992-1999 game. This is a new feature added to help repopulate the world! Bots can't do anything a human player can't — they follow the same rules, rate limits, and game mechanics.

Control game characters programmatically via the WebSocket API.

## Getting an API Key

1. Log in at [lofp.metavert.io](https://lofp.metavert.io)
2. On the main menu, click the ⚙ icon next to your character
3. Click **Generate Key** (optionally check "Allow GM commands" if the character is a GM)
4. **Copy the key immediately** — it's only shown once
5. If you need a new key, generate again (the old one is revoked)

## Connection Flow

1. Connect to `wss://lofp.metavert.io/ws/game` (or `ws://localhost:4993/ws/game` for local)
2. You'll receive a welcome message
3. Send an `auth_apikey` message with your key
4. On success, your character is logged in and you'll receive the room description
5. Send `command` messages to play, receive game output in responses

## Message Format

All messages are JSON with `type` and `data` fields.

### Sending

```json
// Authenticate
{"type": "auth_apikey", "data": {"key": "lofp_abc123..."}}

// Send a game command
{"type": "command", "data": {"input": "look"}}
{"type": "command", "data": {"input": "attack skeleton"}}
{"type": "command", "data": {"input": "say Hello everyone!"}}
```

### Receiving

```json
// Auth result
{"type": "auth_result", "data": {"success": true, "character": "MyBot"}}

// Game output (from your commands)
{"type": "result", "data": {"messages": ["You look around..."], "roomName": "[City Gate]", ...}}

// Broadcast (from other players/monsters in your room)
{"type": "broadcast", "data": {"messages": ["A skeleton attacks you!"]}}
```

### Key Response Fields

| Field | Description |
|-------|-------------|
| `messages` | Array of text lines to display |
| `roomName` | Current room name (on LOOK) |
| `roomDesc` | Room description text |
| `exits` | Array of exit direction names |
| `items` | Array of visible item names |
| `playerState` | Your character's current stats (BP, mana, etc.) |
| `promptIndicators` | Status codes: `!`=bleeding, `J`=joined, `S`=stunned, etc. |

## Bot Behavior

- Bots appear on the WHO list with `[Bot]` next to their name
- Bots follow all normal game rules (combat, death, rate limiting)
- GM bots can use @ commands only if "Allow GM commands" was checked when generating the key
- Rate limit: 4 commands/sec burst, 10 commands per 10 seconds sustained

## Examples

- [Python](bot_example.py) — minimal bot using `websocket-client`
- [Node.js](bot_example.js) — minimal bot using `ws`
- [TypeScript](bot_example.ts) — typed bot using `ws`

## Running the Examples

```bash
# Python
pip install websocket-client
LOFP_API_KEY=lofp_... python bots/bot_example.py

# Node.js
npm install ws
LOFP_API_KEY=lofp_... node bots/bot_example.js

# TypeScript
npm install ws @types/ws tsx
LOFP_API_KEY=lofp_... npx tsx bots/bot_example.ts
```
