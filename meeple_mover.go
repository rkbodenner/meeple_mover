package main

import (
  "encoding/json"
  "fmt"
  "net/http"
  "strconv"
  "github.com/rcrowley/go-tigertonic"
  "github.com/rkbodenner/parallel_universe/collection"
  "github.com/rkbodenner/parallel_universe/session"
)

type Player struct {
  Id int
  Name string
}

var players = []Player{
  {1, "Player One"},
  {2, "Player Two"},
}

var sessions = make(map[uint64]*session.Session)

func corsHandler(handler func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
  return func(w http.ResponseWriter, r *http.Request) {
    header := w.Header()
    header.Add("Access-Control-Allow-Origin", "http://localhost:8000")
    handler(w, r)
  }
}

func collectionHandler(w http.ResponseWriter, r *http.Request) {
  collection := collection.NewCollection()
  var i uint = 1
  for _,game := range collection.Games {
    game.Id = i
    i++
  }
  err := json.NewEncoder(w).Encode(collection)
  if ( nil != err ) {
    fmt.Fprintln(w, err)
  }
}

func playersHandler(w http.ResponseWriter, r *http.Request) {
  err := json.NewEncoder(w).Encode(players)
  if ( nil != err ) {
    fmt.Fprintln(w, err)
  }
}

func sessionHandler(w http.ResponseWriter, r *http.Request) {
  id_str := r.URL.Query().Get("id")
  id, err := strconv.ParseUint(id_str, 10, 64)
  if nil != err {
    http.Error(w, "Not found", http.StatusNotFound)
    return
  }

  session, ok := sessions[id]
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
  sessions[1] = session.NewSession(collection.NewForbiddenIsland(), 2)

  mux := tigertonic.NewTrieServeMux()
  mux.HandleFunc("GET", "/collection", corsHandler(collectionHandler))
  mux.HandleFunc("GET", "/players", corsHandler(playersHandler))
  mux.HandleFunc("GET", "/sessions/{id}", sessionHandler)
  http.ListenAndServe(":8080", mux)
}
