package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/wwgberlin/go-event-sourcing-exercise/namegen"
)

func newGameHandler(w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadFile("./public/static/new_game.html")
	if err != nil {
		log.Println(err)
	}

	w.Write(data)
}

func createGameHandler(w http.ResponseWriter, r *http.Request) {
	gameID := namegen.Generate()
	log.Println(fmt.Errorf("New game created: %s", gameID))

	w.Header().Add("Location", "/game?game_id="+gameID)
}

func gameHandler(w http.ResponseWriter, r *http.Request) {
	gameID := r.URL.Query().Get("game_id")

	if gameID == "" {
		gameID = namegen.Generate()
		log.Println(fmt.Errorf("New game created: %s", gameID))
	}

	game := buildGame(events, gameID)

	var tpl bytes.Buffer
	t := template.Must(template.ParseFiles("templates/board.tmpl", "templates/game.tmpl"))
	if err := t.ExecuteTemplate(&tpl, "base", Page{gameID, game.Draw()}); err != nil {
		panic(err)
	}
	w.Write(tpl.Bytes())
}

// move handler performs data validation and writes it to the event store if everything is correct
func moveHandler(w http.ResponseWriter, r *http.Request) {
	gameID := r.URL.Query().Get("game_id")
	game := buildGame(events, gameID)

	target := r.URL.Query().Get("target")
	move := parseMove(target)

	log.Printf("Attempt to move: %s", target)
	if err := game.Move(move); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		if _, err := w.Write([]byte("Invalid move")); err != nil {
			log.Printf("can't write the response: %v", err)
		}
		log.Println(err)
		return
	}

	events = append(events, Event{
		id:          nextID(events),
		aggregateID: gameID,
		eventData:   target,
		eventType:   EventMove,
	})

	w.WriteHeader(http.StatusOK)
	var tpl bytes.Buffer
	t := template.Must(template.ParseFiles("templates/board.tmpl"))
	if err := t.ExecuteTemplate(&tpl, "board", game.Draw()); err != nil {
		panic(err)
	}
	w.Write(tpl.Bytes())
	log.Println("Success")
}

// promoteHandler performs data validation and writes it to the event store if everything is correct
func promoteHandler(w http.ResponseWriter, r *http.Request) {
	gameID := r.URL.Query().Get("game_id")
	game := buildGame(events, gameID)

	query := r.URL.Query().Get("target")
	promotion := parsePromotion(query)
	if err := game.Promote(promotion); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		if _, err := w.Write([]byte("Invalid move")); err != nil {
			log.Printf("can't write the response: %v", err)
		}
		log.Println(err)
		return
	}

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

// debugHandler writes string representation of current board state to http response
// it doesn't have any information about current game, only a list of moves, from which it builds the state
func debugHandler(w http.ResponseWriter, r *http.Request) {
	gameID := r.URL.Query().Get("game_id")
	game := buildGame(events, gameID)
	if _, err := w.Write([]byte(game.Debug())); err != nil {
		log.Printf("can't write the response: %v", err)
	}
}

func validPromotionsHandler(w http.ResponseWriter, r *http.Request) {
	gameID := r.URL.Query().Get("game_id")
	game := buildGame(events, gameID)

	query := r.URL.Query().Get("target")
	move := parseMove(query)
	promotions := game.ValidPromotions(move)
	strs := make([]string, len(promotions))
	for i := range promotions {
		strs[i] = promotions[i].ImagePath()
	}
	w.Write([]byte(strings.Join(strs, ",")))
	return
}
