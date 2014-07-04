package main

import (
  "database/sql"
  "encoding/json"
  "errors"
  "fmt"
  "net/http"
  "net/url"
  "os"
  "strconv"
  _ "github.com/lib/pq"
  "github.com/rcrowley/go-tigertonic"
  "github.com/rkbodenner/parallel_universe/collection"
  "github.com/rkbodenner/parallel_universe/game"
  "github.com/rkbodenner/parallel_universe/session"
)

var players = make([]*game.Player, 0)
var playerIndex = make(map[uint64]*game.Player)

func initPlayerData(db *sql.DB) {
  rows, err := db.Query("SELECT * FROM players")
  if nil != err {
    fmt.Print(err)
  }
  for rows.Next() {
    var name string
    var id int
    if err := rows.Scan(&id, &name); err != nil {
      fmt.Print(err)
    }
    players = append(players, &game.Player{id, name})
  }

  for _,player := range players {
    playerIndex[(uint64)(player.Id)] = player
  }
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
  sessions[0] = session.NewSession(gameCollection.Games[0], players)
  sessions[0].Step(players[0])
  sessions[0].Step(players[1])

  sessions[1] = session.NewSession(gameCollection.Games[1], players)
  sessions[1].Step(players[0])
  sessions[1].Step(players[1])

  for i,session := range sessions {
    session.Id = (uint)(i+1)
    sessionIndex[(uint64)(i+1)] = session
  }
}


type CollectionHandler struct{}
func (h CollectionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  err := json.NewEncoder(w).Encode(gameCollection)
  if ( nil != err ) {
    fmt.Fprintln(w, err)
  }
}

type GameHandler struct{}
func (h GameHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

type PlayersHandler struct{}
func (h PlayersHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  err := json.NewEncoder(w).Encode(players)
  if ( nil != err ) {
    fmt.Fprintln(w, err)
  }
}

type PlayerHandler struct{}
func (h PlayerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  player_id_str := r.URL.Query().Get("player_id")
  player_id, err := strconv.ParseUint(player_id_str, 10, 64)
  if nil != err {
    http.Error(w, "Not found", http.StatusNotFound)
    return
  }

  player, ok := playerIndex[player_id]
  if ok {
    err = json.NewEncoder(w).Encode(player)
    if ( nil != err ) {
      http.Error(w, "Error", http.StatusInternalServerError)
    }
  } else {
    http.Error(w, "Not found", http.StatusNotFound)
  }
}

type SessionsHandler struct{}
func (h SessionsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  err := json.NewEncoder(w).Encode(sessions)
  if ( nil != err ) {
    http.Error(w, "Error", http.StatusInternalServerError)
  }
}


type SessionCreateHandler struct {
  db *sql.DB
}
type SessionCreateHash struct {
  StartedDate string `json:"started_date"`
  Game string `json:"game"`
  Players []string `json:"players"`
}
type SessionCreateRequest struct {
  Session SessionCreateHash `json:"session"`
}

func storeSessionPlayerAssociations(db *sql.DB, session_id int, player_ids []int) error {
  for _, player_id := range player_ids {
    _, err := db.Exec("INSERT INTO sessions_players(session_id, player_id) VALUES($1, $2)", session_id, player_id)
    if nil != err {
      return errors.New(fmt.Sprintf("Failed to create session's association with a player: %s", err))
    }
  }
  return nil
}

// Persist a new session
func (handler SessionCreateHandler) marshalFunc() (func(*url.URL, http.Header, *SessionCreateRequest) (int, http.Header, *session.Session, error)) {
  return func(u *url.URL, h http.Header, rq *SessionCreateRequest) (int, http.Header, *session.Session, error) {
    var err error

    var game_id uint64
    game_id, err = strconv.ParseUint(rq.Session.Game, 10, 64)
    if nil != err {
      return http.StatusBadRequest, nil, nil, errors.New("Expected integer game ID")
    }

    var session_id int
    err = handler.db.QueryRow("INSERT INTO sessions(id, game_id) VALUES(default, $1) RETURNING id", game_id).Scan(&session_id)
    if nil != err {
      return http.StatusInternalServerError, nil, nil, errors.New("Failed to create session")
    }

    player_ids := make([]int, 0)
    for _, player_id_str := range rq.Session.Players {
      player_id, err := strconv.ParseInt(player_id_str, 10, 32)
      if nil != err {
        return http.StatusBadRequest, nil, nil, errors.New("Expected integer player ID")
      }
      player_ids = append(player_ids, (int)(player_id))
    }

    err = storeSessionPlayerAssociations(handler.db, session_id, player_ids)
    if nil != err {
      return http.StatusInternalServerError, nil, nil, err
    }

    players := make([]*game.Player, 0)  // No need to return a completely accurate session object yet
    session := session.NewSession(gameIndex[game_id], players)
    session.Id = (uint)(session_id)
    return http.StatusCreated, nil, session, nil
  }
}


type SessionHandler struct{}
func (h SessionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  id_str := r.URL.Query().Get("session_id")
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

type StepHandler struct{}
func (h StepHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  session_id_str := r.URL.Query().Get("session_id")
  session_id, err := strconv.ParseUint(session_id_str, 10, 64)
  if nil != err {
    http.Error(w, "Session not found", http.StatusNotFound)
    return
  }
  session,ok := sessionIndex[session_id]
  if !ok {
    http.Error(w, "Session not found", http.StatusNotFound)
    return
  }

  player_id_str := r.URL.Query().Get("player_id")
  player_id, err := strconv.ParseUint(player_id_str, 10, 64)
  if nil != err {
    http.Error(w, "Player not found", http.StatusNotFound)
    return
  }
  player, ok := playerIndex[player_id]
  if !ok {
    http.Error(w, "Player not found", http.StatusNotFound)
    return
  }

  step_desc,err := url.QueryUnescape(r.URL.Query().Get("step_desc"))
  for _,step := range session.SetupSteps {
    if ( step.GetRule().Description == step_desc && step.CanBeOwnedBy(player) ) {
      step.Finish()  // FIXME. Should look in request data to see what to change.
      session.Step(player)
      session.Print()  // FIXME
      return
    }
  }
  http.Error(w, "Step not found", http.StatusNotFound)
}

func main() {
  db, err := sql.Open("postgres", "user=ralph dbname=meeple_mover sslmode=disable")
  if err != nil {
    fmt.Print(err)
  }

  initPlayerData(db)
  initGameData()
  initSessionData()

  var origin string
  origin = os.Getenv("MEEPLE_MOVER_ORIGIN_URL")
  if "" == origin {
    origin = "http://localhost:8000"
  }
  cors := tigertonic.NewCORSBuilder().AddAllowedOrigins(origin).AddAllowedHeaders("Content-Type")

  mux := tigertonic.NewTrieServeMux()
  mux.Handle("GET", "/games", cors.Build(CollectionHandler{}))
  mux.Handle("GET", "/games/{id}", cors.Build(GameHandler{}))
  mux.Handle("GET", "/players", cors.Build(PlayersHandler{}))
  mux.Handle("GET", "/players/{player_id}", cors.Build(PlayerHandler{}))
  mux.Handle("GET", "/sessions", cors.Build(SessionsHandler{}))
  mux.Handle("POST", "/sessions", cors.Build(tigertonic.Marshaled(SessionCreateHandler{db}.marshalFunc())))
  mux.Handle("GET", "/sessions/{session_id}", cors.Build(SessionHandler{}))
  mux.Handle("PUT", "/sessions/{session_id}/players/{player_id}/steps/{step_desc}", cors.Build(StepHandler{}))

  var port string
  port = os.Getenv("PORT")
  if "" == port {
    port = "8080"
  }

  http.ListenAndServe(fmt.Sprintf(":%s", port), mux)

  db.Close()
}
