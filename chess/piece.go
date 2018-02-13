package chess

import (
	"fmt"

	"github.com/notnil/chess"
)

type piece struct {
	Id    string
	Color Color
}

var pieces = map[chess.Piece]piece{
	chess.BlackQueen:  {Id: "q", Color: Black},
	chess.BlackKing:   {Id: "k", Color: Black},
	chess.BlackBishop: {Id: "b", Color: Black},
	chess.BlackKnight: {Id: "n", Color: Black},
	chess.BlackRook:   {Id: "r", Color: Black},
	chess.BlackPawn:   {Id: "p", Color: Black},
	chess.WhiteQueen:  {Id: "q", Color: White},
	chess.WhiteKing:   {Id: "k", Color: White},
	chess.WhiteBishop: {Id: "b", Color: White},
	chess.WhiteKnight: {Id: "n", Color: White},
	chess.WhiteRook:   {Id: "r", Color: White},
	chess.WhitePawn:   {Id: "p", Color: White},
}

func (p piece) ImagePath() string {
	if p.Id == "" {
		return "/images/transparent.png"
	}
	return fmt.Sprintf("/images/%s-%s.png", p.Id, p.Color)
}
