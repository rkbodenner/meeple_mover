package main

import (
  "encoding/json"
  "fmt"
  "net/http"
  "github.com/rkbodenner/parallel_universe/game"
)

func handler(w http.ResponseWriter, r *http.Request) {
  game := game.NewGame(nil, 2)
  err := json.NewEncoder(w).Encode(game)
  if ( nil != err ) {
    fmt.Fprintln(w, err)
  }
}

func main() {
  http.HandleFunc("/", handler)
  http.ListenAndServe(":8080", nil)
}
