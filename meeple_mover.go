package main

import (
  "encoding/json"
  "fmt"
  "net/http"
  "github.com/rkbodenner/parallel_universe/collection"
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

func collectionHandler(w http.ResponseWriter, r *http.Request) {
  err := json.NewEncoder(w).Encode(collection.NewCollection())
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

func main() {
  http.Handle("/collection", corsHandler(http.HandlerFunc(collectionHandler)))
  http.Handle("/players", corsHandler(http.HandlerFunc(playersHandler)))
  http.ListenAndServe(":8080", nil)
}
