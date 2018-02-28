package chess

import (
	"fmt"

	"github.com/notnil/chess"
)

type Piece struct {
	ID    string
	Color Color
}

var pieces = map[chess.Piece]Piece{
	chess.BlackQueen:  {ID: "q", Color: Black},
	chess.BlackKing:   {ID: "k", Color: Black},
	chess.BlackBishop: {ID: "b", Color: Black},
	chess.BlackKnight: {ID: "n", Color: Black},
	chess.BlackRook:   {ID: "r", Color: Black},
	chess.BlackPawn:   {ID: "p", Color: Black},
	chess.WhiteQueen:  {ID: "q", Color: White},
	chess.WhiteKing:   {ID: "k", Color: White},
	chess.WhiteBishop: {ID: "b", Color: White},
	chess.WhiteKnight: {ID: "n", Color: White},
	chess.WhiteRook:   {ID: "r", Color: White},
	chess.WhitePawn:   {ID: "p", Color: White},
}

func (p Piece) ImagePath() string {
	if p.ID == "" {
		return "/images/transparent.png"
	}
	return fmt.Sprintf("/images/%s-%s.png", p.ID, p.Color)
}
