package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"strconv"

	"github.com/wwgberlin/go-event-sourcing-exercise/chess"
	"github.com/wwgberlin/go-event-sourcing-exercise/handlers"
	"github.com/wwgberlin/go-event-sourcing-exercise/namegen"
	"github.com/wwgberlin/go-event-sourcing-exercise/store"
	"golang.org/x/net/websocket"
)

type app struct {
	store *store.EventStore
}
type Board struct {
	Squares [][]chess.Square
	Moves   []string
}
type page struct {
	Name  string
	Board Board
}

type msg struct {
	AggregateId string
	Type        string
	Data        string
}

func (m msg) String() string {
	return fmt.Sprintf(
		"Data=%s AggregateID=%s Type=%s",
		m.Data, m.AggregateId, m.Type,
	)
}

func newApi(d *store.EventStore) *app {
	a := app{store: d}
	a.store.Register(store.NewEventHandler(handlers.MoveHandler))
	a.store.Register(store.NewEventHandler(handlers.PromotionHandler))
	a.store.Register(store.NewEventHandler(handlers.StatusChangeHandler))

	return &a
}

func (a *app) getOrGenerateGameName(gameID string) string {
	if gameID == "" {
		gameID = namegen.Generate()
		log.Println(fmt.Errorf("New game created: %s", gameID))
	}
	return gameID
}

func (a *app) newGameHandler(w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadFile("./public/static/new_game.html")
	if err != nil {
		log.Println(err)
	}

	w.Write(data)
}

func (a *app) createGameHandler(w http.ResponseWriter, r *http.Request) {
	gameID := namegen.Generate()
	log.Println(fmt.Errorf("New game created: %s", gameID))

	w.Header().Add("Location", "/game?game_id="+gameID)
}

func (a *app) gameHandler(w http.ResponseWriter, r *http.Request) {
	gameID := a.getOrGenerateGameName(r.URL.Query().Get("game_id"))
	game := handlers.BuildGame(a.store, gameID, -1)

	var tpl bytes.Buffer
	t := template.Must(template.ParseFiles("templates/board.tmpl", "templates/game.tmpl"))
	if err := t.ExecuteTemplate(&tpl, "base", page{
		Name: gameID, Board: Board{Squares: game.Draw(), Moves: game.Moves()}}); err != nil {
		panic(err)
	}
	w.Write(tpl.Bytes())
}

func (a *app) boardHandler(w http.ResponseWriter, r *http.Request) {
	gameID := a.getOrGenerateGameName(r.URL.Query().Get("game_id"))
	lastEventIDStr := r.URL.Query().Get("last_event_id")
	if lastEventID, err := strconv.ParseInt(lastEventIDStr, 10, 32); err != nil {
		log.Println(err, lastEventIDStr)
		w.WriteHeader(http.StatusBadRequest)
		return
	} else {
		game := handlers.BuildGame(a.store, gameID, int(lastEventID))

		var tpl bytes.Buffer
		t := template.Must(template.ParseFiles("templates/board.tmpl"))
		if err := t.ExecuteTemplate(&tpl, "board", Board{game.Draw(), game.Moves()}); err != nil {
			panic(err)
		}
		w.Write(tpl.Bytes())
	}
}

func (a *app) sliderHandler(w http.ResponseWriter, r *http.Request) {
	gameID := a.getOrGenerateGameName(r.URL.Query().Get("game_id"))
	lastEventIDStr := r.URL.Query().Get("last_event_id")
	if lastEventID, err := strconv.ParseInt(lastEventIDStr, 10, 32); err != nil {
		log.Println(err, lastEventIDStr)
		w.WriteHeader(http.StatusBadRequest)
		return
	} else {
		game := handlers.BuildGame(a.store, gameID, int(lastEventID))

		var tpl bytes.Buffer
		t := template.Must(template.ParseFiles("templates/slider.tmpl"))
		if err := t.ExecuteTemplate(&tpl, "slider", len(game.Moves())); err != nil {
			panic(err)
		}
		w.Write(tpl.Bytes())
	}
}

// debugHandler writes string representation of current board state to http response
// it doesn't have any information about current game, only a list of moves, from which it builds the state
func (a *app) debugHandler(w http.ResponseWriter, r *http.Request) {
	gameID := a.getOrGenerateGameName(r.URL.Query().Get("game_id"))

	game := handlers.BuildGame(a.store, gameID, -1)

	if _, err := w.Write([]byte(game.Debug())); err != nil {
		log.Printf("can't write the response: %v", err)
	}
}

func (a *app) promotionsHandler(w http.ResponseWriter, r *http.Request) {
	gameID := a.getOrGenerateGameName(r.URL.Query().Get("game_id"))
	game := handlers.BuildGame(a.store, gameID, -1)

	query := r.URL.Query().Get("target")
	move := chess.ParseMove(query)
	promotions := game.ValidPromotions(move)

	strs := make([]string, len(promotions))
	for i := range promotions {
		strs[i] = promotions[i].ImagePath()
	}

	w.Write([]byte(strings.Join(strs, ",")))
	return
}

func (a *app) eventIDsHandler(w http.ResponseWriter, r *http.Request) {
	gameID := r.URL.Query().Get("game_id")
	var ids []int
	for _, e := range a.store.GetEvents() {
		if e.AggregateID == gameID {
			ids = append(ids, e.Id)
		}
	}
	res, err := json.Marshal(ids)
	if err != nil {
		log.Println(err)
	}
	w.Write(res)
}

func (a *app) wsEventHandler(ws *websocket.Conn, gameId string) store.EventHandler {
	return store.NewEventHandler(
		func(eventStore *store.EventStore, e store.Event) {
			if e.AggregateID == gameId {
				switch e.EventType {
				case handlers.EventMoveSuccess,
					handlers.EventPromotionSuccess:
					ws.Write([]byte("1"))
				case handlers.EventMoveFail,
					handlers.EventPromotionFail:
					ws.Write([]byte("0"))
				}
			}
		})
}

func (a *app) wsHandler(ws *websocket.Conn) {
	log.Println("websocket connection initiated")

	var m msg
	if err := websocket.JSON.Receive(ws, &m); err != nil {
		log.Println("failed reading json from websocket... closing connection")
		return
	}

	log.Println(m)
	handler := a.wsEventHandler(ws, m.AggregateId)
	a.store.Register(handler)
	for {
		if err := websocket.JSON.Receive(ws, &m); err != nil {
			a.store.Deregister(handler)
			log.Println("websocket closed or json invalid... closing connection")
			return
		}
		e := store.Event{AggregateID: m.AggregateId, EventData: m.Data}
		switch m.Type {
		case "move":
			e.EventType = handlers.EventMoveRequest
		case "promote":
			e.EventType = handlers.EventPromotionRequest
		default:
			continue
		}
		a.store.Persist(e)
	}
}

func (a *app) scoreHandler(w http.ResponseWriter, r *http.Request) {
	data := handlers.BuildScores(a.store)
	output, err := json.Marshal(data)
	if err != nil {
		log.Println(err)
	}
	w.Write(output)
}
