package main

import (
  "fmt"
  "net/http"
  "strings"
  "github.com/rkbodenner/parallel_universe/game"
)

func handler(w http.ResponseWriter, r *http.Request) {
  game := game.NewGame(nil, 2)
  players := make([]string, 0)
  for _,p := range game.Players {
    players = append(players, (string)(p))
  }
  playerText := strings.Join(players, "\n")

  fmt.Fprintf(w, playerText)
}

func main() {
  http.HandleFunc("/", handler)
  http.ListenAndServe(":8080", nil)
}
