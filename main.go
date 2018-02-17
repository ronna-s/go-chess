package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/wwgberlin/go-event-sourcing-exercise/chess"
	"golang.org/x/net/websocket"
)

func main() {
	cmd := newCmd()
	cmd.run()
	api := newApi(cmd)

	http.Handle("/images/", http.StripPrefix("/", http.FileServer(http.Dir("./public/static"))))
	http.HandleFunc("/debug", api.debugHandler)
	http.HandleFunc("/game", api.gameHandler)
	http.HandleFunc("/board", api.boardHandler)
	http.HandleFunc("/", api.newGameHandler)
	http.HandleFunc("/create", api.createGameHandler)
	http.HandleFunc("/replay", api.replayHandler)
	http.HandleFunc("/events", api.eventIDsHandler)
	http.HandleFunc("/promotions", api.promotionsHandler)
	http.HandleFunc("/scores", api.scoreHandler)

	http.Handle("/ws", websocket.Handler(api.wsHandler))

	log.Fatal(http.ListenAndServe(":8080", nil))
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
