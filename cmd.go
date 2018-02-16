package main

type subscriber struct {
	cb func(Event, *subscriber)
}

type cmd struct {
	events      []Event
	eventsCh    chan Event
	subscribers []*subscriber
}

func newSubscriber(cb func(Event, *subscriber)) *subscriber {
	return &subscriber{cb: cb}
}

func newCmd() *cmd {
	var c cmd

	return &c
}

func (a *cmd) run() {
	a.eventsCh = make(chan Event)
	go func() {
		for e := range a.eventsCh {
			a.dispatchEvent(e)
		}
	}()
}

func (c *cmd) nextID(events []Event) int {
	if len(events) == 0 {
		return 0
	}
	return events[len(events)-1].id + 1
}

func (c *cmd) dispatchEvent(e Event) {
	e.id = c.nextID(c.events)
	c.events = append(c.events, e)
	for _, s := range c.subscribers {
		go s.cb(e, s)
	}
}

func (c *cmd) register(s *subscriber) {
	c.subscribers = append(c.subscribers, s)
}

func (c *cmd) unregister(s *subscriber) {
	for i := range c.subscribers {
		if s == c.subscribers[i] {
			c.subscribers = append(c.subscribers[:i], c.subscribers[i+1:]...)
			return
		}
	}
	panic("callback not found")
}
