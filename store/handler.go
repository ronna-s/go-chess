package store

type EventHandler interface {
	cb(*EventStore, Event)
}

type eventHandler struct {
	cbFn func(*EventStore, Event)
}

func (h *eventHandler) cb(store *EventStore, e Event) {
	h.cbFn(store, e)
}

func NewEventHandler(cb func(*EventStore, Event)) EventHandler {
	return &eventHandler{cbFn: cb}
}
