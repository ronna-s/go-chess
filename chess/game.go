package chess

import (
	"errors"

	"github.com/notnil/chess"
)

type (
	Game struct {
		ptr *chess.Game
	}
)

func NewGame() *Game {
	return &Game{chess.NewGame()}
}

func (g *Game) Move(m Move) error {
	validMoves := g.ptr.ValidMoves()
	for i := range validMoves {
		move := validMoves[i]
		if move.S1() == chess.Square(m.from) &&
			move.S2() == chess.Square(m.to) {
			g.ptr.Move(move)
			return nil
		}
	}
	return errors.New("move is invalid")
}

func (g *Game) Promote(p Promotion) error {
	validMoves := g.ptr.ValidMoves()
	for i := range validMoves {
		move := validMoves[i]
		if move.S1() == chess.Square(p.from) &&
			move.S2() == chess.Square(p.to) &&
			move.Promo().String() == p.newPiece {
			g.ptr.Move(move)
			return nil
		}
	}
	return errors.New("promotion is invalid")
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

func (g *Game) getPiece(pos int) piece {
	b := g.ptr.Position().Board()
	return pieces[b.Piece(chess.Square(pos))]
}
func (g *Game) Piece(pos int) {
	g.ptr.Position().Board().Piece(chess.Square(pos))
}

func (g *Game) Debug() string {
	return g.ptr.Position().Board().Draw()
}

func (g *Game) ValidPromotions(m Move) (pieces []piece) {
	validMoves := g.ptr.ValidMoves()
	color := g.getPiece(m.from).Color
	for i := range validMoves {
		move := validMoves[i]
		if move.S1() == chess.Square(m.from) &&
			move.S2() == chess.Square(m.to) {
			pieces = append(pieces, piece{Id: move.Promo().String(), Color: color})
		}
	}
	return
}
