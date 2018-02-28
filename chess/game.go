package chess

import (
	"errors"

	"github.com/notnil/chess"
)

type (
	game interface {
		ValidMoves() []*chess.Move
		Move(*chess.Move) error
		Position() *chess.Position
		Outcome() chess.Outcome
		Moves() []*chess.Move
	}
	Game struct {
		ptr game
	}
)

func NewGame() *Game {
	return &Game{chess.NewGame()}
}

func (g *Game) Move(query string) error {
	m := parseMove(query)
	validMoves := g.ptr.ValidMoves()
	for i := range validMoves {
		move := validMoves[i]
		if move.S1() == chess.Square(m.from) &&
			move.S2() == chess.Square(m.to) {
			return g.ptr.Move(move)
		}
	}
	return errors.New("move is invalid")
}

func (g *Game) Promote(query string) error {
	p := parsePromotion(query)
	validMoves := g.ptr.ValidMoves()
	for i := range validMoves {
		move := validMoves[i]
		if move.S1() == chess.Square(p.from) &&
			move.S2() == chess.Square(p.to) &&
			move.Promo().String() == p.newPiece {
			return g.ptr.Move(move)
		}
	}
	return errors.New("promotion is invalid")
}

func (g *Game) Moves() []string {
	newGame := chess.NewGame()
	strs := make([]string, len(g.ptr.Moves()))
	moves := g.ptr.Moves()
	for i := range moves {
		strs[i] = chess.AlgebraicNotation{}.Encode(newGame.Position(), moves[i])
		newGame.Move(moves[i])
	}
	return strs
}

func (g *Game) Draw() [][]Square {
	board := make([][]Square, 8)
	isWhite := false
	for r := 7; r >= 0; r-- {
		board[7-r] = make([]Square, 8)
		for f := 0; f < 8; f++ {
			pos := r*8 + f
			p := g.getPiece(pos)
			board[7-r][f] = Square{
				Piece: p,
				Color: Color(isWhite),
				Pos:   pos,
			}
			isWhite = !isWhite
		}
		isWhite = !isWhite
	}
	return board
}

func (g *Game) getPiece(pos int) Piece {
	b := g.ptr.Position().Board()
	return pieces[b.Piece(chess.Square(pos))]
}

func (g *Game) Debug() string {
	return g.ptr.Position().Board().Draw()
}

func (g *Game) ValidPromotions(query string) (pieces []Piece) {
	m := parseMove(query)
	validMoves := g.ptr.ValidMoves()
	color := g.getPiece(m.from).Color
	for i := range validMoves {
		move := validMoves[i]
		if move.S1() == chess.Square(m.from) &&
			move.S2() == chess.Square(m.to) {
			pieces = append(pieces, Piece{ID: move.Promo().String(), Color: color})
		}
	}
	return
}

// 0 ongoing
// 1 - won by white
// 2 - won by black
// 3 - draw
func (g *Game) Status() int {
	switch g.ptr.Outcome() {
	case "1-0":
		return 1
	case "0-1":
		return 2
	case "1/2-1/2":
		return 3
	}
	return 0
}
