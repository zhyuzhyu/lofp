package engine

import (
	"math/rand"
	"time"
)

// WeatherNames maps weather state IDs to display names (from GM Manual).
var WeatherNames = map[int]string{
	0:  "Sunny",
	1:  "Partly Cloudy",
	2:  "Overcast",
	3:  "Light Rain",
	4:  "Moderate Rain",
	5:  "Heavy Rain",
	6:  "Thunderstorm",
	7:  "Gale",
	8:  "Hurricane",
	9:  "Hail",
	10: "Sleet",
	11: "Snow Flurries",
	12: "Moderate Snow",
	13: "Heavy Snow",
	14: "Blizzard",
}

// Weather transition messages broadcast to outdoor players.
var weatherTransitionMessages = map[[2]int]string{
	{0, 1}:  "A few clouds drift across the sky.",
	{0, 2}:  "The sky darkens and becomes overcast.",
	{1, 0}:  "The clouds overhead drift away, leaving only clear skies.",
	{1, 2}:  "The sky darkens as clouds gather overhead.",
	{2, 0}:  "The clouds break apart, revealing clear blue sky.",
	{2, 1}:  "The overcast sky begins to lighten as gaps appear in the clouds.",
	{2, 3}:  "A light rain begins to fall.",
	{3, 2}:  "The rain tapers off and the clouds begin to thin.",
	{3, 4}:  "The rain intensifies to a steady downpour.",
	{4, 3}:  "The rain lessens to a light drizzle.",
	{4, 5}:  "The rain grows heavier, driven by gusting wind.",
	{5, 4}:  "The downpour eases somewhat.",
	{5, 6}:  "Thunder rumbles in the distance as lightning flashes across the sky.",
	{6, 5}:  "The thunder fades, though heavy rain continues.",
	{6, 2}:  "The storm passes, leaving only overcast skies.",
	{2, 11}: "Snowflakes begin to fall gently from the gray sky.",
	{11, 2}: "The snow stops and the clouds begin to thin.",
	{11, 12}: "The snow picks up, falling more heavily now.",
	{12, 11}: "The snowfall lightens to scattered flurries.",
	{12, 13}: "It is snowing heavily.",
	{13, 12}: "The heavy snow begins to let up.",
	{13, 14}: "A howling blizzard descends, reducing visibility to almost nothing.",
	{14, 13}: "The blizzard eases to heavy snow.",
}

// weatherTransitions defines likely next states from each weather state.
// Each entry is [nextState, weight]. Higher weight = more likely.
var weatherTransitions = map[int][][2]int{
	0:  {{0, 60}, {1, 30}, {2, 10}},                // Sunny → stay, partly cloudy, overcast
	1:  {{0, 30}, {1, 40}, {2, 30}},                // Partly Cloudy → clear, stay, overcast
	2:  {{1, 20}, {2, 40}, {3, 25}, {11, 15}},      // Overcast → clearing, stay, rain, snow
	3:  {{2, 25}, {3, 40}, {4, 35}},                // Light Rain → clear, stay, moderate
	4:  {{3, 30}, {4, 35}, {5, 35}},                // Moderate Rain → lighter, stay, heavier
	5:  {{4, 35}, {5, 30}, {6, 35}},                // Heavy Rain → moderate, stay, storm
	6:  {{5, 40}, {2, 40}, {6, 20}},                // Thunderstorm → heavy rain, overcast, stay
	7:  {{5, 50}, {6, 30}, {7, 20}},                // Gale → heavy rain, storm, stay
	8:  {{7, 50}, {5, 30}, {8, 20}},                // Hurricane → gale, heavy rain, stay
	9:  {{2, 50}, {3, 30}, {9, 20}},                // Hail → overcast, rain, stay
	10: {{2, 40}, {11, 30}, {10, 30}},              // Sleet → overcast, flurries, stay
	11: {{2, 25}, {11, 35}, {12, 40}},              // Snow Flurries → overcast, stay, moderate
	12: {{11, 30}, {12, 35}, {13, 35}},             // Moderate Snow → flurries, stay, heavy
	13: {{12, 35}, {13, 35}, {14, 30}},             // Heavy Snow → moderate, stay, blizzard
	14: {{13, 50}, {14, 30}, {12, 20}},             // Blizzard → heavy, stay, moderate
}

// GetWeatherDesc returns a weather description for a given region.
func (e *GameEngine) GetWeatherDesc(region int) string {
	if e.RegionWeather == nil {
		return ""
	}
	state, ok := e.RegionWeather[region]
	if !ok {
		state = 0
	}
	if name, ok := WeatherNames[state]; ok {
		return name
	}
	return "Clear"
}

// GetRoomWeather returns a weather line for an outdoor room, or "" for indoor.
func (e *GameEngine) GetRoomWeather(roomNum int) string {
	room := e.rooms[roomNum]
	if room == nil {
		return ""
	}
	if !isOutdoorTerrain(room.Terrain) {
		return ""
	}
	region := room.Region
	desc := e.GetWeatherDesc(region)
	if desc == "" || desc == "Sunny" || desc == "Clear" {
		return ""
	}
	return "The weather is " + desc + "."
}

// isOutdoorTerrain returns true if the terrain type is outdoors.
func isOutdoorTerrain(terrain string) bool {
	switch terrain {
	case "FOREST", "MOUNTAIN", "PLAIN", "SWAMP", "JUNGLE",
		"WASTE", "OUTDOOR_OTHER", "OUTDOOR_FLOOR", "AERIAL":
		return true
	}
	return false
}

// advanceWeather randomly transitions weather for all regions and broadcasts changes.
func (e *GameEngine) advanceWeather() {
	if e.RegionWeather == nil {
		return
	}

	// Collect all regions that have outdoor rooms
	regionSet := make(map[int]bool)
	for _, room := range e.rooms {
		if room.Region > 0 {
			regionSet[room.Region] = true
		}
	}
	// Always include region 0 (default)
	regionSet[0] = true

	for region := range regionSet {
		oldState := e.RegionWeather[region]
		transitions := weatherTransitions[oldState]
		if len(transitions) == 0 {
			continue
		}

		// Weighted random selection
		totalWeight := 0
		for _, t := range transitions {
			totalWeight += t[1]
		}
		roll := rand.Intn(totalWeight)
		newState := oldState
		for _, t := range transitions {
			roll -= t[1]
			if roll < 0 {
				newState = t[0]
				break
			}
		}

		if newState != oldState {
			e.RegionWeather[region] = newState

			// Find the transition message
			msg := weatherTransitionMessages[[2]int{oldState, newState}]
			if msg == "" {
				// Generic fallback
				newName := WeatherNames[newState]
				if newName == "Sunny" {
					msg = "The skies clear."
				} else {
					msg = "The weather changes to " + newName + "."
				}
			}

			// Broadcast to all outdoor rooms in this region
			if e.localRoomBroadcast != nil {
				for num, room := range e.rooms {
					if room.Region == region && isOutdoorTerrain(room.Terrain) {
						e.localRoomBroadcast(num, []string{msg})
					}
				}
			}
		}
	}
}

// StartWeatherCycle starts a background goroutine that changes weather periodically.
func (e *GameEngine) StartWeatherCycle() {
	go func() {
		// Weather changes every 5-10 game hours (5-10 real minutes)
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			e.advanceWeather()
		}
	}()
}
