package main

const (
	EventNone int = iota
	EventMoveRequest
	EventMoveSuccess
	EventMoveFail
	EventPromotionRequest
	EventPromotionSuccess
	EventPromotionFail
)

type Event struct {
	id          int
	aggregateID string
	eventData   string
	eventType   int
}
