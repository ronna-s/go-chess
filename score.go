package main

import "fmt"

type score struct {
	GameName string
	Type     string
}

func (a *api) statusChange(event Event, s *subscriber) {
	game := a.buildGame(event.aggregateID)
	status := game.Status()
	if status == 0 {
		return
	}

	ev := Event{
		aggregateID: event.aggregateID,
		eventData:   event.eventData,
	}
	if status == 1 {
		ev.eventType = EventWhiteWins
	} else if status == 2 {
		ev.eventType = EventBlackWins
	} else if status == 3 {
		ev.eventType = EventDraw
	}
	a.cmd.eventsCh <- ev
}

func (a *api) buildScores() []score {
	var scores []score
	names := make(map[string]int)
	for _, event := range a.cmd.events {
		names[event.aggregateID] = event.eventType
	}
	for name, eventType := range names {
		fmt.Println(name, eventType)
		s := score{
			GameName: name,
		}
		switch eventType {
		case EventWhiteWins:
			s.Type = "PinkWins"
		case EventBlackWins:
			s.Type = "BlueWins"
		case EventDraw:
			s.Type = "Draw"
		}
		scores = append(scores, s)
	}
	return scores
}
