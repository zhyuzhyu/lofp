#!/usr/bin/env python3
"""
Legends of Future Past — Minimal Bot Example (Python)

Prerequisites: pip install websocket-client
Usage: LOFP_API_KEY=lofp_... python bot_example.py
"""

import json
import os
import sys
import time

import websocket

API_KEY = os.environ.get("LOFP_API_KEY")
SERVER = os.environ.get("LOFP_SERVER", "wss://lofp.metavert.io/ws/game")

if not API_KEY:
    print("Set LOFP_API_KEY environment variable")
    sys.exit(1)


def on_message(ws, message):
    msg = json.loads(message)
    msg_type = msg.get("type", "")
    data = msg.get("data", {})

    if msg_type == "auth_result":
        if data.get("success"):
            print(f"[AUTH] Logged in as {data.get('character', '?')}")
            # Send an initial command
            ws.send(json.dumps({"type": "command", "data": {"input": "look"}}))
        else:
            print(f"[AUTH] Failed: {data.get('error')}")
            ws.close()

    elif msg_type == "result":
        messages = data.get("messages", [])
        for line in messages:
            print(line)
        # React to game state here — e.g., auto-attack, auto-heal, navigate

    elif msg_type == "broadcast":
        messages = data.get("messages", [])
        for line in messages:
            print(f"  >> {line}")


def on_open(ws):
    print(f"[WS] Connected to {SERVER}")
    # Authenticate with API key
    ws.send(json.dumps({"type": "auth_apikey", "data": {"key": API_KEY}}))


def on_error(ws, error):
    print(f"[WS] Error: {error}")


def on_close(ws, close_status, close_msg):
    print("[WS] Disconnected")


if __name__ == "__main__":
    ws = websocket.WebSocketApp(
        SERVER,
        on_open=on_open,
        on_message=on_message,
        on_error=on_error,
        on_close=on_close,
    )
    ws.run_forever()
