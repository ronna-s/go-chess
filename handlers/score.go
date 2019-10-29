package handlers

import (
	"github.com/ronna-s/go-chess/store"
)

type score struct {
	GameName string
	Type     string
}

func BuildScores(eventStore *store.EventStore) []score {
	scores := map[string]score{}

	for _, event := range eventStore.Events() {
		switch event.EventType {
		case EventWhiteWins:
			scores[event.AggregateID] = score{
				GameName: event.AggregateID,
				Type:     "Blue wins",
			}
		case EventBlackWins:
			scores[event.AggregateID] = score{
				GameName: event.AggregateID,
				Type:     "Pink wins",
			}
		case EventDraw:
			scores[event.AggregateID] = score{
				GameName: event.AggregateID,
				Type:     "Draw",
			}
		}
	}
	scoresArr := make([]score, len(scores))
	i := 0
	for _, v := range scores {
		scoresArr[i] = v
		i++
	}
	return scoresArr
}

func GameChangedHandler(game Game, event store.Event, eventStore EventPersister) {
	if event.EventType != EventMoveSuccess &&
		event.EventType != EventPromotionSuccess &&
		event.EventType != EventRollbackSuccess {
		return
	}

	status := game.Status()
	if status == 0 {
		return
	}

	ev := store.Event{
		AggregateID: event.AggregateID,
		EventData:   event.EventData,
	}
	if status == 1 {
		ev.EventType = EventWhiteWins
	} else if status == 2 {
		ev.EventType = EventBlackWins
	} else if status == 3 {
		ev.EventType = EventDraw
	}
	eventStore.Persist(ev)
}
