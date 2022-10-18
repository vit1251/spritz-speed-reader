package internal

import (
	"log"
	"time"
)

type reactorEvent struct {
	at time.Time
	cb func()
	done bool
}

type Reactor struct {
	events []*reactorEvent
}

func NewReactor() *Reactor {
	return new(Reactor)
}

func createEvent(at time.Time, cb func()) *reactorEvent {
	event := reactorEvent{
		at: at,
		cb: cb,
		done: false,
	}
	return &event
}

func (self *Reactor) CallAt(at time.Time, cb func()) {
	/* Step 1. Create reactor event */
	event := createEvent(at, cb)
	/* Step 2. Register reactor event */
	self.events = append(self.events, event)
}

func (self *Reactor) CallLater(d time.Duration, cb func()) {
	/* Step 1. Ask system time */
	now := time.Now()
	/* Step 2. Create reactor event */
	self.CallAt(now.Add(d), cb)
}

func (self *Reactor) Process() {
	/* Step 1. Process events */
	for _, event := range self.events {
		wait := time.Until(event.at)
		if wait < 0 {
			event.cb()
			event.done = true
		}
	}
	/* Step 2. Cleanup queue */
	var events []*reactorEvent
	for _, event := range self.events {
		if !event.done {
			events = append(events, event)
		}
	}
//	log.Printf("Reactor queue %+v event(s)", len(events))
	self.events = events
}

// GetNextEventAt provide next time.Time with reactor event
//
func (self *Reactor) GetNextEventAt() *time.Time {
	var result *time.Time
	if len(self.events) > 0 {
		result = &self.events[0].at
		for _, event := range self.events {
			duration := event.at.Sub(*result)
			if duration > 0 {
				result = &event.at
			}
		}
	}
	log.Printf("Next timer at %q", result)
	return result
}
