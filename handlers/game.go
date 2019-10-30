package handlers

import (
	"github.com/ronna-s/go-chess/chess"
	"github.com/ronna-s/go-chess/store"
)

const (
	_                     = iota
	EventMoveRequest
	EventMoveSuccess
	EventMoveFail
	EventPromotionRequest
	EventPromotionSuccess
	EventPromotionFail
	EventWhiteWins
	EventBlackWins
	EventDraw
	EventRollbackRequest
	EventRollbackSuccess
)

type Game interface {
	Move(query string) error
	Promote(query string) error
	Moves() []string
	Status() int
	Draw() [][]chess.Square
	Debug() string
	ValidPromotions(query string) (pieces []chess.Piece)
}

type EventPersister interface {
	Persist(event store.Event)
}

// MoveHandler should listen on events of type EventMoveRequest (and ignore all others)
// It will check if possible to perform the move and persist a new EventMoveSuccess if success
// otherwise it will persist EventMoveFail
// look at store/event.go to see how events are defined
func MoveHandler(game Game, event store.Event, eventStore EventPersister) {
	if event.EventType != EventMoveRequest {
		return
	}

	ev := store.Event{
		AggregateID: event.AggregateID,
	}
	if err := game.Move(event.EventData); err != nil {
		ev.EventType = EventMoveFail
		ev.EventData = err.Error()
	} else {
		ev.EventType = EventMoveSuccess
		ev.EventData = event.EventData
	}

	eventStore.Persist(ev)
}

// PromotionHandler should listen on events of type EventPromotionRequest
// It will check if possible to perform the promotion
// and persist a new EventPromotionSuccess if success,
// otherwise it will persist EventPromotionFail
func PromotionHandler(game Game, event store.Event, eventStore EventPersister) {
	if event.EventType != EventPromotionRequest {
		return
	}
	ev := store.Event{
		AggregateID: event.AggregateID,
	}
	if err := game.Promote(event.EventData); err != nil {
		ev.EventType = EventPromotionFail
		ev.EventData = err.Error()
	} else {
		ev.EventType = EventPromotionSuccess
		ev.EventData = event.EventData
	}
	eventStore.Persist(ev)
}

// Rollback handler should listen on events of type EventRollbackRequest
// and persist a new EventRollbackSuccess to the store (rollback cannot fail)
func RollbackHandler(_ Game, event store.Event, eventStore EventPersister) {
	if event.EventType != EventRollbackRequest {
		return
	}

	ev := store.Event{
		AggregateID: event.AggregateID,
		EventType:   EventRollbackSuccess,
	}
	eventStore.Persist(ev)
}

// FilterEvents is a function that receives an events slice and returns a new
// slice after filtering out:
// 1. events that do not belong to the gameID (AggregateID field)
// 2. events that are not of action types (move, promotion)
// 3. events that have been rolled back
func FilterEvents(events []store.Event, gameID string) []store.Event {
	var res []store.Event
	for _, event := range events {
		if event.AggregateID != gameID {
			continue
		}
		switch event.EventType {
		case EventMoveSuccess, EventPromotionSuccess:
			res = append(res, event)
		case EventRollbackSuccess:
			if len(res) > 0 {
				res = res[:len(res)-1]
				continue
			}
		}
	}
	return res
}

// Aggregate should receive a game, an events slice, gameID and movesCount
// and returns the game after applying the events to it:
// iterate over the events and perform actions (Move, Promote) when appropriate
// stop when you have reached the moves count
// You can assume moveCount will be -1 to perform all actions
func Aggregate(game Game, events []store.Event, gameID string, movesCount int) Game {
	for i, event := range events {
		if i == movesCount {
			break
		}

		switch event.EventType {
		case EventMoveSuccess:
			game.Move(event.EventData)
		case EventPromotionSuccess:
			game.Promote(event.EventData)
		}
	}
	return game
}
