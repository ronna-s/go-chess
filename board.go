package main

import (
	. "github.com/notnil/chess"
	"strconv"
)

type piece struct {
	ImgPath string
	PieceId string
}

type square struct {
	piece
	Color bool
	Row int
	Col string
}

var pieces = map[Piece]piece{
	BlackQueen:  {ImgPath: "/images/queen-3d-pink.png", PieceId: "black_queen"},
	BlackKing:   {ImgPath: "/images/king-3d-pink.png", PieceId: "back_king"},
	BlackBishop: {ImgPath: "/images/bishop-3d-pink-big-hat-2.png", PieceId: "black_bishop"},
	BlackKnight: {ImgPath: "/images/knight-3d-pink.png", PieceId: "black-knight"},
	BlackRook:   {ImgPath: "/images/rook-3d-pink.png", PieceId: "black-rook"},
	BlackPawn:   {ImgPath: "/images/pawn-3d-pink.png", PieceId: "black-pawn"},
	WhiteQueen:  {ImgPath: "/images/queen-3d.png", PieceId: "white_queen"},
	WhiteKing:   {ImgPath: "/images/king-3d.png", PieceId: "white_king"},
	WhiteBishop: {ImgPath: "/images/bishop-3d-big-hat-2.png", PieceId: "white_bishop"},
	WhiteKnight: {ImgPath: "/images/knight-3d.png", PieceId: "white_knight"},
	WhiteRook:   {ImgPath: "/images/rook-3d.png", PieceId: "white_rook"},
	WhitePawn:   {ImgPath: "/images/pawn-3d.png", PieceId: "white_pawn"},
	NoPiece:     {ImgPath: "/images/transparent.png", PieceId: "no_piece"},
}

func draw(b *Board) [][]square {
	board := make([][]square, 8)
	isWhite := false
	rows := []string{"a","b","c","d","e","f","g","h"}
	for r := 7; r >= 0; r-- {
		board[r] = make([]square, 8)
		for f := 0; f < 8; f++ {
			s := r*8 + f
			p := pieces[b.Piece(Square(s))]
			board[r][f] = square{
				Row: r+1,
				Col: rows[f],
				piece: piece{
					PieceId: p.PieceId + strconv.Itoa(s),
					ImgPath: p.ImgPath,
				},
				Color: isWhite,
			}
			isWhite = !isWhite
		}
		isWhite = !isWhite
	}
	return board
}
