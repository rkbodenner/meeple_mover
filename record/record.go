package record

import (
  "database/sql"
  "errors"
  "fmt"
  _ "github.com/lib/pq"
  "github.com/rkbodenner/parallel_universe/game"
  "github.com/rkbodenner/parallel_universe/session"
)

type Record interface {
  Create(*sql.DB) error
  Find(db *sql.DB, id int) error
}

type SessionRecord struct {
  s *session.Session
}

func NewSessionRecord(s *session.Session) *SessionRecord {
  return &SessionRecord{s: s}
}

func (rec *SessionRecord) playerIds() []int {
  playerIds := make([]int, 0)
  for _, player := range rec.s.Players {
    playerIds = append(playerIds, (int)(player.Id))
  }
  return playerIds
}

func (rec *SessionRecord) storeSessionPlayerAssociations(db *sql.DB) (int, error) {
  playerIds := rec.playerIds()
  for i, playerId := range playerIds {
    _, err := db.Exec("INSERT INTO sessions_players(session_id, player_id) VALUES($1, $2)", rec.s.Id, playerId)
    if nil != err {
      return i, errors.New(fmt.Sprintf("Failed to create session's association with a player: %s", err))
    }
  }
  return len(playerIds), nil
}

func (rec *SessionRecord) storeSetupSteps(db *sql.DB) (int, error) {
  for i, step := range rec.s.SetupSteps {
    var err error
    if nil == step.GetOwner() {
      _, err = db.Exec("INSERT INTO setup_steps(session_id, setup_rule_id, player_id, done) VALUES($1, $2, $3, $4)",
        rec.s.Id, step.GetRule().Id, nil, step.IsDone())
    } else {
      _, err = db.Exec("INSERT INTO setup_steps(session_id, setup_rule_id, player_id, done) VALUES($1, $2, $3, $4)",
        rec.s.Id, step.GetRule().Id, step.GetOwner().Id, step.IsDone())
    }
    if nil != err {
      return i, errors.New(fmt.Sprintf("Failed to create setup step: %s", err))
    }
  }
  return len(rec.s.SetupSteps), nil
}

func (rec *SessionRecord) Create(db *sql.DB) error {
  var n int
  var err error

  err = db.QueryRow("INSERT INTO sessions(id, game_id) VALUES(default, $1) RETURNING id", rec.s.Game.Id).Scan(&rec.s.Id)
  if nil != err {
    return err
  }
  fmt.Printf("Created session #%d\n", rec.s.Id)

  n, err = rec.storeSessionPlayerAssociations(db)
  if nil != err {
    return err
  }
  fmt.Printf("Created %d session-player associations\n", n)

  n, err = rec.storeSetupSteps(db)
  if nil != err {
    return err
  }
  fmt.Printf("Created %d setup steps\n", n)

  return nil
}

func (rec *SessionRecord) Find(db *sql.DB, id int) error {
  var err error

  var gameId int
  err = db.QueryRow("SELECT game_id FROM sessions WHERE id = $1", id).Scan(&gameId)
  if nil != err {
    return err
  }

  rec.s.Id = (uint)(id)

  // Eager-load the associated game
  if nil == rec.s.Game {
    game := &game.Game{SetupRules: make([]*game.SetupRule, 0)}
    gameRec := NewGameRecord(game)
    err = gameRec.Find(db, gameId)
    if nil != err {
      return err
    }
    rec.s.Game = game
  }

  // Eager-load the associated players
  var players = make([]*game.Player, 0)
  rows, err := db.Query("SELECT p.id, p.name FROM players p INNER JOIN sessions_players sp ON sp.player_id = p.id WHERE sp.session_id = $1", id)
  if nil != err {
    return err
  }
  for rows.Next() {
    var name string
    var id int
    if err := rows.Scan(&id, &name); err != nil {
      return err
    }
    players = append(players, &game.Player{id, name})
  }
  rec.s.Players = players

  return nil
}

type GameRecord struct {
  g *game.Game
}

func NewGameRecord(g *game.Game) *GameRecord {
  return &GameRecord{g: g}
}

func (rec *GameRecord) Find(db *sql.DB, id int) error {
  var err error

  var name string
  err = db.QueryRow("SELECT name FROM games WHERE id = $1", id).Scan(&name)
  if nil != err {
    return err
  }

  rec.g.Id = (uint)(id)
  rec.g.Name = name

  // TODO: Read setup rules

  return nil
}
