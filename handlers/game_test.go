package handlers

import (
	"fmt"
	"reflect"
	"testing"

	"errors"

	"github.com/wwgberlin/go-event-sourcing-exercise/store"
)

type FakeGame struct {
	Game
	moveFn    func(query string) error
	promoteFn func(query string) error
}

func successFn(query string) error {
	return nil
}
func failFn(query string) error {
	return errors.New("some error")
}
func (g FakeGame) Move(query string) error {
	return g.moveFn(query)
}

func (g FakeGame) Promote(query string) error {
	return g.promoteFn(query)
}

type FakeStore struct {
	persistFn func(store.Event)
}

func (s FakeStore) Persist(event store.Event) {
	s.persistFn(event)
}

func TestMoveHandlerBasic(t *testing.T) {
	var s FakeStore
	var persisted []store.Event
	g := &FakeGame{moveFn: successFn}

	s.persistFn = func(event store.Event) {
		if event.EventType != EventMoveSuccess {
			t.Error("trying to persist event of the wrong type")
		}
		persisted = append(persisted, event)
	}

	testCases := struct {
		events            []store.Event
		expectedToPersist []store.Event
	}{
		events: []store.Event{
			{EventType: EventMoveRequest, EventData: "move1", AggregateID: "some game"},
			{EventType: EventMoveSuccess, EventData: "should ignore", AggregateID: "other game"},
			{EventType: EventPromotionRequest, EventData: "move2", AggregateID: "other game"},
			{EventType: EventMoveRequest, EventData: "move2", AggregateID: "other game"},
		},
		expectedToPersist: []store.Event{
			{EventType: EventMoveSuccess, EventData: "move1", AggregateID: "some game"},
			{EventType: EventMoveSuccess, EventData: "move2", AggregateID: "other game"},
		},
	}

	for _, event := range testCases.events {
		MoveHandler(g, event, s)
	}

	if !reflect.DeepEqual(persisted, testCases.expectedToPersist) {
		t.Error("expected to persist: ", testCases.expectedToPersist, "but persisted:", persisted)
	}
}
func TestMoveHandlerError(t *testing.T) {
	var s FakeStore
	var persisted []store.Event
	g := &FakeGame{moveFn: failFn}

	s.persistFn = func(event store.Event) {
		if event.EventType != EventMoveFail {
			t.Error("trying to persist event of the wrong type")
		}
		persisted = append(persisted, event)
	}

	testCases := struct {
		events            []store.Event
		expectedToPersist []store.Event
	}{
		events: []store.Event{
			{EventType: EventMoveRequest, EventData: "move1", AggregateID: "some game"},
			{EventType: EventMoveRequest, EventData: "move2", AggregateID: "other game"},
			{EventType: EventMoveSuccess, EventData: "should ignore", AggregateID: "other game"},
			{EventType: EventPromotionRequest, EventData: "move2", AggregateID: "other game"},
		},
		expectedToPersist: []store.Event{
			{EventType: EventMoveFail, EventData: "some error", AggregateID: "some game"},
			{EventType: EventMoveFail, EventData: "some error", AggregateID: "other game"},
		},
	}

	for _, event := range testCases.events {
		MoveHandler(g, event, s)
	}

	if !reflect.DeepEqual(persisted, testCases.expectedToPersist) {
		t.Error("expected to persist: ", testCases.expectedToPersist, "but persisted:", persisted)
	}
}

func TestFilterGameMoveEvents(t *testing.T) {
	myGameID := "my game"
	otherGameID := "other game"

	events := []store.Event{
		{EventType: EventMoveRequest, EventData: "ignore 1", AggregateID: myGameID},
		{EventType: EventMoveSuccess, EventData: "append 1", AggregateID: myGameID},
		{EventType: EventMoveSuccess, EventData: "append 2", AggregateID: myGameID},
		{EventType: EventMoveRequest, EventData: "ignore 2", AggregateID: otherGameID},
		{EventType: EventMoveSuccess, EventData: "ignore 3", AggregateID: otherGameID},
		{EventType: EventRollbackSuccess, EventData: "ignore ", AggregateID: otherGameID},
		{EventType: EventRollbackSuccess, EventData: "remove append 2", AggregateID: myGameID},
		{EventType: EventPromotionSuccess, EventData: "append 3", AggregateID: myGameID},
		{EventType: EventPromotionRequest, EventData: "ignore", AggregateID: myGameID},
	}

	filtered := FilterGameMoveEvents(events, myGameID)
	expected := []store.Event{
		{EventType: EventMoveSuccess, EventData: "append 1", AggregateID: myGameID},
		{EventType: EventPromotionSuccess, EventData: "append 3", AggregateID: myGameID},
	}

	if !reflect.DeepEqual(filtered, expected) {
		t.Error("expected to return", expected, "but received", filtered)
	}
}

func TestPromotionHandlerBasic(t *testing.T) {
	var s FakeStore
	var persisted []store.Event
	g := &FakeGame{moveFn: successFn}

	s.persistFn = func(event store.Event) {
		if event.EventType != EventMoveSuccess {
			t.Error("trying to persist event of the wrong type")
		}
		persisted = append(persisted, event)
	}

	testCases := struct {
		events            []store.Event
		expectedToPersist []store.Event
	}{
		events: []store.Event{
			{EventType: EventMoveRequest, EventData: "move1", AggregateID: "some game"},
			{EventType: EventMoveSuccess, EventData: "should ignore", AggregateID: "other game"},
			{EventType: EventPromotionRequest, EventData: "move2", AggregateID: "other game"},
			{EventType: EventMoveRequest, EventData: "move2", AggregateID: "other game"},
		},
		expectedToPersist: []store.Event{
			{EventType: EventMoveSuccess, EventData: "move1", AggregateID: "some game"},
			{EventType: EventMoveSuccess, EventData: "move2", AggregateID: "other game"},
		},
	}

	for _, event := range testCases.events {
		MoveHandler(g, event, s)
	}

	if !reflect.DeepEqual(persisted, testCases.expectedToPersist) {
		t.Error("expected to persist: ", testCases.expectedToPersist, "but persisted:", persisted)
	}
}

func TestPromotionHandlerError(t *testing.T) {
	var s FakeStore
	var persisted []store.Event
	g := &FakeGame{promoteFn: failFn}

	s.persistFn = func(event store.Event) {
		if event.EventType != EventPromotionFail {
			t.Error("trying to persist event of the wrong type")
		}
		persisted = append(persisted, event)
	}

	testCases := struct {
		events            []store.Event
		expectedToPersist []store.Event
	}{
		events: []store.Event{
			{EventType: EventPromotionRequest, EventData: "move1", AggregateID: "some game"},
			{EventType: EventPromotionRequest, EventData: "move2", AggregateID: "other game"},
			{EventType: EventPromotionSuccess, EventData: "should ignore", AggregateID: "other game"},
			{EventType: EventMoveRequest, EventData: "move2", AggregateID: "other game"},
		},
		expectedToPersist: []store.Event{
			{EventType: EventPromotionFail, EventData: "some error", AggregateID: "some game"},
			{EventType: EventPromotionFail, EventData: "some error", AggregateID: "other game"},
		},
	}

	for _, event := range testCases.events {
		PromotionHandler(g, event, s)
	}

	if !reflect.DeepEqual(persisted, testCases.expectedToPersist) {
		t.Error("expected to persist: ", testCases.expectedToPersist, "but persisted:", persisted)
	}
}

func TestRollbackHandler(t *testing.T) {
	var s FakeStore
	var persisted []store.Event

	g := &FakeGame{promoteFn: failFn}

	s.persistFn = func(event store.Event) {
		if event.EventType != EventRollbackSuccess {
			t.Error("trying to persist event of the wrong type")
		}
		persisted = append(persisted, event)
	}

	testCases := struct {
		events            []store.Event
		expectedToPersist []store.Event
	}{
		events: []store.Event{
			{EventType: EventRollbackRequest, AggregateID: "some game"},
			{EventType: EventPromotionSuccess, AggregateID: "other game"},
			{EventType: EventMoveRequest, AggregateID: "other game"},
			{EventType: EventRollbackRequest, AggregateID: "other game"},
		},
		expectedToPersist: []store.Event{
			{EventType: EventRollbackSuccess, AggregateID: "some game"},
			{EventType: EventRollbackSuccess, AggregateID: "other game"},
		},
	}

	for _, event := range testCases.events {
		RollbackHandler(g, event, s)
	}

	if !reflect.DeepEqual(persisted, testCases.expectedToPersist) {
		t.Error("expected to persist: ", testCases.expectedToPersist, "but persisted:", persisted)
	}
}

func TestRebuildGameNoEvents(t *testing.T) {
	game := &FakeGame{
		moveFn: func(query string) error {
			t.Error("shouldn't move")
			return nil
		},
		promoteFn: func(query string) error {
			t.Error("shouldn't promote")
			return nil
		},
	}

	if res := MustRebuildGame(game, []store.Event{}, "some id", -1); res.(*FakeGame) != game {
		t.Error("Return value incorrect")
	}

	if res := MustRebuildGame(game, []store.Event{}, "some id", 0); res.(*FakeGame) != game {
		t.Error("Return value incorrect")
	}

	if res := MustRebuildGame(game, []store.Event{}, "some id", 0); res.(*FakeGame) != game {
		t.Error("Return value incorrect")
	}

}

func TestRebuildGame_MoveEvents(t *testing.T) {
	var queries []string
	game := &FakeGame{
		moveFn: func(query string) error {
			queries = append(queries, query)
			return nil
		},
	}

	myGameID := "my game"

	testCases := struct {
		events        []store.Event
		expectedMoves []string
	}{
		events: []store.Event{
			{AggregateID: myGameID, EventType: EventMoveSuccess, EventData: "Hey"},
			{AggregateID: myGameID, EventType: EventMoveSuccess, EventData: "I'm moving"},
			{AggregateID: myGameID, EventType: EventMoveSuccess, EventData: "Well done!"},
			{AggregateID: myGameID, EventType: EventMoveSuccess, EventData: "All done!"},
		},
		expectedMoves: []string{"Hey", "I'm moving", "Well done!", "All done!"},
	}
	if res := MustRebuildGame(game, testCases.events, myGameID, -1); res.(*FakeGame) != game {
		t.Error("Return value incorrect")
	}

	if !reflect.DeepEqual(testCases.expectedMoves, queries) {
		t.Error("Expected:", testCases.expectedMoves, "but received:", queries)
	}

}

func TestRebuildGame_MovePromotionEvents(t *testing.T) {
	var queries []string

	game := &FakeGame{
		moveFn: func(query string) error {
			queries = append(queries, fmt.Sprintf("move: %s", query))
			return nil
		},
		promoteFn: func(query string) error {
			queries = append(queries, fmt.Sprintf("promotion: %s", query))
			return nil
		},
	}

	myGameID := "my game"

	testCases := struct {
		events        []store.Event
		expectedMoves []string
	}{
		events: []store.Event{
			{AggregateID: myGameID, EventType: EventMoveSuccess, EventData: "Hey"},
			{AggregateID: myGameID, EventType: EventMoveSuccess, EventData: "I'm moving"},
			{AggregateID: myGameID, EventType: EventPromotionSuccess, EventData: "I promote"},
			{AggregateID: myGameID, EventType: EventMoveSuccess, EventData: "Well done!"},
		},
		expectedMoves: []string{"move: Hey", "move: I'm moving", "promotion: I promote", "move: Well done!"},
	}

	if res := MustRebuildGame(game, testCases.events, myGameID, -1); res.(*FakeGame) != game {
		t.Error("Return value incorrect")
	}

	if !reflect.DeepEqual(testCases.expectedMoves, queries) {
		t.Error("Expected:", testCases.expectedMoves, "but received:", queries)
	}
}

func TestRebuildGame_MovePromotionEventsWithLastMoveID(t *testing.T) {
	var queries []string
	game := &FakeGame{
		moveFn: func(query string) error {
			queries = append(queries, fmt.Sprintf("move: %s", query))
			return nil
		},
		promoteFn: func(query string) error {
			queries = append(queries, fmt.Sprintf("promotion: %s", query))
			return nil
		},
	}

	myGameID := "my game"
	otherGameID := "other game"

	testCases := struct {
		events        []store.Event
		expectedMoves []string
	}{
		events: []store.Event{
			{AggregateID: myGameID, EventType: EventMoveSuccess, EventData: "Hey"},
			{AggregateID: myGameID, EventType: EventMoveSuccess, EventData: "I'm moving"},
			{AggregateID: myGameID, EventType: EventPromotionSuccess, EventData: "I promote"},
			{AggregateID: myGameID, EventType: EventMoveSuccess, EventData: "Well done!"},
			{AggregateID: otherGameID, EventType: EventPromotionSuccess, EventData: "Please ignore me"},
			{AggregateID: myGameID, EventType: EventMoveSuccess, EventData: "All done!"},
		},
		expectedMoves: []string{"move: Hey", "move: I'm moving", "promotion: I promote"},
	}

	if res := MustRebuildGame(game, testCases.events, myGameID, 3); res.(*FakeGame) != game {
		t.Error("Return value incorrect")
	}

	if !reflect.DeepEqual(testCases.expectedMoves, queries) {
		t.Error("Expected:", testCases.expectedMoves, "but received:", queries)
	}
}

func TestRebuildGame_MovePromotionEventsWithRollback(t *testing.T) {
	var queries []string
	game := &FakeGame{
		moveFn: func(query string) error {
			queries = append(queries, fmt.Sprintf("move: %s", query))
			return nil
		},
		promoteFn: func(query string) error {
			queries = append(queries, fmt.Sprintf("promotion: %s", query))
			return nil
		},
	}

	myGameID := "my game"
	otherGameID := "other game"

	testCases := struct {
		events        []store.Event
		expectedMoves []string
	}{
		events: []store.Event{
			{AggregateID: myGameID, EventType: EventMoveSuccess, EventData: "Hey"},
			{AggregateID: myGameID, EventType: EventMoveSuccess, EventData: "I'm moving"},
			{AggregateID: myGameID, EventType: EventPromotionSuccess, EventData: "I promote"},
			{AggregateID: myGameID, EventType: EventRollbackSuccess},
			{AggregateID: otherGameID, EventType: EventRollbackSuccess},
		},
		expectedMoves: []string{"move: Hey", "move: I'm moving"},
	}

	if res := MustRebuildGame(game, testCases.events, myGameID, -1); res.(*FakeGame) != game {
		t.Error("Return value incorrect")
	}

	if !reflect.DeepEqual(testCases.expectedMoves, queries) {
		t.Error("Expected:", testCases.expectedMoves, "but received:", queries)
	}
}
