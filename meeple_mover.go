package main

import (
  "encoding/json"
  "fmt"
  "net/http"
  "os"
  "github.com/rkbodenner/parallel_universe/collection"
  "source.datanerd.us/ralph/go_agent"
)

type Player struct {
  Id int
  Name string
}

var players = []Player{
  {1, "Player One"},
  {2, "Player Two"},
}

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

func main() {
  go_agent.StartAgent(os.Getenv("NEW_RELIC_LICENSE_KEY"), "meeple_mover")

  http.HandleFunc("/collection", go_agent.InstrumentHttpHandler("/collection", corsHandler(collectionHandler)))
  http.HandleFunc("/players", go_agent.InstrumentHttpHandler("/players", corsHandler(playersHandler)))
  http.ListenAndServe(":8080", nil)
}
