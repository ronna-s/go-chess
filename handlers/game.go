package handlers

import (
	"github.com/wwgberlin/go-event-sourcing-exercise/chess"
	"github.com/wwgberlin/go-event-sourcing-exercise/store"
)

const (
	EventNone int = iota
	EventMoveRequest
	EventMoveSuccess
	EventMoveFail
	EventPromotionRequest
	EventPromotionSuccess
	EventPromotionFail
	EventWhiteWins
	EventBlackWins
	EventDraw
	EventRollback
)

func filterGameMoveEvents(events []store.Event, gameID string) []store.Event {
	var res []store.Event
	for _, event := range events {
		if event.AggregateID != gameID {
			continue
		}

		if event.EventType == EventRollback && len(res) > 0 {
			res = res[:len(res)-1]
			continue
		}

		if event.EventType == EventMoveSuccess ||
			event.EventType == EventPromotionSuccess {
			res = append(res, event)
		}
	}
	return res

}

func BuildGame(events []store.Event, gameID string, lastMove int) *chess.Game {
	game := chess.NewGame()

	events = filterGameMoveEvents(events, gameID)

	for i, event := range events {
		if i == lastMove {
			break
		}

		switch event.EventType {
		case EventMoveSuccess:
			game.Move(chess.ParseMove(event.EventData))
		case EventPromotionSuccess:
			game.Promote(chess.ParsePromotion(event.EventData))
		}
	}
	return game
}

func MoveHandler(eventStore *store.EventStore, event store.Event) {
	if event.EventType != EventMoveRequest {
		return
	}
	game := BuildGame(eventStore.Events(), event.AggregateID, -1)

	ev := store.Event{
		AggregateID: event.AggregateID,
	}
	if err := game.Move(chess.ParseMove(event.EventData)); err != nil {
		ev.EventType = EventMoveFail
		ev.EventData = err.Error()
	} else {
		ev.EventType = EventMoveSuccess
		ev.EventData = event.EventData
	}
	eventStore.Persist(ev)
}

func PromotionHandler(eventStore *store.EventStore, event store.Event) {
	if event.EventType != EventPromotionRequest {
		return
	}
	game := BuildGame(eventStore.Events(), event.AggregateID, -1)

	ev := store.Event{
		AggregateID: event.AggregateID,
	}
	if err := game.Promote(chess.ParsePromotion(event.EventData)); err != nil {
		ev.EventType = EventPromotionFail
		ev.EventData = err.Error()
	} else {
		ev.EventType = EventPromotionSuccess
		ev.EventData = event.EventData
	}
	eventStore.Persist(ev)
}

func StatusChangeHandler(eventStore *store.EventStore, event store.Event) {
	game := BuildGame(eventStore.Events(), event.AggregateID, -1)
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
