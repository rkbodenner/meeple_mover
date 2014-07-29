package record

import (
  "database/sql"
  "errors"
  "fmt"
  _ "github.com/lib/pq"
  "github.com/rkbodenner/parallel_universe/game"
)

type GameRecord struct {
  g *game.Game
}

func NewGameRecord(g *game.Game) *GameRecord {
  return &GameRecord{g: g}
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

  rec.g.Id = (uint)(id)
  rec.g.Name = name
  rec.g.MinPlayers = minPlayers
  rec.g.MaxPlayers = maxPlayers

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
