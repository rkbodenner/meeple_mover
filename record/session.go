package record

import (
  "database/sql"
  "errors"
  "fmt"
  _ "github.com/lib/pq"
  "github.com/rkbodenner/parallel_universe/game"
  "github.com/rkbodenner/parallel_universe/session"
)

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
    if nil == step.Owner {
      _, err = db.Exec("INSERT INTO setup_steps(session_id, setup_rule_id, player_id, done) VALUES($1, $2, $3, $4)",
        rec.s.Id, step.Rule.Id, nil, step.Done)
    } else {
      _, err = db.Exec("INSERT INTO setup_steps(session_id, setup_rule_id, player_id, done) VALUES($1, $2, $3, $4)",
        rec.s.Id, step.Rule.Id, step.Owner.Id, step.Done)
    }
    if nil != err {
      return i, errors.New(fmt.Sprintf("Failed to create setup step: %s", err))
    }
  }
  return len(rec.s.SetupSteps), nil
}

func (rec *SessionRecord) Create(db *sql.DB) error {
  var err error

  err = db.QueryRow("INSERT INTO sessions(id, game_id) VALUES(default, $1) RETURNING id", rec.s.Game.Id).Scan(&rec.s.Id)
  if nil != err {
    return err
  }

  _, err = rec.storeSessionPlayerAssociations(db)
  if nil != err {
    return err
  }

  _, err = rec.storeSetupSteps(db)
  if nil != err {
    return err
  }

  return nil
}

func (rec *SessionRecord) Find(db *sql.DB, id int) error {
  rec.s.Id = (uint)(id)

  var err error

  var gameId int
  err = db.QueryRow("SELECT game_id FROM sessions WHERE id = $1", id).Scan(&gameId)
  if nil != err {
    return err
  }

  // Eager-load the associated game
  g := &game.Game{}
  gameRec := NewGameRecord(g)
  err = gameRec.Find(db, gameId)
  if nil != err {
    return err
  }
  rec.s.Game = g

  // Eager-load the associated players
  var players = make([]*game.Player, 0)
  var playerRows *sql.Rows
  playerRows, err = db.Query("SELECT p.id, p.name FROM players p INNER JOIN sessions_players sp ON sp.player_id = p.id WHERE sp.session_id = $1", id)
  if nil != err {
    return err
  }
  defer playerRows.Close()
  for playerRows.Next() {
    var name string
    var id int
    if err := playerRows.Scan(&id, &name); err != nil {
      return err
    }
    players = append(players, &game.Player{id, name})
  }
  rec.s.Players = players

  // Eager-load the associated setup steps and associate those in turn according to their belongs-to relationships
  setupSteps := NewSetupStepRecordList()
  err = setupSteps.FindBySession(db, rec.s)
  if nil != err {
    return err
  }
  setupSteps.AssociatePlayers(rec.s.Players)
  setupSteps.AssociateRules(rec.s.Game.SetupRules)
  rec.s.SetupSteps = setupSteps.List()

  // Eager-load setup step assignments
  var assignRows *sql.Rows
  assignRows, err = db.Query("SELECT setup_rule_id, player_id FROM setup_step_assignments WHERE session_id = $1", rec.s.Id)
  if nil != err {
    return err
  }
  defer assignRows.Close()
  assignmentCount := 0
  for assignRows.Next() {
    var setupRuleId, playerId int
    if err := assignRows.Scan(&setupRuleId, &playerId); nil != err {
      return err
    }

    // Find step that matches
    var player *game.Player = nil
    for _, p := range rec.s.Players {
      if p.Id == playerId {
        player = p
        break
      }
    }
    if nil == player {
      return errors.New(fmt.Sprintf("Error assigning step for rule %d to player %d: No such player\n", setupRuleId, playerId))
    }
    var step *game.SetupStep = nil
    for _, s := range rec.s.SetupSteps {
      if s.Rule.Id == setupRuleId && s.CanBeOwnedBy(player) {
        step = s
        break
      }
    }
    if nil == step {
      return errors.New(fmt.Sprintf("Error assigning step for rule %d to player %d: No such step ownable by player\n", setupRuleId, playerId))
    }
    rec.s.SetupAssignments.Set(player, step)
    assignmentCount++
  }

  return nil
}

type SessionRecordList struct {
  records []*SessionRecord
}

func (recs *SessionRecordList) List() []*session.Session {
  sessions := make([]*session.Session, 0)
  for _,rec := range recs.records {
    sessions = append(sessions, rec.s)
  }
  return sessions
}

func (recs *SessionRecordList) FindAll(db *sql.DB) error {
  recs.records = make([]*SessionRecord, 0)
  ids := make([]int, 0)

  rows, err := db.Query("SELECT id FROM sessions")
  if nil != err {
    return err
  }
  defer rows.Close()
  for rows.Next() {
    var id int
    if err := rows.Scan(&id); err != nil {
      return err
    }
    ids = append(ids, id)
  }

  for _, id := range ids {
    session := session.NewEmptySession()
    sessionRec := NewSessionRecord(session)
    err := sessionRec.Find(db, id)
    if nil != err {
      return errors.New(fmt.Sprintf("Error finding session %d: %s", id, err))
    }
    recs.records = append(recs.records, sessionRec)
  }

  return nil
}
