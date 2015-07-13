package record

import (
  "database/sql"
  "fmt"
  "os"
  "testing"
  "github.com/rkbodenner/parallel_universe/game"
)

// These tests must be run on a database with the schema defined in schema.psql

var db *sql.DB

func TestMain(m *testing.M) {
  connectString := fmt.Sprintf("user=ralph dbname=%s sslmode=disable", "meeple_mover_test")

  var err error
  db, err = sql.Open("postgres", connectString)
  if err != nil {
    fmt.Fprintf(os.Stderr, "Error opening database: %s", err)
    os.Exit(1)
  }

  _, err = db.Exec("DELETE FROM players")
  if err != nil {
    fmt.Fprintf(os.Stderr, "Error truncating players table: %s", err)
    os.Exit(1)
  }
  defer db.Close()

  os.Exit(m.Run())
}

func TestPlayer_Create(t *testing.T) {
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
