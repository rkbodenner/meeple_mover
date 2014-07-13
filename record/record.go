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
  fmt.Printf("Loading session #%d\n", id)
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
  fmt.Printf("Loaded game #%d\n", gameId)
  rec.s.Game = g

  // Eager-load the associated players
  var players = make([]*game.Player, 0)
  rows, err := db.Query("SELECT p.id, p.name FROM players p INNER JOIN sessions_players sp ON sp.player_id = p.id WHERE sp.session_id = $1", id)
  if nil != err {
    return err
  }
  defer rows.Close()
  for rows.Next() {
    var name string
    var id int
    if err := rows.Scan(&id, &name); err != nil {
      return err
    }
    players = append(players, &game.Player{id, name})
  }
  fmt.Printf("Loaded %d players\n", len(players))
  rec.s.Players = players

  // Eager-load the associated setup steps and associate those in turn according to their belongs-to relationships
  setupSteps := NewSetupStepRecordList()
  err = setupSteps.FindBySession(db, rec.s)
  if nil != err {
    return err
  }
  setupSteps.AssociatePlayers(rec.s.Players)
  setupSteps.AssociateRules(rec.s.Game.SetupRules)
  fmt.Printf("Loaded %d setup steps\n", len(setupSteps.List()))
  rec.s.SetupSteps = setupSteps.List()

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
    session := session.NewSession(nil, make([]*game.Player, 0))
    sessionRec := NewSessionRecord(session)
    err := sessionRec.Find(db, id)
    if nil != err {
      return errors.New(fmt.Sprintf("Error finding session %d: %s", id, err))
    }
    recs.records = append(recs.records, sessionRec)
  }

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

  // Eager-load the associated game's setup rules
  rules := NewSetupRuleRecordList()
  err = rules.FindByGame(db, rec.g)
  if err != nil {
    return err
  }
  fmt.Printf("Loaded %d setup rules\n", len(rules.List()))
  rec.g.SetupRules = rules.List()

  return nil
}

type GameRecordList struct {
  records []*GameRecord
}

func (recs *GameRecordList) FindAll(db *sql.DB) error {
  recs.records = make([]*GameRecord, 0)
  ids := make([]int, 0)

  rows, err := db.Query("SELECT id FROM games")
  if nil != err {
    return err
  }
  defer rows.Close()
  for rows.Next() {
    var id int
    if err := rows.Scan(&id); nil != err {
      return err
    }
    ids = append(ids, id)
  }

  for _, id := range ids {
    gameRec := &GameRecord{&game.Game{}}
    err := gameRec.Find(db, id)
    if nil != err {
      return errors.New(fmt.Sprintf("Error finding game %d: %s", id, err))
    }
    fmt.Printf("Loaded game %d\n", gameRec.g.Id)
    recs.records = append(recs.records, gameRec)
  }

  return nil
}

func (recs *GameRecordList) List() []*game.Game {
  games := make([]*game.Game, 0)
  for _, rec := range recs.records {
    games = append(games, rec.g)
  }
  return games
}


type SetupRuleRecord struct {
  Rule *game.SetupRule
  // TODO DependencyIds []int
}

type SetupRuleRecordList struct {
  records []*SetupRuleRecord
}

func NewSetupRuleRecordList() *SetupRuleRecordList {
  return &SetupRuleRecordList{make([]*SetupRuleRecord, 0)}
}

func (recs *SetupRuleRecordList) List() []*game.SetupRule {
  steps := make([]*game.SetupRule, 0)
  for _,rec := range recs.records {
    steps = append(steps, rec.Rule)
  }
  return steps
}

func (rules *SetupRuleRecordList) FindByGame(db *sql.DB, g *game.Game) error {
  rules.records = make([]*SetupRuleRecord, 0)

  rows, err := db.Query("SELECT id, description, each_player FROM setup_rules WHERE game_id = $1", g.Id)
  if nil != err {
    return err
  }
  defer rows.Close()
  for rows.Next() {
    rule := &game.SetupRule{}
    record := &SetupRuleRecord{Rule: rule}
    var eachPlayer bool
    if err := rows.Scan(&record.Rule.Id, &record.Rule.Description, &eachPlayer); nil != err {
      return err
    }
    if eachPlayer {
      record.Rule.Arity = "Each player"
    } else {
      record.Rule.Arity = "Once"
    }
    rules.records = append(rules.records, record)
  }

  return nil
}


type SetupStepRecord struct {
  Step *game.SetupStep
  SessionId int
  // TODO: These are just a cache so we can associate objects we create elsewhere
  RuleId int
  OwnerId sql.NullInt64
}

// Only the 'done' field is updatable, since the rest constitute the unique primary key
func (rec *SetupStepRecord) Update(db *sql.DB) error {
  var err error
  if nil == rec.Step.Owner {
    _, err = db.Exec("UPDATE setup_steps SET done = $1 WHERE session_id = $2 AND setup_rule_id = $3 AND player_id IS NULL",
      rec.Step.Done, rec.SessionId, rec.Step.Rule.Id)
  } else {
    _, err = db.Exec("UPDATE setup_steps SET done = $1 WHERE session_id = $2 AND setup_rule_id = $3 AND player_id = $4",
      rec.Step.Done, rec.SessionId, rec.Step.Rule.Id, rec.Step.Owner.Id)
  }
  return err
}

type SetupStepRecordList struct {
  records []*SetupStepRecord
}

func NewSetupStepRecordList() *SetupStepRecordList {
  return &SetupStepRecordList{make([]*SetupStepRecord, 0)}
}

func (recs *SetupStepRecordList) List() []*game.SetupStep {
  steps := make([]*game.SetupStep, 0)
  for _,rec := range recs.records {
    steps = append(steps, rec.Step)
  }
  return steps
}

func (recs *SetupStepRecordList) SetRecords(records []*SetupStepRecord) {
  recs.records = records
}

func (recs *SetupStepRecordList) FindBySession(db *sql.DB, s *session.Session) error {
  recs.records = make([]*SetupStepRecord, 0)

  rows, err := db.Query("SELECT setup_rule_id, player_id, done FROM setup_steps WHERE session_id = $1", s.Id)
  if nil != err {
    return err
  }
  defer rows.Close()
  for rows.Next() {
    step := &game.SetupStep{}
    record := &SetupStepRecord{Step: step}
    if err := rows.Scan(&record.RuleId, &record.OwnerId, &record.Step.Done); nil != err {
      return err
    }
    recs.records = append(recs.records, record)
  }

  return nil
}

func (recs *SetupStepRecordList) AssociatePlayers(players []*game.Player) error {
  for _, rec := range recs.records {
    if !rec.OwnerId.Valid {
      rec.Step.Owner = nil
      continue
    }
    for _, player := range players {
      if player.Id == (int)(rec.OwnerId.Int64) {
        rec.Step.Owner = player
        break
      }
    }
  }
  return nil
}

func (recs *SetupStepRecordList) AssociateRules(rules []*game.SetupRule) error {
  for _, rec := range recs.records {
    for _, rule := range rules {
      if rule.Id == rec.RuleId {
        rec.Step.Rule = rule
        break
      }
    }
  }
  return nil
}
