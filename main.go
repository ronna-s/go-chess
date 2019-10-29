package main

import (
	"log"
	"net/http"

	"github.com/ronna-s/go-chess/store"
	"golang.org/x/net/websocket"
)

func main() {
	store := store.NewEventStore()
	store.Run()
	api := newApi(store)

	http.Handle("/images/", http.StripPrefix("/", http.FileServer(http.Dir("./public/static"))))
	http.HandleFunc("/debug", api.debugHandler)
	http.HandleFunc("/game", api.gameHandler)
	http.HandleFunc("/board", api.boardHandler)
	http.HandleFunc("/slider", api.sliderHandler)
	http.HandleFunc("/", api.newGameHandler)
	http.HandleFunc("/create", api.createGameHandler)
	http.HandleFunc("/events", api.eventIDsHandler)
	http.HandleFunc("/promotions", api.promotionsHandler)
	http.HandleFunc("/scores", api.scoreHandler)

	http.Handle("/ws", websocket.Handler(api.wsHandler))

	log.Fatal(http.ListenAndServe(":8080", nil))
}
