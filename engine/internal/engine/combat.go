package engine

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/jonradoff/lofp/internal/gameworld"
)

// Combat stance constants
const (
	StanceNormal    = 0
	StanceOffensive = 1
	StanceDefensive = 2
	StanceBerserk   = 3
	StanceWary      = 4
)

var stanceNames = map[int]string{
	StanceNormal:    "normal",
	StanceOffensive: "offensive",
	StanceDefensive: "defensive",
	StanceBerserk:   "berserk",
	StanceWary:      "wary",
}

// CombatTarget tracks who a player/monster is fighting.
type CombatTarget struct {
	IsMonster  bool
	MonsterID  int
	PlayerName string
}

// ---- Damage severity tiers (from original session capture) ----

var severityTiers = []struct {
	maxDmg int
	name   string
}{
	{5, "Puny"},
	{10, "Grazing"},
	{15, "Insignificant"},
	{20, "Minor"},
	{25, "Passable"},
	{30, "Good"},
	{40, "Well-aimed"},
	{50, "Masterful"},
	{60, "Grisly"},
	{75, "Severe"},
	{100, "Ghastly"},
	{99999, "Dazzling explosive"},
}

func damageSeverity(dmg int) string {
	for _, tier := range severityTiers {
		if dmg <= tier.maxDmg {
			return tier.name
		}
	}
	return "Devastating"
}

// ---- Attack verb by weapon type (from session capture) ----

func attackVerb(weaponDef *gameworld.ItemDef) (selfVerb, thirdVerb, dmgNoun string) {
	if weaponDef == nil {
		return "swing", "swings", "strike"
	}
	switch weaponDef.Type {
	case "SLASH_WEAPON", "TWOHAND_WEAPON", "CLAW_WEAPON", "DRAKIN_SLASH":
		return "swing", "swings", "slash"
	case "PUNCTURE_WEAPON", "POLE_WEAPON", "POLETHROWN", "DRAKIN_POLE", "STABTHROWN":
		return "thrust", "thrusts", "strike"
	case "CRUSH_WEAPON", "BLUNT_WEAPON", "DRAKIN_CRUSH":
		return "swing", "swings", "strike"
	case "BOW_WEAPON", "THROWN_WEAPON", "DRAKIN_THROWN":
		return "fire", "fires", "strike"
	case "BITE_WEAPON":
		return "bite", "bites", "bite"
	default:
		return "swing", "swings", "strike"
	}
}

// monsterAttackVerb returns verb for a monster's attack based on its weapon/type.
func monsterAttackVerb(def *gameworld.MonsterDef, items map[int]*gameworld.ItemDef) (verb, dmgNoun string) {
	if len(def.Weapons) > 0 {
		wep := items[def.Weapons[0].Archetype]
		if wep != nil {
			_, v, dn := attackVerb(wep)
			return v + " at", dn
		}
	}
	// Default based on body type
	if def.BodyType == "ANIMAL" || def.BodyType == "AVINE" {
		return "slashes at", "slash"
	}
	return "swings at", "strike"
}

// ---- Weapon skill mapping ----

func weaponSkillForType(itemType string) int {
	switch itemType {
	case "SLASH_WEAPON", "TWOHAND_WEAPON":
		return 13
	case "CRUSH_WEAPON", "BLUNT_WEAPON":
		return 9
	case "PUNCTURE_WEAPON", "STABTHROWN":
		return 13
	case "POLE_WEAPON", "POLETHROWN":
		return 25
	case "BOW_WEAPON", "THROWN_WEAPON":
		return 3
	case "DRAKIN_SLASH", "DRAKIN_CRUSH", "DRAKIN_POLE", "DRAKIN_THROWN":
		return 16
	case "CLAW_WEAPON", "BITE_WEAPON":
		return 4
	default:
		return 13
	}
}

// ---- To-Hit Calculation (original: ToHit = minimum d100 roll needed) ----

// calcToHit returns the minimum d100 roll needed to hit.
// Low ToHit = easy target, high ToHit = hard target.
// From the capture: player with good skill vs ursine → ToHit: 5 (easy)
//                   ursine vs well-equipped player → ToHit: 95 (hard)
func calcToHit(attackRating, defenseRating int) int {
	// ToHit = 50 + defense - attack, clamped to [5, 95]
	toHit := 50 + defenseRating - attackRating
	if toHit < 5 {
		toHit = 5
	}
	if toHit > 95 {
		toHit = 95
	}
	return toHit
}

func playerAttackRating(player *Player, weaponDef *gameworld.ItemDef) int {
	rating := 50
	rating += player.Level * 3
	if weaponDef != nil {
		skillID := weaponSkillForType(weaponDef.Type)
		rating += player.Skills[skillID] * 4
	}
	if weaponDef != nil && (weaponDef.Type == "BOW_WEAPON" || weaponDef.Type == "THROWN_WEAPON") {
		rating += player.Agility / 5
	} else {
		rating += player.Strength / 5
	}
	switch player.Stance {
	case StanceOffensive:
		rating += 15
	case StanceDefensive:
		rating -= 15
	case StanceBerserk:
		rating += 25
	case StanceWary:
		rating -= 5
	}
	switch player.Position {
	case 1:
		rating -= 20
	case 2:
		rating -= 30
	case 3:
		rating -= 10
	}
	return rating
}

func playerDefenseRating(player *Player) int {
	rating := 25
	rating += player.Level * 3
	rating += player.Skills[6] * 4
	rating += player.Agility / 5
	switch player.Stance {
	case StanceOffensive:
		rating -= 15
	case StanceDefensive:
		rating += 15
	case StanceBerserk:
		rating -= 25
	case StanceWary:
		rating += 5
	}
	rating += player.DefenseBonus
	switch player.Position {
	case 1:
		rating -= 15
	case 2:
		rating -= 25
	case 3:
		rating -= 10
	}
	return rating
}

func playerArmorPercent(player *Player, items map[int]*gameworld.ItemDef) int {
	total := 0
	for _, worn := range player.Worn {
		def := items[worn.Archetype]
		if def != nil && def.Type == "ARMOR" {
			total += def.Parameter1
		}
	}
	if total > 85 {
		total = 85
	}
	return total
}

// ---- Damage Calculation ----

func playerDamage(player *Player, weaponDef *gameworld.ItemDef) int {
	if weaponDef == nil {
		return rand.Intn(3) + 1 + player.Strength/20
	}
	maxDmg := weaponDef.Parameter1
	if maxDmg <= 0 {
		maxDmg = 3
	}
	dmg := rand.Intn(maxDmg) + 1
	if weaponDef.Type == "BOW_WEAPON" || weaponDef.Type == "THROWN_WEAPON" {
		dmg += player.Agility / 10
	} else {
		dmg += player.Strength / 10
	}
	if player.Stance == StanceBerserk {
		dmg = dmg * 12 / 10
	}
	return dmg
}

func monsterDamage(def *gameworld.MonsterDef) int {
	minDmg := max(1, def.Attack1/20)
	maxDmg := max(2, def.Attack1/5)
	if maxDmg <= minDmg {
		maxDmg = minDmg + 1
	}
	return rand.Intn(maxDmg-minDmg+1) + minDmg
}

func monsterSpecialDamage(def *gameworld.MonsterDef) (int, string) {
	if def.SpecUse <= 0 || rand.Intn(100) >= def.SpecUse {
		return 0, ""
	}
	dmg := def.SpecBase + rand.Intn(max(1, def.SpecDmg))
	return dmg, def.SpecDmgType
}

func applyArmor(dmg, armorPct int) int {
	reduced := dmg * (100 - armorPct) / 100
	if reduced < 0 {
		reduced = 0
	}
	return reduced
}

func applyImmunity(dmg int, immunityLevel int) int {
	switch immunityLevel {
	case 0:
		return 0
	case 1:
		return dmg / 2
	case 3:
		return dmg * 3 / 2
	case 4:
		return dmg * 2
	default:
		return dmg
	}
}

// ---- Body parts (from original session capture) ----

var bodyParts = []string{"head", "body", "right arm", "left arm", "right leg", "left leg", "back"}
var animalParts = []string{"head", "body", "right foreleg", "left foreleg", "right hind leg", "left hind leg", "tail"}

func randomBodyPart(bodyType string) string {
	if bodyType == "ANIMAL" || bodyType == "AVINE" {
		return animalParts[rand.Intn(len(animalParts))]
	}
	return bodyParts[rand.Intn(len(bodyParts))]
}

// ---- Weapon helpers ----

func weaponImmunityType(weaponDef *gameworld.ItemDef) int {
	if weaponDef == nil {
		return 1
	}
	switch weaponDef.Type {
	case "CRUSH_WEAPON", "BLUNT_WEAPON":
		return 1
	default:
		return 2
	}
}

func (e *GameEngine) weaponDisplayName(player *Player, weaponDef *gameworld.ItemDef) string {
	if weaponDef == nil {
		return "fist"
	}
	if player.Wielded != nil {
		return e.formatItemName(weaponDef, player.Wielded.Adj1, player.Wielded.Adj2, player.Wielded.Adj3)
	}
	return strings.ToLower(e.nouns[weaponDef.NameID])
}

func (e *GameEngine) monsterWeaponName(def *gameworld.MonsterDef) string {
	if len(def.Weapons) > 0 {
		wep := e.items[def.Weapons[0].Archetype]
		if wep != nil {
			adj := ""
			if def.Weapons[0].Adj > 0 {
				if a, ok := e.adjectives[def.Weapons[0].Adj]; ok {
					adj = a + " "
				}
			}
			return adj + strings.ToLower(e.nouns[wep.NameID])
		}
	}
	return "claws"
}

// ---- Player attacks Monster ----

func (e *GameEngine) doAttackMonster(player *Player, target string) *CommandResult {
	if player.Dead {
		return &CommandResult{Messages: []string{"You can't do that. You are dead."}}
	}
	if player.Stunned {
		return &CommandResult{Messages: []string{"You are stunned and cannot attack!"}}
	}
	if player.Immobilized {
		return &CommandResult{Messages: []string{"You are rooted to the spot!"}}
	}
	if player.Position == 2 {
		return &CommandResult{Messages: []string{"You can't attack while laying down! Stand up first."}}
	}

	if player.RoundTimeExpiry.After(time.Now()) {
		remaining := int(player.RoundTimeExpiry.Sub(time.Now()).Seconds()) + 1
		return &CommandResult{Messages: []string{fmt.Sprintf("[Wait %d seconds...]", remaining)}}
	}

	inst, def := e.findMonsterInRoom(player, target)
	if inst == nil {
		return &CommandResult{Messages: []string{fmt.Sprintf("You don't see '%s' here to attack.", target)}}
	}

	name := FormatMonsterName(def, e.monAdjs)
	article := articleFor(name, def.Unique)

	var weaponDef *gameworld.ItemDef
	if player.Wielded != nil {
		weaponDef = e.items[player.Wielded.Archetype]
	}

	// Engage
	player.CombatTarget = &CombatTarget{IsMonster: true, MonsterID: inst.ID}
	player.Joined = true
	e.monsterMgr.mu.Lock()
	for i := range e.monsterMgr.instances {
		if e.monsterMgr.instances[i].ID == inst.ID {
			if e.monsterMgr.instances[i].Target == "" {
				e.monsterMgr.instances[i].Target = player.FirstName
			}
			break
		}
	}
	e.monsterMgr.mu.Unlock()

	// Resolve to-hit (original format: [ToHit: X, Roll: Y])
	attackRating := playerAttackRating(player, weaponDef)
	toHit := calcToHit(attackRating, def.Defense)
	roll := rand.Intn(100) + 1 // 1-100

	selfVerb, thirdVerb, dmgNoun := attackVerb(weaponDef)
	weaponName := e.weaponDisplayName(player, weaponDef)
	pronoun := player.Possessive()

	result := &CommandResult{}
	var msgs []string

	// Swing line: "You swing at an ursine with your ice longsword."
	msgs = append(msgs, fmt.Sprintf("You %s at %s%s with your %s.", selfVerb, article, name, weaponName))

	if roll >= toHit {
		// Hit!
		excellent := roll >= 96
		hitLabel := "Hit!"
		if excellent {
			hitLabel = "Excellent Hit!"
		}
		msgs = append(msgs, fmt.Sprintf(" [ToHit: %d, Roll: %d] %s", toHit, roll, hitLabel))

		dmg := playerDamage(player, weaponDef)
		dmg = applyArmor(dmg, def.Armor)
		immType := weaponImmunityType(weaponDef)
		if level, ok := def.Immunities[immType]; ok {
			dmg = applyImmunity(dmg, level)
		}
		if dmg <= 0 {
			dmg = 1
		}

		part := randomBodyPart(def.BodyType)
		severity := damageSeverity(dmg)
		msgs = append(msgs, fmt.Sprintf(" %s %s to %s. [%d Damage]", severity, dmgNoun, part, dmg))

		killed := e.damageMonster(inst.ID, dmg)

		if excellent && !killed {
			// Stun chance on excellent hits
			if rand.Intn(100) < 30 {
				msgs = append(msgs, " It is stunned!")
			}
		}

		if killed {
			deathText := def.TextOverrides["TEXD"]
			if deathText != "" {
				msgs = append(msgs, fmt.Sprintf(" It %s", deathText))
			} else {
				msgs = append(msgs, " It collapses, dead.")
			}
			e.handleMonsterDeath(player, inst, def)
			player.CombatTarget = nil
			player.Joined = false
		}

		result.RoomBroadcast = []string{fmt.Sprintf("%s %s at %s%s with %s %s.", player.FirstName, thirdVerb, article, name, pronoun, weaponName)}
	} else {
		msgs = append(msgs, fmt.Sprintf(" [ToHit: %d, Roll: %d] Miss.", toHit, roll))
		result.RoomBroadcast = []string{fmt.Sprintf("%s %s at %s%s but misses.", player.FirstName, thirdVerb, article, name)}
	}

	result.Messages = msgs

	// Roundtime
	rtSeconds := 5
	if player.Quickness > 80 {
		rtSeconds = 3
	} else if player.Quickness > 50 {
		rtSeconds = 4
	}
	if player.Stance == StanceBerserk {
		rtSeconds--
	}
	if rtSeconds < 2 {
		rtSeconds = 2
	}
	player.RoundTimeExpiry = time.Now().Add(time.Duration(rtSeconds) * time.Second)
	result.Messages = append(result.Messages, fmt.Sprintf("[Round: %d sec]", rtSeconds))

	return result
}

// ---- Monster attacks Player ----

func (e *GameEngine) monsterAttackPlayer(inst *MonsterInstance, def *gameworld.MonsterDef, player *Player) (playerMsgs []string, roomMsgs []string) {
	if player.Dead || !inst.Alive {
		return nil, nil
	}

	name := FormatMonsterName(def, e.monAdjs)
	article := articleFor(name, def.Unique)

	// Special attack first
	if specDmg, specType := monsterSpecialDamage(def); specDmg > 0 {
		specText := def.TextOverrides["TEXX"]
		if specText != "" {
			specText = strings.Replace(specText, "%s", article+name, 1)
			specText = strings.Replace(specText, "%s", player.FirstName, 1)
		} else {
			specText = fmt.Sprintf("%s%s uses a special attack on %s!", strings.Title(article), name, player.FirstName)
		}

		armorPct := playerArmorPercent(player, e.items)
		specDmg = applyArmor(specDmg, armorPct)
		_ = specType

		part := randomBodyPart("HUMAN")
		severity := damageSeverity(specDmg)
		player.BodyPoints -= specDmg
		if player.BodyPoints < 0 {
			player.BodyPoints = 0
		}

		playerMsgs = append(playerMsgs, specText)
		playerMsgs = append(playerMsgs, fmt.Sprintf(" %s burn to %s. [%d Damage]", severity, part, specDmg))
		roomMsgs = append(roomMsgs, specText)

		if player.BodyPoints <= 0 {
			deathMsgs := e.handlePlayerDeath(player, name)
			playerMsgs = append(playerMsgs, deathMsgs...)
			return playerMsgs, roomMsgs
		}
	}

	// Normal attack
	monWeaponName := e.monsterWeaponName(def)
	monVerb, monDmgNoun := monsterAttackVerb(def, e.items)

	playerMsgs = append(playerMsgs, fmt.Sprintf("%s%s %s %s with its %s.", strings.ToUpper(article[:1])+article[1:], name, monVerb, player.FirstName, monWeaponName))

	toHit := calcToHit(def.Attack1, playerDefenseRating(player))
	roll := rand.Intn(100) + 1

	if roll >= toHit {
		excellent := roll >= 96
		hitLabel := "Hit!"
		if excellent {
			hitLabel = "Excellent Hit!"
		}
		playerMsgs = append(playerMsgs, fmt.Sprintf(" [ToHit: %d, Roll: %d] %s", toHit, roll, hitLabel))

		dmg := monsterDamage(def)
		armorPct := playerArmorPercent(player, e.items)
		dmg = applyArmor(dmg, armorPct)
		if dmg <= 0 {
			dmg = 1
		}

		part := randomBodyPart("HUMAN")
		severity := damageSeverity(dmg)
		player.BodyPoints -= dmg
		if player.BodyPoints < 0 {
			player.BodyPoints = 0
		}

		playerMsgs = append(playerMsgs, fmt.Sprintf(" %s %s to %s. [%d Damage]", severity, monDmgNoun, part, dmg))
		roomMsgs = append(roomMsgs, fmt.Sprintf("%s%s attacks %s!", strings.ToUpper(article[:1])+article[1:], name, player.FirstName))

		if player.BodyPoints <= 0 {
			playerMsgs = append(playerMsgs, fmt.Sprintf(" %s%s slays %s.", strings.ToUpper(article[:1])+article[1:], name, player.FirstName))
			deathMsgs := e.handlePlayerDeath(player, name)
			playerMsgs = append(playerMsgs, deathMsgs...)
			roomMsgs = append(roomMsgs, fmt.Sprintf("%s%s slays %s!", strings.ToUpper(article[:1])+article[1:], name, player.FirstName))
		}
	} else {
		playerMsgs = append(playerMsgs, fmt.Sprintf(" [ToHit: %d, Roll: %d] Miss.", toHit, roll))
	}

	return playerMsgs, roomMsgs
}

// ---- Death ----

func (e *GameEngine) handlePlayerDeath(player *Player, killerName string) []string {
	player.Dead = true
	player.CombatTarget = nil
	player.Joined = false
	player.Position = 2

	e.Events.Publish("combat", fmt.Sprintf("%s was killed by %s in room %d", player.FirstName, killerName, player.RoomNumber))

	return []string{
		"",
		"You are dead and can't do much of anything beside wait for someone to attempt to raise you or for Eternity, Inc. to retrieve you. Hope you paid your premium! [You may type DEPART at any time to allow Eternity, Inc. to retrieve you.]",
	}
}

func (e *GameEngine) doDepart(player *Player) *CommandResult {
	if !player.Dead {
		return &CommandResult{Messages: []string{"You are not dead."}}
	}

	player.Dead = false
	player.Position = 0
	player.Bleeding = false
	player.Stunned = false
	player.Poisoned = false
	player.Diseased = false

	player.BodyPoints = player.MaxBodyPoints / 4
	if player.BodyPoints < 1 {
		player.BodyPoints = 1
	}

	player.RoomNumber = e.startRoom

	result := e.doLook(player)
	result.Messages = append([]string{
		"Your spirit coalesces and you feel the sensation of being pulled back into the world...",
		"",
	}, result.Messages...)
	result.RoomBroadcast = []string{fmt.Sprintf("%s's spirit has returned from Eternity.", player.FirstName)}

	return result
}

// damageMonster applies damage to a monster instance. Returns true if killed.
func (e *GameEngine) damageMonster(monsterID int, dmg int) bool {
	e.monsterMgr.mu.Lock()
	defer e.monsterMgr.mu.Unlock()
	for i := range e.monsterMgr.instances {
		if e.monsterMgr.instances[i].ID == monsterID && e.monsterMgr.instances[i].Alive {
			e.monsterMgr.instances[i].CurrentHP -= dmg
			if e.monsterMgr.instances[i].CurrentHP <= 0 {
				e.monsterMgr.instances[i].Alive = false
				e.monsterMgr.instances[i].CurrentHP = 0
				return true
			}
			return false
		}
	}
	return false
}

// ---- Monster Death ----

func (e *GameEngine) handleMonsterDeath(killer *Player, inst *MonsterInstance, def *gameworld.MonsterDef) {
	xp := def.Body + def.Attack1/5 + def.Defense/5
	if xp < 5 {
		xp = 5
	}
	killer.Experience += xp
	e.Events.Publish("combat", fmt.Sprintf("%s killed %s (monster %d) for %d XP in room %d",
		killer.FirstName, def.Name, def.Number, xp, killer.RoomNumber))

	nextLevelXP := killer.Level * 1000
	if nextLevelXP <= 0 {
		nextLevelXP = 1000
	}
	if killer.Experience >= nextLevelXP {
		killer.Level++
		killer.MaxBodyPoints += killer.Constitution / 10
		killer.BodyPoints = killer.MaxBodyPoints
		killer.MaxFatigue += killer.Constitution / 15
		killer.Fatigue = killer.MaxFatigue
		if e.roomBroadcast != nil {
			e.roomBroadcast(killer.RoomNumber, []string{
				fmt.Sprintf("%s has advanced to level %d!", killer.FirstName, killer.Level),
			})
		}
		if e.sendToPlayer != nil {
			e.sendToPlayer(killer.FirstName, []string{
				fmt.Sprintf("Congratulations! You have advanced to level %d! (+%d max BP)", killer.Level, killer.Constitution/10),
			})
		}
	}

	// Don't remove dead monster from room tracking — it shows as "(dead)" in LOOK
	// Just mark as not alive (already done by damageMonster)
}

// ---- Flee ----

func (e *GameEngine) doFlee(player *Player) *CommandResult {
	if player.CombatTarget == nil && !player.Joined {
		return &CommandResult{Messages: []string{"You are not in combat."}}
	}
	if player.Dead {
		return &CommandResult{Messages: []string{"You can't flee. You are dead."}}
	}
	if player.Immobilized {
		return &CommandResult{Messages: []string{"You are rooted to the spot!"}}
	}

	room := e.rooms[player.RoomNumber]
	if room == nil {
		return &CommandResult{Messages: []string{"You have nowhere to flee!"}}
	}

	type exitInfo struct {
		dir    string
		destID int
	}
	var exits []exitInfo
	for dir, dest := range room.Exits {
		if dest > 0 {
			exits = append(exits, exitInfo{dir, dest})
		}
	}
	if len(exits) == 0 {
		return &CommandResult{Messages: []string{"There is nowhere to flee!"}}
	}

	fleeChance := 50 + player.Quickness/5 + player.Agility/10
	if player.Position != 0 {
		fleeChance -= 20
	}
	if rand.Intn(100) >= fleeChance {
		return &CommandResult{
			Messages:      []string{"You try to flee but can't get away!"},
			RoomBroadcast: []string{fmt.Sprintf("%s tries to flee but fails!", player.FirstName)},
		}
	}

	chosen := exits[rand.Intn(len(exits))]
	e.disengageCombat(player)

	dirName := directionNames[chosen.dir]
	if dirName == "" {
		dirName = strings.ToLower(chosen.dir)
	}

	oldRoom := player.RoomNumber
	player.RoomNumber = chosen.destID
	player.Position = 0
	player.Submitting = false

	result := e.doLook(player)
	result.Messages = append([]string{fmt.Sprintf("You flee %s!", dirName)}, result.Messages...)
	result.OldRoom = oldRoom
	result.OldRoomMsg = []string{fmt.Sprintf("%s flees %s!", player.FirstName, dirName)}
	result.RoomBroadcast = []string{fmt.Sprintf("%s arrives, looking panicked.", player.FirstName)}

	return result
}

func (e *GameEngine) disengageCombat(player *Player) {
	if player.CombatTarget != nil && player.CombatTarget.IsMonster {
		e.monsterMgr.mu.Lock()
		for i := range e.monsterMgr.instances {
			if e.monsterMgr.instances[i].ID == player.CombatTarget.MonsterID {
				if e.monsterMgr.instances[i].Target == player.FirstName {
					e.monsterMgr.instances[i].Target = ""
				}
				break
			}
		}
		e.monsterMgr.mu.Unlock()
	}
	player.CombatTarget = nil
	player.Joined = false
}

// ---- Stances ----

func (e *GameEngine) doStance(player *Player, stance int) *CommandResult {
	if stance == StanceBerserk && player.Race != RaceMurg {
		return &CommandResult{Messages: []string{"Only Murg can enter a berserk frenzy."}}
	}
	player.Stance = stance
	return &CommandResult{
		Messages:      []string{fmt.Sprintf("You adopt a %s combat stance.", stanceNames[stance])},
		RoomBroadcast: []string{fmt.Sprintf("%s adopts a %s combat stance.", player.FirstName, stanceNames[stance])},
	}
}

// ---- Monster Combat AI ----

func (e *GameEngine) monsterCombatTick(inst *MonsterInstance, def *gameworld.MonsterDef) {
	if inst.Target == "" || !inst.Alive {
		return
	}

	if e.sessions == nil {
		return
	}
	var target *Player
	for _, p := range e.sessions.OnlinePlayers() {
		if p.FirstName == inst.Target && p.RoomNumber == inst.RoomNumber && !p.Dead {
			target = p
			break
		}
	}

	if target == nil {
		inst.Target = ""
		return
	}

	e.monsterMgr.mu.Unlock()
	playerMsgs, roomMsgs := e.monsterAttackPlayer(inst, def, target)
	e.monsterMgr.mu.Lock()

	if e.sendToPlayer != nil && len(playerMsgs) > 0 {
		e.sendToPlayer(target.FirstName, playerMsgs)
	}
	if e.roomBroadcast != nil && len(roomMsgs) > 0 {
		e.roomBroadcast(inst.RoomNumber, roomMsgs)
	}

	// Monster flee behavior
	if inst.Alive && inst.CurrentHP > 0 {
		hpPct := inst.CurrentHP * 100 / max(1, def.Body)
		shouldFlee := false
		switch {
		case def.Strategy >= 301 && def.Strategy < 500:
			shouldFlee = hpPct < 30
		case def.Strategy >= 201 && def.Strategy < 300:
			shouldFlee = hpPct < 50
		case def.Strategy >= 1 && def.Strategy < 200:
			shouldFlee = hpPct < 60
		}
		if shouldFlee {
			e.monsterFlee(inst, def)
		}
	}
}

func (e *GameEngine) monsterFlee(inst *MonsterInstance, def *gameworld.MonsterDef) {
	room := e.rooms[inst.RoomNumber]
	if room == nil {
		return
	}
	type exitInfo struct {
		dir    string
		destID int
	}
	var exits []exitInfo
	for dir, dest := range room.Exits {
		if dest > 0 {
			exits = append(exits, exitInfo{dir, dest})
		}
	}
	if len(exits) == 0 {
		return
	}
	chosen := exits[rand.Intn(len(exits))]
	name := FormatMonsterName(def, e.monAdjs)
	dirName := directionNames[chosen.dir]
	if dirName == "" {
		dirName = strings.ToLower(chosen.dir)
	}

	fleeText := def.TextOverrides["TEXF"]
	if fleeText != "" {
		e.roomBroadcast(inst.RoomNumber, []string{fleeText + " " + dirName + "."})
	} else {
		article := articleFor(name, def.Unique)
		e.roomBroadcast(inst.RoomNumber, []string{fmt.Sprintf("%s%s flees %s!", strings.ToUpper(article[:1])+article[1:], name, dirName)})
	}

	inst.Target = ""
	e.monsterMgr.moveMonster(e.monsterMgr.indexOfID(inst.ID), chosen.destID)
}

func (e *GameEngine) monsterCheckAggro(player *Player, roomNum int) {
	if e.monsterMgr == nil || player.Dead || player.Hidden || player.GMInvis {
		return
	}

	e.monsterMgr.mu.Lock()
	defer e.monsterMgr.mu.Unlock()

	for i := range e.monsterMgr.instances {
		inst := &e.monsterMgr.instances[i]
		if inst.RoomNumber != roomNum || !inst.Alive || inst.Sedated || inst.Target != "" {
			continue
		}
		def := e.monsters[inst.DefNumber]
		if def == nil || def.Strategy < 301 {
			continue
		}
		inst.Target = player.FirstName
		name := FormatMonsterName(def, e.monAdjs)
		article := articleFor(name, def.Unique)
		if e.sendToPlayer != nil {
			e.sendToPlayer(player.FirstName, []string{fmt.Sprintf("%s%s stands erect and closes with you.", strings.ToUpper(article[:1])+article[1:], name)})
		}
		break
	}
}

// ---- Helpers ----

func articleFor(name string, unique bool) string {
	if unique {
		return ""
	}
	if len(name) > 0 && strings.ContainsRune("aeiouAEIOU", rune(name[0])) {
		return "an "
	}
	return "a "
}

func (mm *monsterManager) indexOfID(id int) int {
	for i := range mm.instances {
		if mm.instances[i].ID == id {
			return i
		}
	}
	return -1
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
