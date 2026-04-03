package engine

import (
	"fmt"
	"time"
)

var gameEpoch = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

var MonthNames = []string{
	"", "Abra", "Brama", "Manta", "Dretmar", "Alabea", "Phobrus",
	"Melma", "Banamea", "Olum", "Mixus", "Farnum", "Folster", "Feast",
}

func gameTimeSinceEpoch() time.Duration { return time.Since(gameEpoch) }
func GameMinutes() int                  { return int(gameTimeSinceEpoch().Minutes()) }
func GameHour() int                     { return GameMinutes() % 24 }
func GameDay() int                      { return (GameMinutes()/24)%343 + 1 }
func GameMonth() int {
	m := ((GameDay() - 1) / 28) + 1
	if m > 12 { m = 13 }
	return m
}
func GameYear() int { return GameMinutes()/(24*343) + 1028 }
func IsNight() bool { h := GameHour(); return h < 5 || h > 19 }
func IsDay() bool   { return !IsNight() }

func GameMonthName() string {
	m := GameMonth()
	if m >= 1 && m < len(MonthNames) {
		return MonthNames[m]
	}
	return "Feast"
}

// StartTimeCycle starts a background goroutine that publishes time change events.
func (e *GameEngine) StartTimeCycle() {
	go func() {
		lastHour := GameHour()
		wasNight := IsNight()
		ticker := time.NewTicker(10 * time.Second) // check every 10 real seconds
		defer ticker.Stop()
		for range ticker.C {
			if !e.Events.HasSubscribers() {
				lastHour = GameHour()
				wasNight = IsNight()
				continue
			}
			hour := GameHour()
			night := IsNight()
			if hour != lastHour {
				period := "day"
				if night { period = "night" }
				e.Events.Publish("time", fmt.Sprintf("Hour %d:00 — %s of %s, Day %d of %s, Year %d",
					hour, period, GameMonthName(), GameDay(), GameMonthName(), GameYear()))
				if night != wasNight {
					if night {
						e.Events.Publish("time", "Night falls across the Shattered Realms.")
					} else {
						e.Events.Publish("time", "Dawn breaks across the Shattered Realms.")
					}
				}
				lastHour = hour
				wasNight = night
			}
		}
	}()
}
