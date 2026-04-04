# Changelog

## v0.97 — 2026-04-04

### Combat System
- **ATTACK/KILL <monster>** — original format output: `[ToHit: X, Roll: Y] Hit!/Miss./Excellent Hit!`
- **Damage severity tiers**: Puny, Grazing, Insignificant, Minor, Passable, Good, Masterful, Grisly, Severe, Ghastly
- **Attack verbs by weapon type**: swings (slash), thrusts (pierce/pole), slashes (claw)
- **Body part targeting**: head, body, right/left arm, right/left leg, back; animal parts for non-humanoids
- **Weapon elemental crits (VAL3)**: 10-50% chance heat/cold/electric bonus damage with VAL5 max
- **Racial slayer weapons (VAL3 21-32)**: bonus damage vs Undead, Dragon, Giant, Troll, etc.
- **Weapon poison (VAL4 51-100)**: delivers poison on hit
- **MAGICWEAPON gating**: some monsters require enchanted weapons (level 1/2/3), shows TEXI text
- **Monster guard behavior**: guard monsters intercept attacks on their charge
- **Cry for law (strategy 1-25, 101-125)**: attacking lawful NPCs alerts nearby guards
- **Monster poison/disease/fatigue**: % chance per hit to inflict conditions
- **EXTRABODY**: monsters have 50-100% extra HP not counted toward XP
- **Weather combat modifiers**: rain -10, heavy rain -20, thunderstorm -30, gale -40, hurricane -50 (outdoor only)
- **Arena rooms**: ExtraMod ARENA prevents lethal damage
- **Alignment shifts**: killing evil monsters → more good, good → more evil
- **Hostile monsters** (strategy 301+) auto-attack players entering rooms
- **Monster flee AI**: strategy-based (fight to death 501+, flee when wounded 301-500, animals flee 201-300)
- **Monster special attacks**: SPECUSE/SPECDMG with TEXX text overrides
- **FLEE command**: escape combat (quickness/agility based chance, random exit)
- **Combat stances**: OFFENSIVE (+15/-15), DEFENSIVE (-15/+15), BERSERK (+25/-25, Murg), WARY (-5/+5)
- **Roundtime**: `[Round: X sec]` format, quickness-based
- **Death**: Eternity, Inc. message (matching original text) → DEPART to respawn
- **Real XP table**: 100-level XP/build-point table from GM Manual, build points = 20 + 10*level

### Magic System
- **PREPARE <spell>** then **CAST [target]** — two-step spell casting
- **60+ spells** across 5 schools with mana costs and casting times
- **Offensive**: Flame Bolt, Force Blade, Lightning Bolt, Freezing Sphere, Call Meteor, Chain Lightning, Energy Maelstrom
- **Healing**: Body Restoration I/II/III, Invigoration I/II, Reconstruction, Regeneration
- **Defense**: Mystic Armor (+20), Globe of Protection (+50/+100), Spectral Shield (+20), Spell Shield (+25)
- **Buffs**: Strength I/II/III, Agility I/II/III, Fly, Invisibility, Haste
- **Spellcraft skill checks**: school skill + spellcraft + willpower vs spell level
- **Magic resistance**: monsters can resist spells, elemental immunities apply

### Psionic System
- **PSI <discipline>** then **PROJECT [target]** — two-step psionic projection
- **Mind over Matter**: Kinetic Thrust, Pyrokinetics, Cryokinetics, Electrify, Wall of Force (+25), Force Field (+75), Flight
- **Mind over Mind**: Psychic Blast, Psychic Crush, Terror, Pain, Screen/Shield/Barrier/Fortress defense chain
- **Psi point costs**, skill checks (Psionics + school skill + willpower), psi resistance

### World Systems
- **SKIN <dead monster>** — weighted random drops from SkinItem definitions with SkinAdj
- **Container traps (VAL4)**: 13 trap types (needle, gas, acid, blade, explosive, glyph spells) trigger on OPEN
- **Highlander BLEND**: hide in mountain/cave terrain (race-restricted)

### Race-Specific Emotes
- **Drakin**: flick tongue, bare teeth, spread/fold wings, swish tail
- **Aelfen**: rub ears · **Highlander**: pull beard
- **Wolfling**: scratch behind ear, bare fangs, chase tail, scent air, whine, droop tail
- **23 self-emotes**: fume, squint, hum, sneeze, crack knuckles, bat eyelashes, bounce, strike pose, and more

### Infrastructure
- **PlayerMessageFunc**: targeted messages from background tasks (combat, aggro)
- **Monster combat tick**: integrated into 3-second behavior loop
- **Combat disengagement**: moving rooms, fleeing, or target dying clears combat state
- **Character name validation**: reserved-word filter (monster names, profanity, game terms)

## v0.96 — 2026-04-04

### Monster System
- **Monster spawning**: Monsters appear in room descriptions from original MonsterList data
- **TEX1-4 ambient text**: Monsters emit random flavor text on a timer based on Speed
- **Wandering**: Non-hostile monsters (strategy < 301) wander between rooms via exits
- **TEXG/TEXE/TEXM**: Custom text for spawn, room entry, and movement (direction appended)
- **Examine monsters**: `exam skeleton` shows monster description
- **Emote targeting**: Monsters targetable by all emotes (`point skeleton`, etc.)
- **@spawn/@genmon fixed**: GM commands now actually create monster instances
- **@genmon creates sedated** (inactive), **@spawn creates active** (will act)
- **Speed-based ticking**: Speed 1 = every 3s, Speed 3 (default) = every 9s

### New Emotes (30+)
- **lick, nibble, bark, claw, curse, duck, hiss, hold, hula, jig, moan, massage, pinch, play, purr, roar, snarl, snuggle, wag, wait, write, yowl, thump, applaud, peer, grunt, dip, handraise, handshake, headshake, pick, gesture**
- **Self-targeting overrides**: `spit me` → drool, `lick me` → lick lips, `laugh me` → laugh at self, `kick me` → kick self, `thump me` → thump head
- **KISS body parts**: head, nose, lips, ears, neck, chest, hand, navel, leg, knee, feet
- **Submit-gated**: lips, navel, leg, knee, feet require target to be submitting

### Submit System
- **SUBMIT/UNSUBMIT** commands — accept intimate emotes from other players
- **LICK** behavior changes: non-submitted → passionate kiss, submitted → full body lick
- Moving to a new room automatically clears submit state

### Infrastructure
- **CEVENT ECHO delivery**: Script ECHO messages from cyclic events now broadcast to players in rooms
- **RoomBroadcast callback**: Engine background tasks (monsters, CEVENTs) can push messages to rooms via WebSocket
- Monster article handling: "an orc" vs "a skeleton", unique monsters without article

## v0.95 — 2026-04-03

### Lock & Unlock
- **LOCK/UNLOCK commands** — match KEY items via Val3, proper feedback for missing/wrong keys
- Latched items block open; locked items block passage

### Ordinal Targeting
- **"2 gate", "other gate", "second gate"** — target the Nth matching item in a room
- Works across all 19 item-matching functions (get, drop, look, open, close, wield, wear, etc.)

### Mechanoid Emote
- **EMOTE/UNEMOTE** — Mechanoid racial ability to toggle emotional state (race 7 only)
- **ACT** remains the general-purpose roleplaying command for all races

### Verb Aliases & Commands
- **ORDER** as BUY synonym
- **UNLIGHT/IGNITE** as EXTINGUISH/LIGHT aliases
- **QUAFF** as DRINK alias, **SHOUT** as YELL alias, **PLACE** as DROP alias
- **RECALL** with no args runs room-level IFVERB RECALL -1 scripts
- **ACTBRIEF/RPBRIEF** toggle commands
- **POUR** verb stub
- **A** as shorthand for ACT (freeform roleplay)

## v0.94 — 2026-04-03

### Script Engine
- **226 named global variables** (DANWATER, TECHSWITCH, etc.) loaded from VARIABLE definitions, synchronized across machines
- **MUL/DIV/MOD** arithmetic operations in scripts
- **GENMON/ZAPMON** — spawn and remove monsters from scripts
- **GFLAG** — set flags for all players in a room
- **RELOGIN/IFLOGIN** — force player room on login
- **IFFULLDESC/IFIN** — new conditional blocks
- **NEWPUT** — place items inside containers from scripts
- **DUMMY1-5** volatile scratch variables
- **PVAL 0-19** persistent system-wide variables (MongoDB-backed)
- **CEVENT system** — cyclic timed events parsed and executing on 3-second ticks
- **Implicit ENDIF** — parser handles missing ENDIF when new verb block starts (matches original engine)
- **Room-level IFVERB blocks** now fire from doItemInteraction and position commands
- **Direction verb resolution** — "O" resolves to "OUT" for IFPREVERB matching, all abbreviations expanded

### Parsing
- **REGIONDEF** expanded: DepartRoom, Weather, Treasure, Teleport, Summoning, spell modifiers, MineAdj
- **MONEYDEF** — multi-currency per region definitions
- **FORAGEDEF/MINDEF** — full forage and mining definitions parsed
- **Room REGION field** parsed and accessible in scripts
- **Monster psionics** — PSI, PSIUSE, PSISKILL, PSIRESIST, PSILEVEL, DISCIPLINE
- **Monster combat** — IMMUNITY (10 types), WEAPON, WEAPONPLUS, MAGICWEAPON, SPECUSE/SPECDMG, EXTRABODY, FATIGUE

### Food & Drink
- **Bite/sip tracking** — PARAMETER1 defines bites for food, sips for liquid. VAL2 tracks remaining. Items removed only when empty.
- **Spell on consume** — ITEMVAL3 triggers spell effect on first bite (e.g., Mindlink from thesnia leaf)
- **Item scripts run on eat** — root-level IFVAR blocks execute before checking spell effects

### Items & Verbs
- **Store adjective in ADJ3** — bought items place variety adjective in last slot, leaving ADJ1/ADJ2 for crafting
- **Worn/wielded item visibility** — all verbs (touch, sniff, wave, emotes) now find inventory, worn, and wielded items
- **SIT/LAY/KNEEL/STAND trigger IFVERB scripts** — position commands run room-level verb scripts
- **CLIMB handles MOVE** — non-portal items with IFPREVERB CLIMB + MOVE now work
- **SNIFF/SMELL/LISTEN** try item interaction first, fall back to emote
- **@peek named variables** — `@peek DANWATER` works, plus PVAL and full getVar fallback
- **@yank fix** — updates live session player, not just database

### Telepathy & Spells
- **Mindlink (#403)** — first working spell, grants 1hr telepathy via food consumption
- **Telepathy system** — THINK broadcasts to telepathy-enabled players only
- **Ephemeral innate telepathy** — race 8 starts with telepathy active
- **TELEPATHY command** — toggle on/off

### Original Fidelity
- Global messages: "has just entered/left the Realms"
- WHO: 4-column grid with "There are N adventurers in the Realms."
- EAT: "You take a bite of X. (N bites remaining)"
- Death: Eternity, Inc. message + DEPART stub
- HP → BP (Body Points) throughout UI
- Article fix: "an axe" not "a axe"

### Admin & Infrastructure
- Event Monitor: real-time admin WebSocket feed
- Session Capture: record, view, download gameplay
- Capture delete with confirmation
- /healthz endpoint for vibectl
- Backend connection indicator
- Admin Players/Users sorted by recency with relative timestamps

## v0.93 — 2026-04-03

### New Commands
- **HIDE/SNEAK**: Stealth system — conceal yourself and move while hidden
- **FLY/ASCEND/DESCEND/LAND**: Flight system for Drakin race and spell-granted flight
- **MARK**: Set teleport anchor points (1-10) for future spell use
- **BALANCE**: View bank account balance
- **SPELL**: List all known spells
- **THINK**: Telepathy broadcast to all players
- **CANT**: Covert message requiring Legerdemain skill
- **UNDRESS**: Remove outermost worn item
- **UNPROMPT**: Toggle off prompt indicators
- **VERSION/CREDITS**: Game info

### Combat & Spell Stubs (40+ verbs)
- All combat verbs stubbed with TODO: ATTACK, KILL, ADVANCE, RETREAT, GUARD, BACKSTAB, BITE, AVOID
- Combat stances: BERSERK, FRENZY, DEFENSIVE, OFFENSIVE, WARY, NORMAL
- Spell verbs: INVOKE, PREPARE, CHANT, COMMAND, MASTER
- Ranged: NOCK, LOAD, SPECIALIZE
- Skills: DISARM, STEAL, STALK, TEACH, SELFTRAIN, UNLEARN, ANOINT, TRAP, SURVEY, SPLIT
- Racial: BLEND, CALL, TRANSFORM, MOLD, DISGUISE
- Social: SUBMIT, ARREST, ENROLL, INITIATE, JOIN, FOLLOW, LEAVE, DISBAND
- Other: TEND, BREAK, PUT, FILL, SKIN, REPAIR, WORK

### Script Variables Expansion
- Physical attributes: HEI/HEIT, WEI/WEIT, AGE/AGET
- Form states: WOLFFORM, SLIMEFORM, OTHERFORM, UNDEAD, DISGUISED
- Status: SLEEPING, SUBMITTING, ROUNDTIME, SPELLNUM, POSITION
- Wealth: WEALTH (total copper)
- Room: WILDERNESS, ASTRAL, TERRAIN (numeric), OBJWEIGHT

### Player System
- New fields: Height, Weight, Age (with true variants), Marks, PreparedSpell
- Form states: WolfForm, SlimeForm, Disguised
- Status: Sleeping, Submitting, Undead

### Fixes
- **Article fix**: "an axe" instead of "a axe" — auto-detects vowel sounds
- **Body Points**: All "HP" references changed to "BP" throughout UI
- **GO command**: Non-portal items with IFPREVERB GO scripts now work (e.g., stairways)
- **CLEARVERB + MOVE**: Scripts that block default GO but provide MOVE now work correctly
- **Session Capture**: Fixed null array bug preventing line recording
- **Copy-paste**: Terminal text now selectable on Windows
- **Stop Recording**: Fixed stale WebSocket reference

### Infrastructure
- `/healthz` endpoint for vibectl health monitoring
- Backend connection indicator in nav bar
- Character list shows "Connecting to server..." when backend is unavailable
- Create Character button hidden until backend responds
- Event Monitor disconnect messages

## v0.92 — 2026-04-03

### Admin Event Monitor
- Real-time WebSocket event feed for admin monitoring
- Categories: system, time, monster, script, world state, weather
- Background time cycle publishes hour/day/night transitions
- Category filter, clear button, auto-scroll

### Session Capture
- Record game sessions to MongoDB
- Start/stop from modal overlay during gameplay
- View previous captures with color-coded output
- Download as .txt file
- Auto-stops on disconnect

## v0.91 — 2026-04-03

### Script Engine Expansion
- **80+ variables** supported in IFVAR conditions: player stats (STR/AGI/CON/QUI/WIL/PER/EMP), resources (BODYPOINTS/MANAPOINTS/PSIPOINTS/FATPOINTS), state (DEAD/FLYING/KNEELING/HIDDEN), organization (ORG/ORGRANK/ALIGN), room info (RNUM/OUTDOOR/TERRAIN/PLRSINROOM/MONINROOM), time (TIM/DAY/NIGHT/DATE/MONTH/YEAR), weather (WEA), exits (EXITN/EXITS/etc.), flags (FLAG1-4), skills (SKILL0-35)
- **IFSAY** script blocks now execute on player speech (enables NPC dialogue and quest triggers)
- **IFCARRY** condition checks if player carries a specific item
- **IFNOITEM** condition (negation of IFITEM)
- **IFTOUCH** block execution for touch-type verbs
- **AFFECT** action switches script context to another room for multi-room effects
- **RANDOM** action for random number generation in scripts
- **DAMAGEPLR** action for environmental damage (traps, falls)
- **STRCVT** action with %0-%9 string substitution in ECHO text
- **POSITION** action forces player position from scripts
- **String substitution** expanded: %s (he/she), %i (him/her), %h (his/her), %p (group name), %m (monster), %c (newline), %0-%9 (STRCVT)

### World Systems
- **In-game clock and calendar**: 343-day year, 12 months, day/night cycle (1 real minute = 1 game hour)
- **Monster spawning**: Monsters populate rooms from MonsterList data, displayed in room descriptions
- **Weather system**: 15 weather states per region, shown in outdoor room descriptions
- **Room lighting**: Dark rooms require light sources; DAY_LIGHT rooms dark at night

### New Commands
- **DRINK/SIP**: Consume liquid/food items
- **LIGHT/EXTINGUISH**: Light and douse LIGHTABLE items for dark room navigation
- **FLIP**: Toggle FLIPPED/UNFLIPPED state on FLIPABLE items
- **LATCH/UNLATCH**: Toggle latched state; latched items can't be opened
- **DEPOSIT/WITHDRAW**: Banking in BANK rooms
- **TRAIN**: Skill training in rooms with TRAINING definitions (36 skills)
- **MINE**: Mining in MINEA/B/C rooms (stub)
- **FORAGE**: Foraging in wilderness terrain (stub)
- **CAST**: Spell casting with 150+ registered spells across 5 schools (stub behavior)
- **CRAFT/FORGE/SMELT/WEAVE/DYE/BREW/ANALYZE**: Crafting stubs

### Spell Registry
- 150+ spells registered across 5 schools: Conjuration (100-144), Enchantment (200-250), Necromancy (301-356), General Magic (400-415), Druidic (500-538)
- CAST command recognizes all spells by name prefix

### Monster System
- Expanded MonsterDef with 20+ new fields: alignment, magic resistance, mana, spell use, poison/disease, skin items, text overrides, guard, stealable, eternal, discorporate
- All monster fields parsed from scripts
- Passive monster spawning in rooms

### Player Systems
- Organization/Guild fields (ORG, ORGRANK)
- Alignment tracking
- Build points
- Banking (BankGold/Silver/Copper)
- Known spells registry
- Transient flags (FLAG1-4, reset on room entry)
- 36 named skills with training system

### Training System
- TRAINING room definitions parsed from scripts
- TRAIN command with skill listing, level caps, and gold cost
- Cost formula: (current_level+1)^2 * 10 copper per level

## v0.9 — 2026-04-03

### Game Engine
- Script parser loads 2273 rooms, 1990 items, 297 monsters from original .SCR files
- Fixed description parser: `*DESCRIPTION_START ITEM READ/EXAM` no longer overwrites room descriptions
- Fixed item deduplication: later script definitions correctly override earlier ones (ANTI.SCR placeholders replaced)
- Fixed script block parser: missing ENDIF no longer consumes subsequent item definitions
- Seasonal scripts (ASCRIPT/WSCRIPT/SSCRIPT/PSCRIPT) skipped; base definitions used
- Formatted text preservation: descriptions with leading whitespace or blank lines retain formatting (poems, maps, etc.)

### Movement & Portals
- Portal traversal via GO command with VAL2 destination
- PORTAL_CLIMB, PORTAL_UP, PORTAL_DOWN, PORTAL_OVER, PORTAL_THROUGH support
- CLIMB verb for climbable portals
- Closed/locked portals block passage
- Directional look (LOOK N, LOOK NORTH, etc.) shows adjacent room description

### Script Execution
- IFENTRY blocks execute on room entry (movement, login, character creation)
- IFVAR conditions: INTNUM, ITEMBIT, ITEMVAL, ITEMADJ, LEV, RAC, SKILL, MISTFORM
- IFITEM conditions: check item open/closed/locked/unlocked state
- IFPREVERB/IFVERB execution on item interactions
- Actions: ECHO (PLAYER/ALL/OTHERS), EQUAL, ADD, SUB, NEWITEM, GMMSG, CLEARVERB, MOVE, SHOWROOM
- Actions: LOCK, UNLOCK, OPEN, CLOSE, REMOVEITEM, SETITEMVAL
- Nested conditional blocks supported
- Script text placeholders: %N, %n, %a, %h, %e, %o

### Commands
- Movement: N/S/E/W/NE/NW/SE/SW/UP/DOWN/OUT, GO, CLIMB
- Looking: LOOK, EXAMINE, INSPECT, LOOK IN/ON/UNDER/BEHIND, directional LOOK
- Items: GET, DROP, INVENTORY, WIELD, UNWIELD, WEAR, REMOVE
- Containers: OPEN, CLOSE
- Interaction: PULL, PUSH, TURN, RUB, TAP, TOUCH, SEARCH, DIG, RECALL
- Communication: SAY ('), WHISPER, YELL, RECITE, EMOTE
- Commerce: BUY, SELL (1482 store items across 20+ shops)
- Social: 60+ emotes (SMILE, BOW, KICK, etc.) with targeted second-person messages
- Info: STATUS, HEALTH, WEALTH, SKILLS, WHO, HELP, ASSIST
- Position: SIT, STAND, KNEEL, LAY
- Roleplay: ACT, EMOTE, RECITE
- READ items with room-scoped descriptions

### Multiplayer
- WebSocket-based real-time gameplay
- Cross-machine coordination via MongoDB Change Streams
- Player presence synced across multiple Fly.io machines
- Room state changes (item state, drops, pickups, script mutations) coordinated
- Emote targeting: second-person messages to target, third-person to room
- Room broadcasts, global broadcasts, GM broadcasts, whispers
- WebSocket reconnection with exponential backoff

### Characters & Auth
- Google OAuth login with 30-day JWT sessions
- Character creation with 8 races, stat rolling, name validation
- Character name uniqueness enforced (no logging into others' characters)
- Session persistence across server restarts
- Starting gear granted via IFENTRY scripts at room 201

### Commerce
- BUY command with store items (STOREITEM parsed from scripts)
- Efficient currency deduction (copper first, then silver, then gold with change)
- SELL command in rooms with BUY_ARMOR/BUY_SKINS/BUY_JEWELRY modifiers
- Price display as gold/silver/copper

### Admin Interface
- Rooms tab: searchable index with detail view, clickable exits, enriched items
- Items tab: searchable index with full properties, type/slot expansion
- Monsters tab: searchable index with combat stats and properties
- Players tab: character detail, GM toggle, account reassignment
- Users tab: account detail, admin toggle, character list
- Logs tab: filterable game event log with hyperlinked players/users
- Exact number match sorting in search results

### Logging
- MongoDB-backed game log with 90-day TTL
- Events: user login/logout, character game enter/exit, character creation, GM grant/revoke
- Log entries include user name, email, account ID
- Admin logs UI with event/player filtering

### Security
- Admin auth required for GM toggle, game world API endpoints
- WebSocket origin validation against frontend URL
- 64KB WebSocket message size limit
- Character name validation (alpha + ' + -, max 20 chars)
- Race/gender input validation
- Command input truncated to 500 characters

### Infrastructure
- Go + gorilla/mux backend, React 19 + TypeScript + Vite + Tailwind 4 frontend
- MongoDB Atlas for persistence
- Fly.io production deployment (2 machines, ord region)
- Custom domain: lofp.metavert.io
- Multi-stage Docker build (14MB production image)
- Separate dev/prod configs
