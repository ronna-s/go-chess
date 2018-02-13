package main

import (
	"bytes"
	"html/template"
	"log"
	"net/http"
	"strings"

	"fmt"

	"github.com/wwgberlin/go-event-sourcing-exercise/chess"
)

const (
	EventMove int = iota
	EventPromotion
	EventRollback
)

type Event struct {
	id        int
	eventData string
	eventType int
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
		game := buildGame(events)
		var tpl bytes.Buffer
		t := template.Must(template.ParseFiles("templates/board.tmpl", "templates/game.tmpl"))
		if err := t.ExecuteTemplate(&tpl, "base", game.Draw()); err != nil {
			panic(err)
		}
		writer.Write(tpl.Bytes())
	})

	game := chess.NewGame()

	http.Handle("/images/", http.StripPrefix("/", http.FileServer(http.Dir("./public"))))

	http.HandleFunc("/debug", debugHandler(game))
	http.HandleFunc("/move", moveHandler(game))
	http.HandleFunc("/promote", promoteHandler(game))
	http.HandleFunc("/valid_promotions", validPromotions(game))

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func buildGame(events []Event) *chess.Game {
	game := chess.NewGame()

	for _, event := range events {
		switch event.eventType {
		case EventMove:
			game.Move(parseMove(event.eventData))
		case EventPromotion:
			game.Promote(parsePromotion(event.eventData))
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

// move handler performs data validation and writes it to the event store if everything is correct
func moveHandler(game *chess.Game) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.RawQuery
		move := parseMove(query)
		if err := game.Move(move); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			if _, err := w.Write([]byte("Invalid move")); err != nil {
				log.Printf("can't write the response: %v", err)
			}
			log.Println(err)
		} else {
			events = append(events, Event{
				id:        nextID(events),
				eventType: EventMove,
				eventData: query,
			})

			w.WriteHeader(http.StatusOK)
			var tpl bytes.Buffer
			t := template.Must(template.ParseFiles("templates/board.tmpl"))
			if err := t.ExecuteTemplate(&tpl, "board", game.Draw()); err != nil {
				panic(err)
			}
			w.Write(tpl.Bytes())
		}
		return
	}
}

// move handler performs data validation and writes it to the event store if everything is correct
func promoteHandler(game *chess.Game) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.RawQuery
		promotion := parsePromotion(query)
		if err := game.Promote(promotion); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			if _, err := w.Write([]byte("Invalid move")); err != nil {
				log.Printf("can't write the response: %v", err)
			}
			log.Println(err)
		} else {
			events = append(events, Event{
				id:        nextID(events),
				eventData: query,
				eventType: EventPromotion,
			})

			w.WriteHeader(http.StatusOK)
			var tpl bytes.Buffer
			t := template.Must(template.ParseFiles("templates/board.tmpl"))
			if err := t.ExecuteTemplate(&tpl, "board", game.Draw()); err != nil {
				panic(err)
			}
			w.Write(tpl.Bytes())
		}
		return
	}
}

// boardHandler writes string representation of current board state to http response
// it doesn't have any information about current game, only a list of moves, from which it builds the state
func debugHandler(game *chess.Game) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		game := buildGame(events)
		if _, err := w.Write([]byte(game.Debug())); err != nil {
			log.Printf("can't write the response: %v", err)
		}
	}
}

func validPromotions(game *chess.Game) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.RawQuery
		move := parseMove(query)
		promotions := game.ValidPromotions(move)
		strs := make([]string, len(promotions))
		for i := range promotions {
			strs[i] = promotions[i].ImagePath()
		}
		w.Write([]byte(strings.Join(strs, ",")))
		return

	}
}
