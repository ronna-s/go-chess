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

// MoveHandler should listen on events of type EventMoveRequest
// It will check if possible to perform the move
// and persist a new EventMoveSuccess if success
// otherwise it will persist EventMoveFail
func MoveHandler(game Game, event store.Event, eventStore EventPersister) {
}

// PromotionHandler should listen on events of type EventPromotionRequest
// It will check if possible to perform the promotion
// and persist a new EventPromotionSuccess if success,
// otherwise it will persist EventPromotionFail
func PromotionHandler(game Game, event store.Event, eventStore EventPersister) {
}

// Rollback handler should listen on events of type EventRollbackRequest
// and persist a new EventRollbackSuccess to the store (rollback cannot fail)
func RollbackHandler(_ Game, event store.Event, eventStore EventPersister) {
}

// FilterGameMoveEvents is a function that receives an events slice and returns a new
// slice after filtering out:
// 1. events that do not belong to the gameID (AggregateID field)
// 2. events that are not of action types (move, promotion)
// 3. events that have been rolled back
func FilterGameMoveEvents(events []store.Event, gameID string) []store.Event {
	return events
}

// MustRebuildGame should receive a game, an events slice, gameID and movesCount
// and returns the game after applying the events to it:
// iterate over the events and perform actions (Move, Promote) when appropriate
// stop when you have reached the moves count
// You can assume moveCount will be -1 to perform all actions
func MustRebuildGame(game Game, events []store.Event, gameID string, movesCount int) Game {
	return game
}
