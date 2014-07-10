package record

import (
  "database/sql"
  "errors"
  "fmt"
  _ "github.com/lib/pq"
  "github.com/rkbodenner/parallel_universe/session"
)

type Record interface {
  Create(*sql.DB) error
  Read(*sql.DB) error
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
