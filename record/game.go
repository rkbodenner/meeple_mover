package record

import (
  "database/sql"
  "errors"
  "fmt"
  _ "github.com/lib/pq"
  "github.com/rkbodenner/parallel_universe/game"
)

type GameRecord struct {
  Game *game.Game
}

func NewGameRecord(g *game.Game) *GameRecord {
  return &GameRecord{g}
}

func (rec *GameRecord) Find(db *sql.DB, id int) error {
  var err error

  var name string
  var minPlayers int
  var maxPlayers int
  err = db.QueryRow("SELECT name, min_players, max_players FROM games WHERE id = $1", id).Scan(&name, &minPlayers, &maxPlayers)
  if nil != err {
    return err
  }

  rec.Game.Id = (uint)(id)
  rec.Game.Name = name
  rec.Game.MinPlayers = minPlayers
  rec.Game.MaxPlayers = maxPlayers

  // Eager-load the associated game's setup rules
  rules := NewSetupRuleRecordList()
  err = rules.FindByGame(db, rec.Game)
  if err != nil {
    return err
  }
  fmt.Printf("Loaded %d setup rules\n", len(rules.List()))
  rec.Game.SetupRules = rules.List()

  return nil
}

func (rec *GameRecord) Create(db *sql.DB) error {
  err := db.QueryRow("INSERT INTO games(id, name, min_players, max_players) VALUES(default, $1, $2, $3) RETURNING id",
    rec.Game.Name, rec.Game.MinPlayers, rec.Game.MaxPlayers).Scan(&rec.Game.Id)
  if nil != err {
    return err
  }

  for _, rule := range rec.Game.SetupRules {
    ruleRec := &SetupRuleRecord{rule}
    err = ruleRec.Create(db)
    if nil != err {
      return err
    }
  }

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
    fmt.Printf("Loaded game %d\n", gameRec.Game.Id)
    recs.records = append(recs.records, gameRec)
  }

  return nil
}

func (recs *GameRecordList) List() []*game.Game {
  games := make([]*game.Game, 0)
  for _, rec := range recs.records {
    games = append(games, rec.Game)
  }
  return games
}
