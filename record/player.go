package record

import (
  "database/sql"
  _ "github.com/lib/pq"
  "github.com/rkbodenner/parallel_universe/game"
)

type PlayerRecord struct {
  Player *game.Player
}

func (playerRec *PlayerRecord) Create(db *sql.DB) error {
  err := db.QueryRow("INSERT INTO players(id, name) VALUES(default, $1) RETURNING id", playerRec.Player.Name).Scan(&playerRec.Player.Id)
  return err
}

func (playerRec *PlayerRecord) Find(db *sql.DB, id int) error {
  err := db.QueryRow("SELECT name FROM players WHERE id = $1", id).Scan(&playerRec.Player.Name)
  if nil == err {
    playerRec.Player.Id = id
  }
  return err
}

func (playerRec *PlayerRecord) Delete(db *sql.DB) error {
  _, err := db.Exec("DELETE FROM players WHERE id=$1", playerRec.Player.Id)
  return err
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
    players = append(players, rec.Player)
  }
  return players
}
