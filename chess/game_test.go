package chess

import (
	"testing"

	"errors"

	"github.com/notnil/chess"
)

type fakeGame struct {
	validMovesFn func() []*chess.Move
	moveFn       func(*chess.Move) error
	positionFn   func() *chess.Position
	outcomeFn    func() chess.Outcome
}

func (g *fakeGame) ValidMoves() []*chess.Move {
	return g.validMovesFn()
}
func (g *fakeGame) Move(m *chess.Move) error {
	return g.moveFn(m)
}
func (g *fakeGame) Position() *chess.Position {
	return g.positionFn()
}
func (g *fakeGame) Outcome() chess.Outcome {
	return g.outcomeFn()
}
func (g *fakeGame) Moves() []*chess.Move {
	return nil
}
func TestGame_Move(t *testing.T) {
	f := &fakeGame{}
	g := Game{f}
	f.positionFn = func() *chess.Position {
		return chess.NewGame().Position()
	}
	f.validMovesFn = func() []*chess.Move {
		return chess.NewGame().ValidMoves()
	}
	if g.Move("1-1") == nil {
		t.Error("Move should have failed")
	}
	f.moveFn = func(move *chess.Move) error {
		if move.S1() != 12 || move.S2() != 20 {
			t.Errorf("trying to move with incorrect args %v %v", move.S1(), move.S2())
		}
		return errors.New("some error")
	}
	if err := g.Move("12-20"); err == nil {
		t.Error("move should have failed")
	} else if err.Error() != "some error" {
		t.Error("wrong error returned by move")
	}
	f.moveFn = func(move *chess.Move) error {
		return nil
	}
	if g.Move("6-21") != nil {
		t.Error("Move should have succeeded")
	}
}

func TestGame_Promote(t *testing.T) {
	f := &fakeGame{}
	g := Game{f}
	var position *chess.Position
	f.positionFn = func() *chess.Position {
		return position
	}
	f.validMovesFn = func() []*chess.Move {
		g := chess.NewGame()
		g.MoveStr("e4")
		g.MoveStr("d5")
		g.MoveStr("exd5")
		g.MoveStr("e5")
		g.MoveStr("dxe6e.p.")
		g.MoveStr("f5")
		g.MoveStr("e7")
		g.MoveStr("g6")
		position = g.Position()
		return g.ValidMoves()
	}
	f.moveFn = func(move *chess.Move) error {
		if move.S1() != 52 || move.S2() != 61 {
			t.Error("Trying to promote with incorrect args")
		}
		return nil
	}

	if err := g.Promote("52-61-r"); err != nil {
		t.Error("Promotion should have succeeded")
	}
	f.moveFn = func(move *chess.Move) error {
		return nil
	}
	if err := g.Promote("52-61-r"); err != nil {
		t.Error("Promotion should have succeeded")
	}

}
