package chess

import (
	"fmt"
	"log"
)

func parseMove(query string) Move {
	var from, to int
	if _, err := fmt.Sscanf(query, "%d-%d", &from, &to); err != nil {
		log.Println(err)
	}
	return newMove(from, to)
}

func parsePromotion(query string) Promotion {
	var (
		from, to int
		newPiece string
	)

	if _, err := fmt.Sscanf(query, "%d-%d-%s", &from, &to, &newPiece); err != nil {
		log.Println(err)
	}

	return newPromotion(from, to, newPiece)
}
