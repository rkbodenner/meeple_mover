package record

import (
  "fmt"
  "database/sql"
  "testing"
  "github.com/rkbodenner/parallel_universe/game"
)

// These tests must be run on a database with a schema defined in schema.psql, but otherwise empty

var db *sql.DB

// TODO: This is available in Go 1.4
//func TestMain(m *testing.M) {
func setup(t* testing.T) {
  connectString := fmt.Sprintf("user=ralph dbname=%s sslmode=disable", "meeple_mover_test")

  var err error
  db, err = sql.Open("postgres", connectString)
  if err != nil {
    t.Fatal(err)
  }

  _, err = db.Exec("DELETE FROM players")
  if err != nil {
    t.Fatal(err)
  }
//  defer db.Close()

//  os.Exit(m.run())
}

func TestPlayer_Create(t *testing.T) {
  setup(t)
  defer db.Close()

  bogusId := 0
  player := &game.Player{Id: bogusId, Name: "Bob"}
  playerRecord := &PlayerRecord{player}

  err := playerRecord.Create(db)
  if nil != err {
    t.Fatal(err)
  }
  if playerRecord.Player.Id == bogusId {
    t.Fatal("ID not updated in player object on create")
  }
}

func TestPlayer_Find(t *testing.T) {
  setup(t)
  defer db.Close()

  _, err := db.Exec("INSERT INTO players VALUES(41, 'Joe')")
  if nil != err {
    t.Fatal(err)
  }

  player := &game.Player{}
  playerRecord := &PlayerRecord{player}
  err = playerRecord.Find(db, 41)
  if nil != err {
    t.Fatal(err)
  }

  if player.Id != 41 {
    t.Fatal("ID not updated in player object on find")
  }
  if player.Name != "Joe" {
    t.Fatal("Name not updated in player object on find")
  }
}

func TestPlayer_Delete(t *testing.T) {
  setup(t)
  defer db.Close()

  _, err := db.Exec("INSERT INTO players VALUES(42, 'Joe')")
  if nil != err {
    t.Fatal(err)
  }

  player := &game.Player{Id: 42}
  playerRecord := &PlayerRecord{player}
  err = playerRecord.Delete(db)
  if nil != err {
    t.Fatal(err)
  }
}
