package main

import (
	"log"
	"net/http"

	"github.com/notnil/chess"
	"bytes"
	"html/template"
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

	http.HandleFunc("/game/", func(writer http.ResponseWriter, request *http.Request) {
		game := chess.NewGame()

		for _, event := range events {
			game.MoveStr(event.move)
		}

		var tpl bytes.Buffer
		t := template.New("game.tmpl")
		t.ParseFiles("templates/game.tmpl")
		if err := t.Execute(&tpl, draw(game.Position().Board())); err!=nil{
			panic(err)
		}
		writer.Write(tpl.Bytes())
	})


	game := chess.NewGame()

	http.Handle("/images/", http.StripPrefix("/", http.FileServer(http.Dir("./public"))))

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
