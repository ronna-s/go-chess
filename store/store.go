package store

import "log"

type EventStore struct {
	events       []Event
	eventsCh     chan Event
	registerCh   chan *EventListener
	unregisterCh chan *EventListener
	listeners    []*EventListener
}

func NewEventStore() *EventStore {
	var c EventStore

	return &c
}

func (store *EventStore) Events() []Event {
	return store.events
}

func (store *EventStore) Run() {
	store.eventsCh = make(chan Event)
	store.registerCh = make(chan *EventListener)
	store.unregisterCh = make(chan *EventListener)

	go func() {
		for {
			select {
			case e := <-store.eventsCh:
				e.Id = store.nextID(store.events)
				store.events = append(store.events, e)
				for _, s := range store.listeners {
					s.notify(store, e)
				}
			case reg := <-store.registerCh:
				store.listeners = append(store.listeners, reg)
			case unreg := <-store.unregisterCh:
				for i := range store.listeners {
					if unreg == store.listeners[i] {
						store.listeners = append(store.listeners[:i], store.listeners[i+1:]...)
						log.Println("event handler deregistered")
						break
					}
				}
			}
		}
	}()
}

func (store *EventStore) nextID(events []Event) int {
	if len(events) == 0 {
		return 0
	}
	return events[len(events)-1].Id + 1
}

func (store *EventStore) Persist(e Event) {
	go func() {
		store.eventsCh <- e
	}()
}

func (store *EventStore) Register(s *EventListener) {
	go func() {
		store.registerCh <- s
	}()
}

func (store *EventStore) Unregister(s *EventListener) {
	go func() {
		store.unregisterCh <- s
	}()
}
