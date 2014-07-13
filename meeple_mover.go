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
  "github.com/rkbodenner/meeple_mover/record"
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

// Add IDs to setup rules for a game, if they exist in the DB
func initSetupRuleIds(db *sql.DB, g *game.Game) error {
  rows, err := db.Query("SELECT id, description FROM setup_rules WHERE game_id = $1", g.Id)
  if nil != err {
     return err
  }
  for rows.Next() {
    var id int
    var description string
    if err := rows.Scan(&id, &description); nil != err {
      return err
    }
    for _, rule := range g.SetupRules {
      if rule.Description == description {
        rule.Id = id
        break
      }
    }
  }
  return nil
}

func initGameData(db *sql.DB) error {
  for i,game := range gameCollection.Games {
    game.Id = (uint)(i+1)
    gameIndex[(uint64)(i+1)] = game

    if err := initSetupRuleIds(db, game); err != nil {
      return err
    }
  }
  return nil
}

var sessions []*session.Session
var sessionIndex = make(map[uint64]*session.Session)

func initSessionData(db *sql.DB) {
  // Fixture data
  sessions = make([]*session.Session, 2)
  sessions[0] = session.NewSession(gameCollection.Games[0], players)
  sessions[0].Step(players[0])
  sessions[0].Step(players[1])

  sessions[1] = session.NewSession(gameCollection.Games[1], players)
  sessions[1].Step(players[0])
  sessions[1].Step(players[1])

  for i,session := range sessions[:1] {
    session.Id = (uint)(i+1)
    sessionIndex[(uint64)(i+1)] = session
  }

  /*
  // Select record from DB
  session := session.NewSession(nil, make([]*game.Player, 0))
  sessionRec := record.NewSessionRecord(session)
  err := sessionRec.Find(db, 68)
  if nil != err {
    fmt.Printf("Error finding session 68: %s\n", err)
  }
  sessions = append(sessions, session)
  /*
  // All records from DB
  records := &record.SessionRecordList{}
  err := records.FindAll(db)
  if nil != err {
    fmt.Println(err)
  }
  sessions = records.List()
  */
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

func fetchPlayersById(db *sql.DB, playerIds []int) ([]*game.Player, error) {
  players := make([]*game.Player, len(playerIds))

  for i, playerId := range playerIds {
    var name string
    err := db.QueryRow("SELECT name FROM players WHERE id = $1", playerId).Scan(&name)
    if err != nil {
       return players, err
    }
    players[i] = &game.Player{playerId, name}
  }

  return players, nil
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

    player_ids := make([]int, 0)
    for _, player_id_str := range rq.Session.Players {
      player_id, err := strconv.ParseInt(player_id_str, 10, 32)
      if nil != err {
        return http.StatusBadRequest, nil, nil, errors.New("Expected integer player ID")
      }
      player_ids = append(player_ids, (int)(player_id))
    }

    var players []*game.Player
    players, err = fetchPlayersById(handler.db, player_ids)
    if nil != err {
      return http.StatusInternalServerError, nil, nil, err
    }
    fmt.Printf("Found %d matching players\n", len(players))

    session := session.NewSession(gameIndex[game_id], players)
    session.StepAllPlayers()

    err = record.NewSessionRecord(session).Create(handler.db)
    if nil != err {
      return http.StatusInternalServerError, nil, nil, err
    }

    sessionIndex[(uint64)(session.Id)] = session

    session.Print()

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

type StepHandler struct{
  db *sql.DB
}
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
    if ( step.Rule.Description == step_desc && step.CanBeOwnedBy(player) ) {
      step.Finish()  // FIXME. Should look in request data to see what to change.

      rec := &record.SetupStepRecord{Step: step, SessionId: (int)(session.Id)}
      err := rec.Update(h.db)
      if nil != err {
        // FIXME: Revert to the previous state if we can't save.
        http.Error(w, fmt.Sprintf("Error saving update to step: %s", err), http.StatusInternalServerError)
        return
      }
      session.Step(player)
      session.Print()
      return
    }
  }
  http.Error(w, "Step not found", http.StatusNotFound)
}

func main() {
  connectString := "user=ralph dbname=meeple_mover sslmode=disable"
  herokuConnectString := os.Getenv("HEROKU_POSTGRESQL_SILVER_URL")
  if herokuConnectString != "" {
    connectString = herokuConnectString
  }

  db, err := sql.Open("postgres", connectString)
  if err != nil {
    fmt.Print(err)
  } else {
    fmt.Println("Connected to database")
  }

  initPlayerData(db)
  initGameData(db)
  initSessionData(db)

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
  mux.Handle("PUT", "/sessions/{session_id}/players/{player_id}/steps/{step_desc}", cors.Build(StepHandler{db}))

  var port string
  port = os.Getenv("PORT")
  if "" == port {
    port = "8080"
  }

  http.ListenAndServe(fmt.Sprintf(":%s", port), mux)

  db.Close()
}
