package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

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

func newApi(d *store.EventStore) *app {
	a := app{store: d}

	cbs := []func(game handlers.Game, event store.Event, eventStore handlers.EventPersister){
		handlers.MoveHandler,
		handlers.PromotionHandler,
		handlers.GameChangedHandler,
		handlers.RollbackHandler,
	}

	for i := range cbs {
		func(i int) {
			a.store.Register(store.NewEventHandler(
				func(store *store.EventStore, event store.Event) {
					events := handlers.FilterGameMoveEvents(store.Events(), event.AggregateID)
					game := handlers.MustRebuildGame(chess.NewGame(), events, event.AggregateID, -1)
					cbs[i](game, event, store)
				},
			),
			)
		}(i)
	}

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
	events := handlers.FilterGameMoveEvents(a.store.Events(), gameID)
	game := handlers.MustRebuildGame(chess.NewGame(), events, gameID, -1)

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
	lastMoveStr := r.URL.Query().Get("last_move")
	if lastMove, err := strconv.ParseInt(lastMoveStr, 10, 32); err != nil {
		log.Println(err, lastMoveStr)
		w.WriteHeader(http.StatusBadRequest)
		return
	} else {
		events := handlers.FilterGameMoveEvents(a.store.Events(), gameID)
		game := handlers.MustRebuildGame(chess.NewGame(), events, gameID, int(lastMove))

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
	lastMoveStr := r.URL.Query().Get("last_move")
	if lastMove, err := strconv.ParseInt(lastMoveStr, 10, 32); err != nil {
		log.Println(err, lastMoveStr)
		w.WriteHeader(http.StatusBadRequest)
		return
	} else {
		events := handlers.FilterGameMoveEvents(a.store.Events(), gameID)
		game := handlers.MustRebuildGame(chess.NewGame(), events, gameID, int(lastMove))

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

	events := handlers.FilterGameMoveEvents(a.store.Events(), gameID)
	game := handlers.MustRebuildGame(chess.NewGame(), events, gameID, -1)

	if _, err := w.Write([]byte(game.Debug())); err != nil {
		log.Printf("can't write the response: %v", err)
	}
}

func (a *app) promotionsHandler(w http.ResponseWriter, r *http.Request) {
	gameID := a.getOrGenerateGameName(r.URL.Query().Get("game_id"))

	events := handlers.FilterGameMoveEvents(a.store.Events(), gameID)
	game := handlers.MustRebuildGame(chess.NewGame(), events, gameID, -1)

	query := r.URL.Query().Get("target")
	promotions := game.ValidPromotions(query)

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
	for _, e := range a.store.Events() {
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
					handlers.EventPromotionSuccess,
					handlers.EventRollbackSuccess:
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

	var m struct {
		AggregateId string
		Type        string
		Data        string
	}

	if err := websocket.JSON.Receive(ws, &m); err != nil {
		log.Println("failed reading json from websocket... closing connection")
		return
	}

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
		case "rollback":
			e.EventType = handlers.EventRollbackRequest
		default:
			continue
		}
		a.store.Persist(e)
	}
}

func (a *app) scoreHandler(w http.ResponseWriter, r *http.Request) {
	data := handlers.BuildScores(a.store)
	var tpl bytes.Buffer
	t := template.Must(template.ParseFiles("templates/scoreboard.tmpl"))
	if err := t.ExecuteTemplate(&tpl, "score_board", data); err != nil {
		panic(err)
	}
	w.Write(tpl.Bytes())
}
