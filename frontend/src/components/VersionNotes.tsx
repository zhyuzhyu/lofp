export default function VersionNotes({ onBack }: { onBack: () => void }) {
  return (
    <div className="flex items-start justify-center h-full p-8 overflow-y-auto">
      <div className="max-w-3xl w-full font-mono">
        <div className="flex items-center justify-between mb-6">
          <h1 className="text-amber-500 text-2xl font-bold">Version Notes</h1>
          <button onClick={onBack} className="text-gray-400 hover:text-white text-sm">&larr; Back</button>
        </div>

        <div className="space-y-6 text-sm">
          <section>
            <h2 className="text-amber-400 text-lg font-bold mb-1">v10.0.1 &mdash; April 4, 2026</h2>
            <p className="text-gray-400 mb-3">Character management, script variables, admin tools, and bug fixes.</p>

            <div className="space-y-4 mb-8">
              <div>
                <h3 className="text-green-400 font-bold mb-1">Character Management</h3>
                <ul className="text-gray-300 space-y-1 ml-4 list-disc">
                  <li>Delete characters from main menu (soft-delete with confirmation modal)</li>
                  <li>Unique first names enforced on character creation</li>
                  <li>Admin: browse deleted characters, recover with optional rename</li>
                  <li>Name validation split: exact match for monsters, substring for slurs only</li>
                </ul>
              </div>
              <div>
                <h3 className="text-green-400 font-bold mb-1">Script Engine</h3>
                <ul className="text-gray-300 space-y-1 ml-4 list-disc">
                  <li>New variables: WARRANT, GFLAG1-4, NUMPLRS, ARENADEATH</li>
                  <li>Position variables: SITTING, LAYING, STANDING, KNEELING</li>
                  <li>WIELDED, WEALTH, REGION variables added</li>
                  <li>Warrant field on player for law/crime system</li>
                </ul>
              </div>
              <div>
                <h3 className="text-green-400 font-bold mb-1">Fixes</h3>
                <ul className="text-gray-300 space-y-1 ml-4 list-disc">
                  <li>Fixed login blocked by reserved-word check (&ldquo;Pendragon&rdquo; contains &ldquo;dragon&rdquo;)</li>
                  <li>Name validation moved to character creation only, not login</li>
                  <li>Version notes accessible via /version-notes deeplink</li>
                </ul>
              </div>
            </div>

            <h2 className="text-amber-400 text-lg font-bold mb-1">v10.0.0 &mdash; April 4, 2026</h2>
            <p className="text-gray-400 mb-3">Major milestone: complete combat, magic, psionics, crafting, alchemy, and full skill system.</p>

            <div className="space-y-4 mb-8">
              <div>
                <h3 className="text-green-400 font-bold mb-1">Crafting System</h3>
                <ul className="text-gray-300 space-y-1 ml-4 list-disc">
                  <li>MINE ore from MINEA/B/C rooms (grade determines purity)</li>
                  <li>SMELT ore into refined metal at FORGE rooms</li>
                  <li>CRAFT weapons/armor at FORGE, clothing at LOOM, wood items at FLETCHER</li>
                  <li>FORAGE terrain-based materials (wood, plants, reagents, dyes)</li>
                  <li>DYE materials at LOOM rooms with natural and crafted dyes</li>
                  <li>ANALYZE ore purity and reagent properties</li>
                  <li>BREW potions via alchemy &mdash; 32 recipes from original game data</li>
                </ul>
              </div>
              <div>
                <h3 className="text-green-400 font-bold mb-1">Skill System</h3>
                <ul className="text-gray-300 space-y-1 ml-4 list-disc">
                  <li>36 skills with build point costs from original documentation</li>
                  <li>Skill prerequisites enforced (magic needs Spellcraft, etc.)</li>
                  <li>Weapon skills: +5 attack per rank &middot; Dodge: +5 defense per rank</li>
                  <li>Martial Arts: +5 attack/+2 defense unarmed, 10+ hits magic monsters</li>
                  <li>Combat Maneuvering: -1s roundtime, 2%/rank dodge special attacks</li>
                  <li>Endurance: +4 BP per rank, 1%/rank elemental damage reduction</li>
                  <li>ANOINT weapons with poison (Trap &amp; Poison Lore)</li>
                  <li>TEND wounds with Healing skill (+50% same-race bonus)</li>
                </ul>
              </div>
              <div>
                <h3 className="text-green-400 font-bold mb-1">Treasure &amp; Loot</h3>
                <ul className="text-gray-300 space-y-1 ml-4 list-disc">
                  <li>Treasure table system based on monster TREASURE level</li>
                  <li>Coin drops, weapon/armor drops, spell scrolls, locked chests</li>
                  <li>Magic weapon bonuses and premium materials (elkyri, adamantine)</li>
                  <li>Trapped chests with 13 trap types and spell glyphs</li>
                  <li>Monster weapon drops on death &middot; SEARCH corpses for loot</li>
                  <li>MONEY items auto-convert to currency on pickup</li>
                </ul>
              </div>
              <div>
                <h3 className="text-green-400 font-bold mb-1">Combat Fidelity</h3>
                <ul className="text-gray-300 space-y-1 ml-4 list-disc">
                  <li>Fatigue drain on melee attacks, fatigue ToHit penalties</li>
                  <li>Weapon clash on roll &lt; 3 &mdash; weapons can be damaged or broken</li>
                  <li>Backstab requires puncture weapon (daggers, rapiers)</li>
                  <li>Death = 90% XP penalty toward current build point</li>
                  <li>Spellcraft formula: 25% + EMP/10 + skill*5, fumble on 98+</li>
                  <li>Mana cost = spell level &middot; NOCK/LOAD ranged weapons with ammo</li>
                </ul>
              </div>
              <div>
                <h3 className="text-green-400 font-bold mb-1">Monster System</h3>
                <ul className="text-gray-300 space-y-1 ml-4 list-disc">
                  <li>Demand-based spawning: monsters appear when players enter rooms</li>
                  <li>Monsters unload after 3 minutes with no players (ETERNAL exempt)</li>
                  <li>Psi defense auto-activation on spawn (Wall of Force, Psychic Shield, etc.)</li>
                  <li>Hidden/Invisible distinction &mdash; Invisibility spell not broken by movement</li>
                  <li>Corpse decay after 60 seconds &middot; Dead monsters show as &ldquo;(dead)&rdquo;</li>
                </ul>
              </div>
            </div>

            <h2 className="text-amber-400 text-lg font-bold mb-1">v0.97 &mdash; April 4, 2026</h2>
            <p className="text-gray-400 mb-3">Combat, spells, psionics, and GM Manual fidelity &mdash; fight monsters, cast spells, project disciplines.</p>

            <div className="space-y-4 mb-8">
              <div>
                <h3 className="text-green-400 font-bold mb-1">Combat System</h3>
                <ul className="text-gray-300 space-y-1 ml-4 list-disc">
                  <li>ATTACK/KILL &lt;monster&gt; &mdash; original format: [ToHit: X, Roll: Y] Hit!/Miss./Excellent Hit!</li>
                  <li>Damage severity tiers: Puny, Grazing, Minor, Passable, Good, Masterful, Grisly, Severe, Ghastly</li>
                  <li>Attack verbs by weapon type: swings (slash), thrusts (pierce/pole), slashes (claw)</li>
                  <li>Body part targeting: head, body, arms, legs, back, tail</li>
                  <li>Weapon elemental crits (VAL3): 10-50% chance heat/cold/electric bonus damage</li>
                  <li>Racial slayer weapons (VAL3 21-32): bonus damage vs specific monster races</li>
                  <li>Weapon poison (VAL4): delivers poison on hit</li>
                  <li>MAGICWEAPON gating: some monsters require enchanted weapons to hit</li>
                  <li>Monster guard behavior: guards intercept attacks on their charge</li>
                  <li>Cry for law: attacking lawful NPCs alerts nearby guards</li>
                  <li>Monster poison, disease, and fatigue attacks on hit</li>
                  <li>EXTRABODY: monsters have extra HP not counted toward XP</li>
                  <li>Weather combat modifiers: rain/snow/storms reduce attack accuracy</li>
                  <li>Arena rooms prevent lethal damage</li>
                  <li>Alignment shifts on monster kills</li>
                  <li>Hostile monsters (strategy 301+) auto-attack on room entry</li>
                  <li>Monster flee AI based on strategy type and HP percentage</li>
                  <li>Combat stances: OFFENSIVE, DEFENSIVE, BERSERK (Murg), WARY, NORMAL</li>
                  <li>FLEE to escape combat, [Round: X sec] roundtime</li>
                  <li>Death &rarr; Eternity, Inc. &rarr; DEPART to respawn</li>
                  <li>Real XP/build-point table from original GM Manual (100 levels)</li>
                </ul>
              </div>
              <div>
                <h3 className="text-green-400 font-bold mb-1">Magic System</h3>
                <ul className="text-gray-300 space-y-1 ml-4 list-disc">
                  <li>PREPARE &lt;spell&gt; then CAST [target] two-step casting</li>
                  <li>60+ spells across 5 schools: Conjuration, Enchantment, Necromancy, General, Druidic</li>
                  <li>Offensive spells: Flame Bolt, Lightning Bolt, Freezing Sphere, Call Meteor, and more</li>
                  <li>Healing: Body Restoration I/II/III, Invigoration, Reconstruction, Regeneration</li>
                  <li>Defense: Mystic Armor (+20), Globe of Protection (+50/+100), Spectral Shield</li>
                  <li>Buffs: Strength I/II/III, Agility I/II/III, Fly, Invisibility, Haste</li>
                  <li>Mana costs, spellcraft skill checks, magic resistance</li>
                </ul>
              </div>
              <div>
                <h3 className="text-green-400 font-bold mb-1">Psionic System</h3>
                <ul className="text-gray-300 space-y-1 ml-4 list-disc">
                  <li>PSI &lt;discipline&gt; then PROJECT [target] two-step projection</li>
                  <li>Mind over Matter: Kinetic Thrust, Pyrokinetics, Cryokinetics, Electrify, Wall of Force, Flight</li>
                  <li>Mind over Mind: Psychic Blast, Psychic Crush, Terror, Pain, Psychic Screen/Shield/Barrier/Fortress</li>
                  <li>Psi point costs, psionic skill checks, psi resistance</li>
                </ul>
              </div>
              <div>
                <h3 className="text-green-400 font-bold mb-1">World Systems</h3>
                <ul className="text-gray-300 space-y-1 ml-4 list-disc">
                  <li>SKIN dead monsters for components (weighted random drops from SkinItem definitions)</li>
                  <li>Container traps: 13 types (needles, gas, acid, blades, explosives, glyph spells)</li>
                  <li>Highlander BLEND in mountain/cave terrain</li>
                </ul>
              </div>
              <div>
                <h3 className="text-green-400 font-bold mb-1">Race-Specific Emotes</h3>
                <ul className="text-gray-300 space-y-1 ml-4 list-disc">
                  <li>Drakin: flick tongue, bare teeth, spread/fold wings, swish tail</li>
                  <li>Aelfen: rub ears &middot; Highlander: pull beard &middot; Wolfling: bare fangs, chase tail, scent air</li>
                  <li>23 new self-emotes: fume, squint, hum, sneeze, crack knuckles, bat eyelashes, and more</li>
                </ul>
              </div>
            </div>

            <h2 className="text-amber-400 text-lg font-bold mb-1">v0.96 &mdash; April 4, 2026</h2>
            <p className="text-gray-400 mb-3">Living world &mdash; monster spawning, ambient text, wandering, 30+ new emotes, and submit system.</p>

            <div className="space-y-4 mb-8">
              <div>
                <h3 className="text-green-400 font-bold mb-1">Monster System</h3>
                <ul className="text-gray-300 space-y-1 ml-4 list-disc">
                  <li>Monsters spawn from original script data and appear in room descriptions</li>
                  <li>TEX1-4 random ambient text &mdash; monsters emit flavor text on a timer</li>
                  <li>Non-hostile monsters wander between rooms via exits</li>
                  <li>TEXG/TEXE/TEXM text overrides for spawn, entry, and movement</li>
                  <li>Examine monsters to see their descriptions</li>
                  <li>Target monsters with emotes (point skeleton, kick rat, etc.)</li>
                  <li>GM @spawn and @genmon commands actually create monsters</li>
                </ul>
              </div>
              <div>
                <h3 className="text-green-400 font-bold mb-1">New Emotes (30+)</h3>
                <ul className="text-gray-300 space-y-1 ml-4 list-disc">
                  <li>lick, nibble, bark, claw, curse, duck, hiss, hold, hula, jig, moan, massage, and more</li>
                  <li>Self-targeting overrides: spit me, lick me, laugh me, kick me, thump me</li>
                  <li>KISS with body part qualifiers (head, nose, lips, etc.)</li>
                  <li>Submit-gated interactions (kiss lips/navel/feet require target to submit)</li>
                </ul>
              </div>
              <div>
                <h3 className="text-green-400 font-bold mb-1">Submit System</h3>
                <ul className="text-gray-300 space-y-1 ml-4 list-disc">
                  <li>SUBMIT/UNSUBMIT &mdash; accept intimate emotes from other players</li>
                  <li>LICK behavior changes based on submit state</li>
                  <li>Moving to a new room automatically clears submit</li>
                </ul>
              </div>
              <div>
                <h3 className="text-green-400 font-bold mb-1">Fixes</h3>
                <ul className="text-gray-300 space-y-1 ml-4 list-disc">
                  <li>CEVENT script ECHO messages now delivered to players in the room</li>
                  <li>Monster article handling: &ldquo;an orc&rdquo; vs &ldquo;a skeleton&rdquo;</li>
                </ul>
              </div>
            </div>

            <h2 className="text-amber-400 text-lg font-bold mb-1">v0.95 &mdash; April 3, 2026</h2>
            <p className="text-gray-400 mb-3">LEGENDS.DOC fidelity pass &mdash; lock/unlock, ordinal targeting, Mechanoid emote, and verb aliases.</p>

            <div className="space-y-4 mb-8">
              <div>
                <h3 className="text-green-400 font-bold mb-1">Lock &amp; Unlock</h3>
                <ul className="text-gray-300 space-y-1 ml-4 list-disc">
                  <li>LOCK/UNLOCK commands match KEY items via Val3</li>
                  <li>Proper messages for missing keys, wrong keys, already locked/unlocked</li>
                </ul>
              </div>
              <div>
                <h3 className="text-green-400 font-bold mb-1">Ordinal Targeting</h3>
                <ul className="text-gray-300 space-y-1 ml-4 list-disc">
                  <li>&ldquo;2 gate&rdquo;, &ldquo;other gate&rdquo;, &ldquo;second gate&rdquo; target the Nth matching item</li>
                  <li>Works across all 19 item-matching functions (get, drop, look, open, close, etc.)</li>
                </ul>
              </div>
              <div>
                <h3 className="text-green-400 font-bold mb-1">Mechanoid Emote</h3>
                <ul className="text-gray-300 space-y-1 ml-4 list-disc">
                  <li>EMOTE/UNEMOTE is now a Mechanoid racial ability (toggle emotional state)</li>
                  <li>ACT remains the general-purpose roleplaying command</li>
                </ul>
              </div>
              <div>
                <h3 className="text-green-400 font-bold mb-1">Verb Aliases &amp; Commands</h3>
                <ul className="text-gray-300 space-y-1 ml-4 list-disc">
                  <li>ORDER as BUY synonym, UNLIGHT/IGNITE/QUAFF/SHOUT/PLACE aliases</li>
                  <li>RECALL with no args runs room-level IFVERB RECALL scripts</li>
                  <li>ACTBRIEF/RPBRIEF toggle commands</li>
                  <li>POUR verb stub</li>
                </ul>
              </div>
            </div>

            <h2 className="text-amber-400 text-lg font-bold mb-1">v0.94 &mdash; April 3, 2026</h2>
            <p className="text-gray-400 mb-3">Deep script engine, named variables, CEVENT system, food mechanics, and original fidelity.</p>

            <div className="space-y-4 mb-8">
              <div>
                <h3 className="text-green-400 font-bold mb-1">Script Engine</h3>
                <ul className="text-gray-300 space-y-1 ml-4 list-disc">
                  <li>226 named global variables synchronized across servers</li>
                  <li>CEVENT cyclic event system &mdash; timed world events every 3 seconds</li>
                  <li>Arithmetic (MUL/DIV/MOD), monster spawning (GENMON/ZAPMON), persistent PVALs</li>
                  <li>Implicit ENDIF handling matches original engine behavior</li>
                </ul>
              </div>
              <div>
                <h3 className="text-green-400 font-bold mb-1">Food, Drink &amp; Spells</h3>
                <ul className="text-gray-300 space-y-1 ml-4 list-disc">
                  <li>Food tracks bites, drinks track sips &mdash; items consumed over multiple uses</li>
                  <li>Mindlink spell (#403) &mdash; eat a thesnia leaf to gain telepathy for one hour</li>
                  <li>THINK command broadcasts to telepathy-enabled players</li>
                </ul>
              </div>
              <div>
                <h3 className="text-green-400 font-bold mb-1">Original Fidelity</h3>
                <ul className="text-gray-300 space-y-1 ml-4 list-disc">
                  <li>Messages match the original: &ldquo;has just entered the Realms&rdquo;, WHO grid format</li>
                  <li>Body Points (not HP), proper articles (&ldquo;an axe&rdquo;)</li>
                  <li>SIT/LAY/KNEEL trigger room scripts, direction abbreviations resolve for IFPREVERB</li>
                </ul>
              </div>
            </div>

            <h2 className="text-amber-400 text-lg font-bold mb-1">v0.93 &mdash; April 3, 2026</h2>
            <p className="text-gray-400 mb-3">Stealth, flight, combat stubs, session capture, and 70+ new verbs.</p>

            <div className="space-y-4 mb-8">
              <div>
                <h3 className="text-green-400 font-bold mb-1">Stealth &amp; Flight</h3>
                <ul className="text-gray-300 space-y-1 ml-4 list-disc">
                  <li>HIDE to conceal yourself, SNEAK to move while hidden</li>
                  <li>FLY/ASCEND/DESCEND/LAND &mdash; Drakin can always fly, others need spells</li>
                  <li>MARK to set teleport anchors for future spell use</li>
                </ul>
              </div>
              <div>
                <h3 className="text-green-400 font-bold mb-1">Session Capture</h3>
                <ul className="text-gray-300 space-y-1 ml-4 list-disc">
                  <li>Record your gameplay sessions from the Capture button</li>
                  <li>View and download previous captures as .txt files</li>
                </ul>
              </div>
              <div>
                <h3 className="text-green-400 font-bold mb-1">Admin Tools</h3>
                <ul className="text-gray-300 space-y-1 ml-4 list-disc">
                  <li>Real-time Event Monitor for script execution, time cycles, and world state</li>
                  <li>Backend health monitoring with /healthz endpoint</li>
                </ul>
              </div>
              <div>
                <h3 className="text-green-400 font-bold mb-1">Fixes</h3>
                <ul className="text-gray-300 space-y-1 ml-4 list-disc">
                  <li>Proper articles: &ldquo;an axe&rdquo; instead of &ldquo;a axe&rdquo;</li>
                  <li>GO command works for non-portal items with scripts (stairways, etc.)</li>
                  <li>Text is now selectable/copyable in the terminal</li>
                  <li>HP renamed to BP (Body Points) throughout</li>
                </ul>
              </div>
            </div>

            <h2 className="text-amber-400 text-lg font-bold mb-1">v0.91 &mdash; April 3, 2026</h2>
            <p className="text-gray-400 mb-3">Major systems expansion &mdash; script engine, world systems, and player features.</p>

            <div className="space-y-4 mb-8">
              <div>
                <h3 className="text-green-400 font-bold mb-1">Script Engine</h3>
                <ul className="text-gray-300 space-y-1 ml-4 list-disc">
                  <li>80+ script variables for conditions (stats, resources, room info, time, weather, flags)</li>
                  <li>IFSAY blocks &mdash; NPCs and objects respond to what you say</li>
                  <li>AFFECT for multi-room script effects</li>
                  <li>Environmental damage, random events, forced positioning</li>
                  <li>Full string substitution: pronouns, item names, newlines, converted numbers</li>
                </ul>
              </div>
              <div>
                <h3 className="text-green-400 font-bold mb-1">World Systems</h3>
                <ul className="text-gray-300 space-y-1 ml-4 list-disc">
                  <li>In-game clock and calendar with day/night cycle</li>
                  <li>Monsters spawn in rooms and appear in descriptions</li>
                  <li>Weather system with 15 states shown in outdoor rooms</li>
                  <li>Dark rooms require light sources to see</li>
                </ul>
              </div>
              <div>
                <h3 className="text-green-400 font-bold mb-1">New Commands</h3>
                <ul className="text-gray-300 space-y-1 ml-4 list-disc">
                  <li>Drink, light/extinguish, flip, latch/unlatch</li>
                  <li>Banking (deposit/withdraw in bank rooms)</li>
                  <li>Skill training with 36 named skills</li>
                  <li>150+ spells registered across 5 schools (casting coming soon)</li>
                  <li>Mining, foraging, and crafting commands (stubs for now)</li>
                </ul>
              </div>
            </div>

            <h2 className="text-amber-400 text-lg font-bold mb-1">v0.9 &mdash; April 3, 2026</h2>
            <p className="text-gray-400 mb-3">First public release of Legends of Future Past, resurrected from the original 1990s script files.</p>

            <div className="space-y-4">
              <div>
                <h3 className="text-green-400 font-bold mb-1">Explore the World</h3>
                <ul className="text-gray-300 space-y-1 ml-4 list-disc">
                  <li>Over 2,200 rooms to explore across the Shattered Realms</li>
                  <li>Nearly 2,000 items and 300 monsters parsed from the original game scripts</li>
                  <li>Move in all compass directions, climb portals, go through gates and doors</li>
                  <li>Look directionally to see what lies ahead before moving</li>
                  <li>Rich room descriptions with original formatted text preserved (poems, maps, ASCII art)</li>
                  <li>Examine items in rooms with scripted descriptions</li>
                  <li>Read signs, plaques, manuscripts, and scrolls</li>
                </ul>
              </div>

              <div>
                <h3 className="text-green-400 font-bold mb-1">Interact with Everything</h3>
                <ul className="text-gray-300 space-y-1 ml-4 list-disc">
                  <li>Pull, push, turn, rub, tap, touch, search, and dig</li>
                  <li>Buy weapons, armor, and supplies from shops (1,400+ items for sale)</li>
                  <li>Sell items at appropriate merchants</li>
                  <li>Open, close, lock and unlock doors and containers</li>
                  <li>Look inside, on top of, under, and behind objects</li>
                  <li>Recall lore about items based on your knowledge skill</li>
                  <li>Script-driven puzzles and interactions throughout the world</li>
                </ul>
              </div>

              <div>
                <h3 className="text-green-400 font-bold mb-1">Roleplaying & Communication</h3>
                <ul className="text-gray-300 space-y-1 ml-4 list-disc">
                  <li>60+ social emotes: smile, bow, kick, hug, dance, and many more</li>
                  <li>Targeted emotes show second-person messages to the target</li>
                  <li>Say, whisper, yell, recite, and custom emote commands</li>
                  <li>See other players in rooms with position descriptions</li>
                  <li>WHO list shows all online players</li>
                </ul>
              </div>

              <div>
                <h3 className="text-green-400 font-bold mb-1">Characters</h3>
                <ul className="text-gray-300 space-y-1 ml-4 list-disc">
                  <li>8 races: Human, Aelfen, Highlander, Wolfling, Murg, Drakin, Mechanoid, Ephemeral</li>
                  <li>Stat rolling based on racial ranges</li>
                  <li>Starting gear and newbie guidance at the City Gate</li>
                  <li>Persistent inventory, equipment, skills, and internal variables</li>
                  <li>Multiple characters per account</li>
                </ul>
              </div>

              <div>
                <h3 className="text-green-400 font-bold mb-1">Multiplayer</h3>
                <ul className="text-gray-300 space-y-1 ml-4 list-disc">
                  <li>Real-time WebSocket gameplay</li>
                  <li>Cross-server coordination &mdash; all players share the same world</li>
                  <li>Automatic reconnection if connection is lost</li>
                  <li>Google sign-in with 30-day session persistence</li>
                </ul>
              </div>

              <div>
                <h3 className="text-yellow-400 font-bold mb-1">Coming Soon</h3>
                <ul className="text-gray-400 space-y-1 ml-4 list-disc">
                  <li>Combat system</li>
                  <li>Magic and psionics (spells, casting, concentration)</li>
                  <li>NPC/monster AI and spawning</li>
                  <li>Crafting, mining, and foraging</li>
                  <li>Tutorial room sequence</li>
                  <li>Seasonal world variations</li>
                  <li>Level progression and experience</li>
                </ul>
              </div>
            </div>
          </section>
        </div>

        <div className="mt-8 pt-4 border-t border-[#333] text-gray-600 text-xs text-center">
          Legends of Future Past &mdash; Originally created in the 1990s, resurrected from original script files.
        </div>
      </div>
    </div>
  )
}
