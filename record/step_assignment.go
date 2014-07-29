package record

import (
  "database/sql"
  "fmt"
  _ "github.com/lib/pq"
  "github.com/rkbodenner/parallel_universe/game"
  "github.com/rkbodenner/parallel_universe/session"
)

type SetupStepAssignmentRecord struct {
  Session *session.Session
  Player *game.Player
  Rule *game.SetupRule
}

func (rec *SetupStepAssignmentRecord) Create(db *sql.DB) error {
  _, err := db.Exec("INSERT INTO setup_step_assignments(session_id, player_id, setup_rule_id) VALUES($1, $2, $3)",
    rec.Session.Id, rec.Player.Id, rec.Rule.Id)
  if nil != err {
    return err
  }
  fmt.Printf("Created step assignment of player #%d for rule #%d\n", rec.Player.Id, rec.Rule.Id)
  return nil
}

func (rec *SetupStepAssignmentRecord) Delete(db *sql.DB) error {
  _, err := db.Exec("DELETE FROM setup_step_assignments WHERE session_id=$1 AND player_id=$2 AND setup_rule_id=$3",
    rec.Session.Id, rec.Player.Id, rec.Rule.Id)
  if nil != err {
    return err
  }
  fmt.Printf("Deleted step assignment of player #%d for rule #%d\n", rec.Player.Id, rec.Rule.Id)
  return nil
}
