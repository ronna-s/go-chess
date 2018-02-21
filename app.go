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
	"github.com/wwgberlin/go-event-sourcing-exercise/db"
	"github.com/wwgberlin/go-event-sourcing-exercise/handlers"
	"github.com/wwgberlin/go-event-sourcing-exercise/namegen"
	"golang.org/x/net/websocket"
)

type app struct {
	db *db.EventStore
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

func newApi(d *db.EventStore) *app {
	a := app{db: d}
	a.db.Register(db.NewEventHandler(handlers.MoveHandler))
	a.db.Register(db.NewEventHandler(handlers.PromotionHandler))
	a.db.Register(db.NewEventHandler(handlers.StatusChangeHandler))

	return &a
}

func (a *app) getGame(gameID string) *chess.Game {
	if gameID == "" {
		gameID = namegen.Generate()
		log.Println(fmt.Errorf("New game created: %s", gameID))
	}
	return handlers.BuildGame(a.db, gameID)
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
	gameID := r.URL.Query().Get("game_id")
	game := a.getGame(gameID)

	var tpl bytes.Buffer
	t := template.Must(template.ParseFiles("templates/board.tmpl", "templates/game.tmpl"))
	if err := t.ExecuteTemplate(&tpl, "base", page{
		Name: gameID, Board: Board{Squares: game.Draw(), Moves: game.Moves()}}); err != nil {
		panic(err)
	}
	w.Write(tpl.Bytes())
}

func (a *app) boardHandler(w http.ResponseWriter, r *http.Request) {
	gameID := r.URL.Query().Get("game_id")
	game := a.getGame(gameID)

	var tpl bytes.Buffer
	t := template.Must(template.ParseFiles("templates/board.tmpl"))
	if err := t.ExecuteTemplate(&tpl, "board", Board{game.Draw(), game.Moves()}); err != nil {
		panic(err)
	}
	w.Write(tpl.Bytes())
}

// debugHandler writes string representation of current board state to http response
// it doesn't have any information about current game, only a list of moves, from which it builds the state
func (a *app) debugHandler(w http.ResponseWriter, r *http.Request) {
	gameID := r.URL.Query().Get("game_id")
	game := a.getGame(gameID)

	if _, err := w.Write([]byte(game.Debug())); err != nil {
		log.Printf("can't write the response: %v", err)
	}
}

func (a *app) promotionsHandler(w http.ResponseWriter, r *http.Request) {
	gameID := r.URL.Query().Get("game_id")
	game := a.getGame(gameID)

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

// TODO: err on invalid events and games; extend error handling with err codes and msgs
func (a *app) replayHandler(w http.ResponseWriter, r *http.Request) {
	gameID := r.URL.Query().Get("game_id")
	lastEventString := r.URL.Query().Get("target")
	lastEventID, err := strconv.Atoi(lastEventString)

	if err != nil {
		log.Println(err)
	}

	var eventSubset []db.Event
	for _, e := range a.db.GetEvents() {
		if e.Id == lastEventID {
			break
		}
		if e.AggregateID == gameID {
			eventSubset = append(eventSubset, e)
		}
	}

	game := handlers.BuildGame(a.db, gameID)

	w.WriteHeader(http.StatusOK)
	var tpl bytes.Buffer
	t := template.Must(template.ParseFiles("templates/board.tmpl", "templates/game.tmpl"))
	if err := t.ExecuteTemplate(&tpl, "base", page{Name: gameID, Board: Board{Squares: game.Draw(), Moves: game.Moves()}}); err != nil {
		panic(err)
	}
	w.Write(tpl.Bytes())
}

func (a *app) eventIDsHandler(w http.ResponseWriter, r *http.Request) {
	gameID := r.URL.Query().Get("game_id")
	var ids []int
	for _, e := range a.db.GetEvents() {
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

func (a *app) wsEventHandler(ws *websocket.Conn, gameId string) db.EventHandler {
	return db.NewEventHandler(
		func(eventStore *db.EventStore, e db.Event) {
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
	sub := a.wsEventHandler(ws, m.AggregateId)
	a.db.Register(sub)
	for {
		if err := websocket.JSON.Receive(ws, &m); err != nil {
			a.db.Deregister(sub)
			log.Println("websocket closed or json invalid... closing connection")
			return
		}
		e := db.Event{AggregateID: m.AggregateId, EventData: m.Data}
		switch m.Type {
		case "move":
			e.EventType = handlers.EventMoveRequest
		case "promote":
			e.EventType = handlers.EventPromotionRequest
		default:
			continue
		}
		a.db.Persist(e)
	}
}

func (a *app) scoreHandler(w http.ResponseWriter, r *http.Request) {
	data := handlers.BuildScores(a.db)
	output, err := json.Marshal(data)
	if err != nil {
		log.Println(err)
	}
	w.Write(output)
}
