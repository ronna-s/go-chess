package main

import (
	"log"
	"net/http"

	"github.com/notnil/chess"
	"bytes"
	"html/template"
	"strconv"
	"strings"
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
			game.Move(queryToMove(game, event.move))
		}

		var tpl bytes.Buffer
		t := template.Must(template.ParseFiles("templates/board.tmpl", "templates/game.tmpl"))
		if err := t.ExecuteTemplate(&tpl, "base", draw(game.Position().Board())); err != nil {
			panic(err)
		}
		writer.Write(tpl.Bytes())
	})

	game := chess.NewGame()

	http.Handle("/images/", http.StripPrefix("/", http.FileServer(http.Dir("./public"))))

	http.HandleFunc("/board", boardHandler)
	http.HandleFunc("/move", moveHandler(game))
	http.HandleFunc("/promote", promoteHandler(game))
	http.HandleFunc("/valid_promotions", validPromotions(game))

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func queryToMove(game *chess.Game, query string) *chess.Move {
	moves := strings.Split(query, "-")
	fromSquare, _ := strconv.ParseInt(moves[0], 10, 8)
	toSquare, _ := strconv.ParseInt(moves[1], 10, 8)
	var promo string
	if len(moves) > 2 {
		promo = moves[2]
	}
	for _, move := range game.ValidMoves() {
		if move.S1() == chess.Square(fromSquare) &&
			move.S2() == chess.Square(toSquare) &&
			move.Promo().String() == promo {
			return move
		}
	}
	return nil
}
func queryToPromotion(game *chess.Game, query string) *chess.Move {
	moves := strings.Split(query, "-")
	fromSquare, _ := strconv.ParseInt(moves[0], 10, 8)
	toSquare, _ := strconv.ParseInt(moves[1], 10, 8)
	for _, move := range game.ValidMoves() {
		if move.S1() == chess.Square(fromSquare) &&
			move.S2() == chess.Square(toSquare) &&
			move.Promo().String() == moves[2] {
			return move
		}
	}
	return nil
}

// move handler performs data validation and writes it to the event store if everything is correct
func moveHandler(game *chess.Game) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.RawQuery
		move := queryToMove(game, query)
		if err := game.Move(move); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			if _, err := w.Write([]byte("Invalid move")); err != nil {
				log.Printf("can't write the response: %v", err)
			}
			log.Println(err)
		} else {
			events = append(events, Event{
				id:   nextID(events),
				move: query,
			})

			w.WriteHeader(http.StatusOK)
			var tpl bytes.Buffer
			t := template.Must(template.ParseFiles("templates/board.tmpl"))
			if err := t.ExecuteTemplate(&tpl, "board", draw(game.Position().Board())); err != nil {
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
		move := queryToPromotion(game, query)
		if err := game.Move(move); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			if _, err := w.Write([]byte("Invalid move")); err != nil {
				log.Printf("can't write the response: %v", err)
			}
			log.Println(err)
		} else {
			events = append(events, Event{
				id:   nextID(events),
				move: query,
			})

			w.WriteHeader(http.StatusOK)
			var tpl bytes.Buffer
			t := template.Must(template.ParseFiles("templates/board.tmpl"))
			if err := t.ExecuteTemplate(&tpl, "board", draw(game.Position().Board())); err != nil {
				panic(err)
			}
			w.Write(tpl.Bytes())
		}
		return
	}
}

// boardHandler writes string representation of current board state to http response
// it doesn't have any information about current game, only a list of moves, from which it builds the state
func boardHandler(w http.ResponseWriter, r *http.Request) {
	game := chess.NewGame()

	for _, event := range events {
		move := queryToMove(game, event.move)
		game.Move(move)
	}

	if _, err := w.Write([]byte(game.Position().Board().Draw())); err != nil {
		log.Printf("can't write the response: %v", err)
	}
}

func validPromotions(game *chess.Game) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.RawQuery
		moves := strings.Split(query, "-")
		fromSquare, _ := strconv.ParseInt(moves[0], 10, 8)
		toSquare, _ := strconv.ParseInt(moves[1], 10, 8)
		color := ""
		if chess.White == game.Position().Board().Piece(chess.Square(fromSquare)).Color() {
			color = "white"
		} else {
			color = "black"
		}
		var promotions []string
		for i := range game.ValidMoves() {
			move := game.ValidMoves()[i]
			if move.S1() == chess.Square(fromSquare) && move.S2() == chess.Square(toSquare) {
				promotions = append(promotions, "/images/"+move.Promo().String()+"-"+color+".png")
			}
		}
		w.Write([]byte(strings.Join(promotions, ",")))
		return

	}
}
