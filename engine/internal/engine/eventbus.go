package engine

import (
	"sync"
	"time"
)

// EngineEvent represents a real-time engine event for the admin monitor.
type EngineEvent struct {
	Time     time.Time `json:"time"`
	Category string    `json:"category"` // "monster", "time", "script", "weather", "system"
	Message  string    `json:"message"`
}

// EventBus distributes engine events to subscribed admin monitors.
type EventBus struct {
	mu          sync.RWMutex
	subscribers map[chan EngineEvent]struct{}
}

func NewEventBus() *EventBus {
	return &EventBus{
		subscribers: make(map[chan EngineEvent]struct{}),
	}
}

// Subscribe returns a channel that receives events. Call Unsubscribe to clean up.
func (eb *EventBus) Subscribe() chan EngineEvent {
	ch := make(chan EngineEvent, 100)
	eb.mu.Lock()
	eb.subscribers[ch] = struct{}{}
	eb.mu.Unlock()
	return ch
}

// Unsubscribe removes a subscriber channel.
func (eb *EventBus) Unsubscribe(ch chan EngineEvent) {
	eb.mu.Lock()
	delete(eb.subscribers, ch)
	eb.mu.Unlock()
	close(ch)
}

// HasSubscribers returns true if anyone is listening.
func (eb *EventBus) HasSubscribers() bool {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	return len(eb.subscribers) > 0
}

// Publish sends an event to all subscribers. Non-blocking; drops if buffer full.
func (eb *EventBus) Publish(category, message string) {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	if len(eb.subscribers) == 0 {
		return
	}
	event := EngineEvent{
		Time:     time.Now(),
		Category: category,
		Message:  message,
	}
	for ch := range eb.subscribers {
		select {
		case ch <- event:
		default:
			// Drop if subscriber is behind
		}
	}
}
