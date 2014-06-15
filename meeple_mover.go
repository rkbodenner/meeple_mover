package main

import (
  "encoding/json"
  "fmt"
  "net/http"
  "strconv"
  "github.com/rcrowley/go-tigertonic"
  "github.com/rkbodenner/parallel_universe/collection"
  "github.com/rkbodenner/parallel_universe/game"
  "github.com/rkbodenner/parallel_universe/session"
)

var players = []*game.Player{
  &game.Player{1, "Player One"},
  &game.Player{2, "Player Two"},
}

var gameCollection = collection.NewCollection()
var gameIndex = make(map[uint64]*game.Game)

func initGameData() {
  for i,game := range gameCollection.Games {
    game.Id = (uint)(i+1)
    gameIndex[(uint64)(i+1)] = game
  }
}

var sessions []*session.Session
var sessionIndex = make(map[uint64]*session.Session)

func initSessionData() {
  sessions = make([]*session.Session, 2)
  sessions[0] = session.NewSession(gameCollection.Games[0], 2)
  sessions[1] = session.NewSession(gameCollection.Games[1], 2)

  for i,session := range sessions {
    session.Id = (uint)(i+1)
    sessionIndex[(uint64)(i+1)] = session
  }
}

func corsHandler(handler func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
  return func(w http.ResponseWriter, r *http.Request) {
    header := w.Header()
    header.Add("Access-Control-Allow-Origin", "http://localhost:8000")
    handler(w, r)
  }
}

func collectionHandler(w http.ResponseWriter, r *http.Request) {
  err := json.NewEncoder(w).Encode(gameCollection)
  if ( nil != err ) {
    fmt.Fprintln(w, err)
  }
}

func gameHandler(w http.ResponseWriter, r *http.Request) {
  id_str := r.URL.Query().Get("id")
  id, err := strconv.ParseUint(id_str, 10, 64)
  if nil != err {
    http.Error(w, "Not found", http.StatusNotFound)
    return
  }

  game, ok := gameIndex[id]
  if ok {
    err := json.NewEncoder(w).Encode(game)
    if ( nil != err ) {
      http.Error(w, "Error", http.StatusInternalServerError)
    }
  } else {
    http.Error(w, "Not found", http.StatusNotFound)
  }
}

func playersHandler(w http.ResponseWriter, r *http.Request) {
  err := json.NewEncoder(w).Encode(players)
  if ( nil != err ) {
    fmt.Fprintln(w, err)
  }
}

func sessionsHandler(w http.ResponseWriter, r *http.Request) {
  err := json.NewEncoder(w).Encode(sessions)
  if ( nil != err ) {
    http.Error(w, "Error", http.StatusInternalServerError)
  }
}

func sessionHandler(w http.ResponseWriter, r *http.Request) {
  id_str := r.URL.Query().Get("id")
  id, err := strconv.ParseUint(id_str, 10, 64)
  if nil != err {
    http.Error(w, "Not found", http.StatusNotFound)
    return
  }

  session, ok := sessionIndex[id]
  if ok {
    err := json.NewEncoder(w).Encode(session)
    if ( nil != err ) {
      http.Error(w, "Error", http.StatusInternalServerError)
    }
  } else {
    http.Error(w, "Not found", http.StatusNotFound)
  }
}

func main() {
  initGameData()
  initSessionData()

  mux := tigertonic.NewTrieServeMux()
  mux.HandleFunc("GET", "/games", corsHandler(collectionHandler))
  mux.HandleFunc("GET", "/games/{id}", corsHandler(gameHandler))
  mux.HandleFunc("GET", "/players", corsHandler(playersHandler))
  mux.HandleFunc("GET", "/sessions", corsHandler(sessionsHandler))
  mux.HandleFunc("GET", "/sessions/{id}", corsHandler(sessionHandler))
  http.ListenAndServe(":8080", mux)
}
