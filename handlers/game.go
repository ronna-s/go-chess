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

// FilterEvents is a function that receives an events slice and returns a new
// slice after filtering out:
// 1. events that do not belong to the gameID (AggregateID field)
// 2. events that are not of action types (move, promotion)
// 3. events that have been rolled back
func FilterEvents(events []store.Event, gameID string) []store.Event {
	return events
}

// Aggregate should receive a game, an events slice, gameID and movesCount
// and returns the game after applying the events to it:
// iterate over the events and perform actions (Move, Promote) when appropriate
// stop when you have reached the moves count
// You can assume moveCount will be -1 to perform all actions
func Aggregate(game Game, events []store.Event, gameID string, movesCount int) Game {
	return game
}
