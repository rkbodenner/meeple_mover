package main

import (
  "encoding/json"
  "fmt"
  "net/http"
  "github.com/rkbodenner/parallel_universe/game"
)

type Player struct {
  Id int
  Name string
}

var players = []Player{
  {1, "Player One"},
  {2, "Player Two"},
}

func corsHandler(h http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    header := w.Header()
    header.Add("Access-Control-Allow-Origin", "http://localhost:8000")
    h.ServeHTTP(w, r)
  })
}

func playersHandler(w http.ResponseWriter, r *http.Request) {
  err := json.NewEncoder(w).Encode(players)
  if ( nil != err ) {
    fmt.Fprintln(w, err)
  }
}

func handler(w http.ResponseWriter, r *http.Request) {
  game := game.NewGame(nil, 2)
  err := json.NewEncoder(w).Encode(game)
  if ( nil != err ) {
    fmt.Fprintln(w, err)
  }
}

func main() {
  http.Handle("/", corsHandler(http.HandlerFunc(handler)))
  http.Handle("/players", corsHandler(http.HandlerFunc(playersHandler)))
  http.ListenAndServe(":8080", nil)
}
