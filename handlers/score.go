package handlers

import (
	"fmt"

	"github.com/wwgberlin/go-event-sourcing-exercise/store"
)

type score struct {
	GameName string
	Type     string
}

func BuildScores(eventStore *store.EventStore) []score {
	var scores []score
	names := make(map[string]int)
	for _, event := range eventStore.GetEvents() {
		names[event.AggregateID] = event.EventType
	}
	for name, eventType := range names {
		fmt.Println(name, eventType)
		s := score{
			GameName: name,
		}
		switch eventType {
		case EventWhiteWins:
			s.Type = "PinkWins"
		case EventBlackWins:
			s.Type = "BlueWins"
		case EventDraw:
			s.Type = "Draw"
		}
		scores = append(scores, s)
	}
	return scores
}
