# LoFP — Legends of Future Past

Resurrecting a 1990s MUD from original script files. The original game engine source code is lost; only the content scripts and documentation survive. We reverse-engineer a working game from those.

## Architecture

- **Backend**: Go + gorilla/mux + MongoDB at `engine/`
- **Frontend**: React 19 + TypeScript + Vite + Tailwind 4 at `frontend/`
- **Original Scripts**: `original/scripts/` (read-only reference, 333 .SCR files)
- Frontend dev server: port 4992, Backend server: port 4993
- Start both: `./start.sh`

## Build Check

```sh
cd engine && go build ./...
cd frontend && npx tsc --noEmit
```

## Reference Documentation

These original documents are the primary sources for game mechanics and world data:

| Document | Path | Description |
|----------|------|-------------|
| GM Manual | `original/GM Pages/MANUAL.DOC` | Comprehensive GM reference: combat, stats, items, monsters, skills, XP tables, all GM commands |
| GM Script Guide | `original/GMSCRIPT.DOC` | Script language reference: room/item/monster definitions, IFVERB/IFENTRY/IFSAY blocks, variables |
| Player Manual | `original/legends/LEGENDS.DOC` | Player-facing docs: commands, spells, psionics, skills, races, crafting |
| Session Capture | `original/legends/shirla.cap` | 1996 gameplay session capture — invaluable for combat output format, spell interactions, original message text |
| Script Config | `original/scripts/LEGENDS.CFG` | Master config listing all .SCR files to load in order |
| GM Pages | `original/GM Pages/` | Additional GM reference materials (various .DOC files) |

When implementing new features, always cross-reference these documents and the session capture to match original behavior.

## Multi-Machine Coordination

Production runs multiple Fly.io machines. ALL mutable world state must be coordinated via the MongoDB-backed hub (`engine/internal/hub/`):

- **Messages**: broadcasts, whispers, global announcements → published to `events` collection, delivered via Change Streams
- **Player presence**: WHO list, room occupancy → `presence` collection with TTL heartbeat
- **Room state changes**: item open/close/lock, item drops/pickups, script mutations (vals, itembits) → `room_state_change` events via hub
- **Any new mutable state** added to rooms, items, or global data MUST call `notifyRoomChange()` or publish through the hub, or players on different machines will see inconsistent worlds

## After Server Changes

After making changes to the backend (engine/), restart the Go server:
```sh
kill $(lsof -ti:4993) 2>/dev/null; sleep 1; cd engine && go run cmd/lofp/main.go &
```
Load .env first if needed: `source .env`

## Script Language

The game world is defined in a custom scripting language (documented in `original/GMSCRIPT.DOC`):
- **Rooms**: `NUMBER`, `NAME`, `*DESCRIPTION_START/END`, `EXIT`, `ITEM`, terrain, lighting
- **Items**: `INUMBER`, `NAME` (noun ref), type, weight, volume, substance, worn slots
- **Monsters**: `MNUMBER`, body parts, stats, AI strategy, weapons, spells
- **Events**: `IFVERB/IFPREVERB/IFSAY/IFENTRY/IFVAR...ENDIF` conditional blocks
- **Variables**: Named variables + internal vars (stats, time, flags, item vals)
- Config file: `original/scripts/LEGENDS.CFG` lists all scripts to load in order

## Current State (v10.0.2)

- Script parser loads 2273+ rooms, 1990+ items, 297 monsters with case-insensitive file loading
- Full combat system: original [ToHit/Roll] format, weapon crits/slayers, fatigue, weapon clash
- 60+ spells across 5 schools, 30+ psionic disciplines
- Complete crafting: mining, smelting, forging, weaving, dyeing, foraging, alchemy (32 recipes)
- 36 skills with build point costs, prerequisites, and mechanical effects
- Treasure system: coin drops, weapon/armor/scroll/chest drops scaled by monster TREASURE level
- Monster AI: hostile aggro, flee behavior, special attacks, guard, demand-based spawning
- ELSE branches in script conditionals, all portal types supported
- 150+ emotes, race-specific, submit-gated interactions
- Character soft-delete with admin recovery, unique first names
- Security: rate limiting, connection caps, chat flood protection, HTML sanitization
- WebSocket-based real-time multiplayer
- Admin panel for rooms/items/monsters/players/logs
- Production: Fly.io (2 machines, ord region) at lofp.metavert.io

## Units

| Unit | Code | Path | Description |
|------|------|------|-------------|
| Engine | ENG | `engine` | Go backend: script parser, game engine, command interpreter, MongoDB persistence |
| Frontend | FRONT | `frontend` | React + Tailwind: player UI (text client) and admin interface |
| Scripts | SCR | `original` | Original game script files and documentation (read-only reference) |
