package record

import (
  "database/sql"
)

type Record interface {
  Create(*sql.DB) error
  Find(db *sql.DB, id int) error
}
