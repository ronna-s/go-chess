package handlers

import (
	"github.com/wwgberlin/go-event-sourcing-exercise/store"
)

type score struct {
	GameName string
	Type     string
}

func BuildScores(eventStore *store.EventStore) []score {
	var scores []score
	names := make(map[string]int)
	for _, event := range eventStore.Events() {
		names[event.AggregateID] = event.EventType
	}
	for name, eventType := range names {
		s := score{
			GameName: name,
		}
		switch eventType {
		case EventWhiteWins:
			s.Type = "Blue won"
		case EventBlackWins:
			s.Type = "Pink won"
		case EventDraw:
			s.Type = "Draw"
		}
		scores = append(scores, s)
	}
	return scores
}
