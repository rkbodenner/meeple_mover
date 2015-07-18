/*

Store records for a game in the database.

Useful for persisting game data after prototyping the rules in Go.

*/

package main

import (
  "database/sql"
  "flag"
  "fmt"
  "os"
  _ "github.com/lib/pq"
  "github.com/rkbodenner/meeple_mover/record"
  "github.com/rkbodenner/parallel_universe/collection"
)

func main() {
  var gameName string
  flag.StringVar(&gameName, "game", "", "Name of the game to store in the database")
  var databaseName string
  flag.StringVar(&databaseName, "dbname", "meeple_mover_test", "Name of the database")
  var dryRun bool
  flag.BoolVar(&dryRun, "dry-run", false, "Run without creating any records")
  flag.Parse()

  if "" == gameName {
    flag.Usage()
    os.Exit(1)
  }

  fmt.Printf("Searching for %s...\n", gameName)

  connectString := fmt.Sprintf("user=ralph dbname=%s sslmode=disable", databaseName)
  db, err := sql.Open("postgres", connectString)
  if nil != err {
    fmt.Print(err)
  }
  defer db.Close()

  shelf := collection.NewCollection().Games
  for _, game := range shelf {
    if gameName == game.Name {
      fmt.Printf("Found %s in the collection\n", gameName)
      rec := record.NewGameRecord(game)
      if !dryRun {
        err := rec.Create(db)
        if nil != err {
          fmt.Println(err)
        } else {
          fmt.Printf("Stored %s as ID %d\n", gameName, rec.Game.Id)
        }
      }
    }
  }
}
