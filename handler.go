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
	"github.com/wwgberlin/go-event-sourcing-exercise/namegen"
	"golang.org/x/net/websocket"
)

type api struct {
	cmd *cmd
}

type page struct {
	Name  string
	Board [][]chess.Square
}

type msg struct {
	Target      string
	AggregateId string
	Type        string
}

func newApi(cmd *cmd) *api {
	a := api{cmd: cmd}
	cmd.register(newSubscriber(a.moveSub))
	cmd.register(newSubscriber(a.promoteSub))
	cmd.register(newSubscriber(a.statusChange))

	return &a
}

func (a *api) moveSub(event Event, s *subscriber) {
	if event.eventType != EventMoveRequest {
		return
	}
	game := a.buildGame(event.aggregateID)

	ev := Event{
		aggregateID: event.aggregateID,
	}
	if err := game.Move(parseMove(event.eventData)); err != nil {
		ev.eventType = EventMoveFail
		ev.eventData = err.Error()
	} else {
		ev.eventType = EventMoveSuccess
		ev.aggregateID = event.aggregateID
		ev.eventData = event.eventData
	}
	a.cmd.eventsCh <- ev
}

func (a *api) promoteSub(event Event, s *subscriber) {
	if event.eventType != EventPromotionRequest {
		return
	}
	game := a.buildGame(event.aggregateID)

	ev := Event{
		aggregateID: event.aggregateID,
	}
	if err := game.Promote(parsePromotion(event.eventData)); err != nil {
		ev.eventType = EventPromotionFail
		ev.eventData = err.Error()
	} else {
		ev.eventType = EventPromotionSuccess
		ev.aggregateID = event.aggregateID
		ev.eventData = event.eventData
	}
	a.cmd.eventsCh <- ev
}

func (a *api) buildGame(gameID string) *chess.Game {
	game := chess.NewGame()

	for _, event := range a.cmd.events {
		if event.aggregateID != gameID {
			continue
		}

		switch event.eventType {
		case EventMoveSuccess:
			game.Move(parseMove(event.eventData))
		case EventPromotionSuccess:
			game.Promote(parsePromotion(event.eventData))
		}
	}
	return game
}

func (a *api) getGame(gameID string) *chess.Game {
	if gameID == "" {
		gameID = namegen.Generate()
		log.Println(fmt.Errorf("New game created: %s", gameID))
	}
	return a.buildGame(gameID)
}
func (a *api) newGameHandler(w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadFile("./public/static/new_game.html")
	if err != nil {
		log.Println(err)
	}

	w.Write(data)
}

func (a *api) createGameHandler(w http.ResponseWriter, r *http.Request) {
	gameID := namegen.Generate()
	log.Println(fmt.Errorf("New game created: %s", gameID))

	w.Header().Add("Location", "/game?game_id="+gameID)
}

func (a *api) gameHandler(w http.ResponseWriter, r *http.Request) {
	gameID := r.URL.Query().Get("game_id")
	game := a.getGame(gameID)

	var tpl bytes.Buffer
	t := template.Must(template.ParseFiles("templates/board.tmpl", "templates/game.tmpl"))
	if err := t.ExecuteTemplate(&tpl, "base", page{gameID, game.Draw()}); err != nil {
		panic(err)
	}
	w.Write(tpl.Bytes())
}

func (a *api) boardHandler(w http.ResponseWriter, r *http.Request) {
	gameID := r.URL.Query().Get("game_id")
	game := a.getGame(gameID)

	var tpl bytes.Buffer
	t := template.Must(template.ParseFiles("templates/board.tmpl"))
	if err := t.ExecuteTemplate(&tpl, "board", game.Draw()); err != nil {
		panic(err)
	}
	w.Write(tpl.Bytes())
}

// debugHandler writes string representation of current board state to http response
// it doesn't have any information about current game, only a list of moves, from which it builds the state
func (a *api) debugHandler(w http.ResponseWriter, r *http.Request) {
	gameID := r.URL.Query().Get("game_id")
	game := a.getGame(gameID)

	if _, err := w.Write([]byte(game.Debug())); err != nil {
		log.Printf("can't write the response: %v", err)
	}
}

func (a *api) promotionsHandler(w http.ResponseWriter, r *http.Request) {
	gameID := r.URL.Query().Get("game_id")
	game := a.getGame(gameID)

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

// TODO: err on invalid events and games; extend error handling with err codes and msgs
func (a *api) replayHandler(w http.ResponseWriter, r *http.Request) {
	gameID := r.URL.Query().Get("game_id")
	lastEventString := r.URL.Query().Get("target")
	lastEventID, err := strconv.Atoi(lastEventString)

	if err != nil {
		log.Println(err)
	}

	var eventSubset []Event
	for _, e := range a.cmd.events {
		if e.id == lastEventID {
			break
		}
		if e.aggregateID == gameID {
			eventSubset = append(eventSubset, e)
		}
	}

	game := a.buildGame(gameID)

	w.WriteHeader(http.StatusOK)
	var tpl bytes.Buffer
	t := template.Must(template.ParseFiles("templates/board.tmpl", "templates/game.tmpl"))
	if err := t.ExecuteTemplate(&tpl, "base", page{gameID, game.Draw()}); err != nil {
		panic(err)
	}
	w.Write(tpl.Bytes())
}

func (a *api) eventIDsHandler(w http.ResponseWriter, r *http.Request) {
	gameID := r.URL.Query().Get("game_id")
	var ids []int
	for _, e := range a.cmd.events {
		if e.aggregateID == gameID {
			ids = append(ids, e.id)
		}
	}
	res, err := json.Marshal(ids)
	if err != nil {
		log.Println(err)
	}
	w.Write(res)
}

func (a *api) handleMessage(ws *websocket.Conn, m msg) Event {
	var s *subscriber
	var e Event
	switch m.Type {
	case "move":
		e = Event{
			aggregateID: m.AggregateId,
			eventType:   EventMoveRequest,
			eventData:   m.Target,
		}
		s = newSubscriber(func(e Event, s *subscriber) {
			if e.aggregateID == m.AggregateId {
				if e.eventType == EventMoveSuccess {
					ws.Write([]byte("1"))
					a.cmd.unregister(s)
				} else if e.eventType == EventMoveFail {
					ws.Write([]byte("0"))
					a.cmd.unregister(s)
				}
			}
		})
	case "promote":
		e = Event{
			aggregateID: m.AggregateId,
			eventType:   EventPromotionRequest,
			eventData:   m.Target,
		}
		s = newSubscriber(func(e Event, s *subscriber) {
			if e.aggregateID == m.AggregateId {
				if e.eventType == EventPromotionSuccess {
					ws.Write([]byte("1"))
					a.cmd.unregister(s)
				} else if e.eventType == EventPromotionFail {
					ws.Write([]byte("0"))
					a.cmd.unregister(s)
				}
			}
		})
	}
	a.cmd.register(s)
	return e
}
func (a *api) wsHandler(ws *websocket.Conn) {
	for {
		var m msg
		if err := websocket.JSON.Receive(ws, &m); err != nil {
			return
		}
		a.cmd.eventsCh <- a.handleMessage(ws, m)
	}
}

func (a *api) scoreHandler(w http.ResponseWriter, r *http.Request) {
	data := a.buildScores()
	output, err := json.Marshal(data)
	if err != nil {
		log.Println(err)
	}
	w.Write(output)
}
