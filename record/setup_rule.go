package record

import (
  "database/sql"
  "fmt"
  _ "github.com/lib/pq"
  "github.com/rkbodenner/parallel_universe/game"
  "github.com/rkbodenner/parallel_universe/session"
)

type SetupRuleRecord struct {
  Rule *game.SetupRule
}

func (rec *SetupRuleRecord) Create(db *sql.DB) error {
  fmt.Println("Not actually creating a setup rule in the DB")
  return nil
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
  var err error

  var rows *sql.Rows
  rows, err = db.Query("SELECT id, description, each_player, details FROM setup_rules WHERE game_id = $1", g.Id)
  if nil != err {
    return err
  }
  defer rows.Close()
  for rows.Next() {
    rule := &game.SetupRule{}
    record := &SetupRuleRecord{Rule: rule}
    var eachPlayer bool
    var details sql.NullString
    if err := rows.Scan(&record.Rule.Id, &record.Rule.Description, &eachPlayer, &details); nil != err {
      return err
    }
    if eachPlayer {
      record.Rule.Arity = "Each player"
    } else {
      record.Rule.Arity = "Once"
    }
    if details.Valid {
      record.Rule.Details = details.String
    } else {
      record.Rule.Details = ""
    }
    rules.records = append(rules.records, record)
  }

  // Eager-load dependencies for the rules
  for _, parentRec := range rules.records {
    var depsRows *sql.Rows
    var depCount int = 0
    depsRows, err = db.Query("SELECT child_id FROM setup_rule_dependencies WHERE parent_id = $1", parentRec.Rule.Id)
    if nil != err {
      return err
    }
    defer depsRows.Close()
    for depsRows.Next() {
      var childId int
      if err := depsRows.Scan(&childId); nil != err {
        return err
      }
      depCount++
      for _, childRec := range rules.records {
        if childRec.Rule.Id == childId {
          childRec.Rule.Dependencies = append(childRec.Rule.Dependencies, parentRec.Rule)
          break  // Optimization assumes unique ID
        }
      }
    }
    if depCount > 0 {
      fmt.Printf("Loaded %d dependencies on rule #%d\n", depCount, parentRec.Rule.Id)
    }
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
