package record

import (
  "database/sql"
  "testing"
  "github.com/rkbodenner/parallel_universe/game"
)

func TestSetupStepRecordList_AssociatePlayers(t *testing.T) {
  doug := &game.Player{Id: 42, Name: "Douglas"}
  players := []*game.Player{
    doug,
    &game.Player{Id: 666, Name: "Bob"},
  }
  ownedRecord  := &SetupStepRecord{OwnerId: sql.NullInt64{Int64: 42, Valid: true},  Step: &game.SetupStep{}}
  globalRecord := &SetupStepRecord{OwnerId: sql.NullInt64{           Valid: false}, Step: &game.SetupStep{}}
  recs := []*SetupStepRecord{ownedRecord, globalRecord}

  list := NewSetupStepRecordList()
  list.SetRecords(recs)

  err := list.AssociatePlayers(players)
  if nil != err {
    t.Fatal(err)
  }
  if ownedRecord.Step.Owner != doug {
    t.Fatal("Step should be owned by player with ID 42")
  }
  if globalRecord.Step.Owner != nil {
    t.Fatal("Step should not be onwned by a specific player")
  }
}

func TestSetupStepRecordList_AssociateRules(t *testing.T) {
  ruleInList := &game.SetupRule{Id: 42}
  ruleNotInList := &game.SetupRule{Id: 666}

  recordWithRule := &SetupStepRecord{RuleId: 42, Step: &game.SetupStep{}}
  recordWithoutRule := &SetupStepRecord{RuleId: 99, Step: &game.SetupStep{}}
  recs := []*SetupStepRecord{recordWithRule, recordWithoutRule}

  list := NewSetupStepRecordList()
  list.SetRecords(recs)

  err := list.AssociateRules([]*game.SetupRule{ruleInList, ruleNotInList})
  if nil != err {
    t.Fatal(err)
  }
  if recordWithRule.Step.Rule != ruleInList {
    t.Fatal("Step should have rule with ID 42")
  }
  if recordWithoutRule.Step.Rule != nil {
    t.Fatal("Step should have no associate rule")
  }
}
