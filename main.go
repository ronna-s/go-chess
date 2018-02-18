package main

import (
	"log"
	"net/http"

	"github.com/wwgberlin/go-event-sourcing-exercise/db"
	"golang.org/x/net/websocket"
)

func main() {
	db := db.NewEventStore()
	db.Run()
	api := newApi(db)

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
