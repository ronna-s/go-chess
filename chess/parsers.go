package chess

import (
	"fmt"
	"log"
)

func ParseMove(query string) Move {
	var from, to int
	if _, err := fmt.Sscanf(query, "%d-%d", &from, &to); err != nil {
		log.Println(err)
	}
	return NewMove(from, to)
}

func ParsePromotion(query string) Promotion {
	var (
		from, to int
		newPiece string
	)

	if _, err := fmt.Sscanf(query, "%d-%d-%s", &from, &to, &newPiece); err != nil {
		log.Println(err)
	}

	return NewPromotion(from, to, newPiece)
}
