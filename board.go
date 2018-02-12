package main

import (
	. "github.com/notnil/chess"
)

type piece struct {
	ImgPath string
	Type string
}

type square struct {
	piece
	Color bool
	squareId Square
	Id int
}

var pieces = map[Piece]piece{
	BlackQueen:  {ImgPath: "/images/queen-3d-pink.png", Type: "black_queen"},
	BlackKing:   {ImgPath: "/images/king-3d-pink.png", Type: "black_king"},
	BlackBishop: {ImgPath: "/images/bishop-3d-pink-big-hat-2.png", Type: "black_bishop"},
	BlackKnight: {ImgPath: "/images/knight-3d-pink.png", Type: "black_knight"},
	BlackRook:   {ImgPath: "/images/rook-3d-pink.png", Type: "black_rook"},
	BlackPawn:   {ImgPath: "/images/pawn-3d-pink.png", Type: "black_pawn"},
	WhiteQueen:  {ImgPath: "/images/queen-3d.png", Type: "white_queen"},
	WhiteKing:   {ImgPath: "/images/king-3d.png", Type: "white_king"},
	WhiteBishop: {ImgPath: "/images/bishop-3d-big-hat-2.png", Type: "white_bishop"},
	WhiteKnight: {ImgPath: "/images/knight-3d.png", Type: "white_knight"},
	WhiteRook:   {ImgPath: "/images/rook-3d.png", Type: "white_rook"},
	WhitePawn:   {ImgPath: "/images/pawn-3d.png", Type: "white_pawn"},
	NoPiece:     {ImgPath: "/images/transparent.png", Type: "no_piece"},
}

func draw(b *Board) [][]square {
	board := make([][]square, 8)
	isWhite := false
	for r := 7; r >= 0; r-- {
		board[7-r] = make([]square, 8)
		for f := 0; f < 8; f++ {
			s := r*8 + f
			p := pieces[b.Piece(Square(s))]
			board[7-r][f] = square{
				piece: p,
				Color: isWhite,
				Id: r * 8 + f,
			}
			isWhite = !isWhite
		}
		isWhite = !isWhite
	}
	return board
}
