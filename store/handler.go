package store

type EventListener struct {
	NotifFn func(*EventStore, Event)
}

func (h *EventListener) notify(store *EventStore, e Event) {
	h.NotifFn(store, e)
}

func NewEventHandler(cb func(*EventStore, Event)) *EventListener {
	return &EventListener{NotifFn: cb}
}
