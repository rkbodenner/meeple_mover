package record

import (
  "database/sql"
  _ "github.com/lib/pq"
  "github.com/rkbodenner/parallel_universe/game"
)

type PlayerRecord struct {
  p *game.Player
}

type PlayerRecordList struct {
  records []*PlayerRecord
}

func (recs *PlayerRecordList) FindAll(db *sql.DB) error {
  rows, err := db.Query("SELECT * FROM players")
  if nil != err {
    return err
  }
  defer rows.Close()

  recs.records = make([]*PlayerRecord, 0)
  for rows.Next() {
    var name string
    var id int
    if err := rows.Scan(&id, &name); err != nil {
      return err
    }
    recs.records = append(recs.records, &PlayerRecord{&game.Player{id, name}})
  }

  return nil
}

func (recs *PlayerRecordList) List() []*game.Player {
  players := make([]*game.Player, 0)
  for _, rec := range recs.records {
    players = append(players, rec.p)
  }
  return players
}
