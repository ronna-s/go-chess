package main

import (
	"log"
	"net/http"

	"fmt"

	"github.com/wwgberlin/go-event-sourcing-exercise/chess"
)

const (
	EventMove int = iota
	EventPromotion
	EventRollback
)

type Event struct {
	id          int
	aggregateID string
	eventData   string
	eventType   int
}

type Page struct {
	Name  string
	Board [][]chess.Square
}

var events []Event

func nextID(events []Event) int {
	if len(events) == 0 {
		return 0
	}
	return events[len(events)-1].id + 1
}

func main() {
	http.Handle("/images/", http.StripPrefix("/", http.FileServer(http.Dir("./public"))))

	http.HandleFunc("/debug", debugHandler)
	http.HandleFunc("/move", moveHandler)
	http.HandleFunc("/promote", promoteHandler)
	http.HandleFunc("/game", gameHandler)
	http.HandleFunc("/game/", newGameHandler)
	http.HandleFunc("/create", createGameHandler)
	http.HandleFunc("/replay", replayHandler)
	http.HandleFunc("/events", eventIDsHandler)
	http.HandleFunc("/valid_promotions", validPromotionsHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func buildGame(events []Event, gameID string) *chess.Game {
	game := chess.NewGame()

	for _, event := range events {
		if event.aggregateID != gameID {
			continue
		}

		switch event.eventType {
		case EventMove:
			game.Move(parseMove(event.eventData))
		case EventPromotion:
			if err := game.Promote(parsePromotion(event.eventData)); err != nil {
				panic(err)
			}
		case EventRollback:
			//todo
		}
	}
	return game
}

func parseMove(query string) chess.Move {
	var from, to int
	if _, err := fmt.Sscanf(query, "%d-%d", &from, &to); err != nil {
		log.Println(err)
	}
	return chess.NewMove(from, to)
}

func parsePromotion(query string) chess.Promotion {
	var (
		from, to int
		newPiece string
	)

	if _, err := fmt.Sscanf(query, "%d-%d-%s", &from, &to, &newPiece); err != nil {
		log.Println(err)
	}

	return chess.NewPromotion(from, to, newPiece)
}
