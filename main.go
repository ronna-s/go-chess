package main

import (
	"log"
	"net/http"

	"github.com/notnil/chess"
)

type Event struct {
	id   int
	move string
}

func nextID(events []Event) int {
	if len(events) == 0 {
		return 0
	}
	return events[len(events)-1].id + 1
}

// eventstore
var events []Event

func main() {
	game := chess.NewGame()

	http.HandleFunc("/move", moveHandler(game))
	http.HandleFunc("/board", boardHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// move handler performs data validation and writes it to the event store if everything is correct
func moveHandler(game *chess.Game) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		move := r.URL.RawQuery

		if err := game.MoveStr(move); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			if _, err := w.Write([]byte("Invalid move")); err != nil {
				log.Printf("can't write the response: %v", err)
			}
			log.Println(err)
			return
		}

		events = append(events, Event{
			id:   nextID(events),
			move: move,
		})

		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("Moved to " + r.URL.RawQuery)); err != nil {
			log.Printf("can't write the response: %v", err)
		}
	}
}

// boardHandler writes string representation of current board state to http response
// it doesn't have any information about current game, only a list of moves, from which it builds the state
func boardHandler(w http.ResponseWriter, r *http.Request) {
	game := chess.NewGame()

	for _, event := range events {
		game.MoveStr(event.move)
	}

	if _, err := w.Write([]byte(game.Position().Board().Draw())); err != nil {
		log.Printf("can't write the response: %v", err)
	}
}
