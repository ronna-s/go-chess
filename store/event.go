package store

import "fmt"

type Event struct {
	Id          int
	AggregateID string
	EventData   string
	EventType   int
}

func (ev Event) String() string {
	return fmt.Sprintf("Event<Id: %d, AggregateID: %s, EventData: %s, EventType: %d>",
		ev.Id, ev.AggregateID, ev.EventData, ev.EventType)
}
