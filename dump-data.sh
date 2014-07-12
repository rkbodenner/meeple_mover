#!/bin/sh
pg_dump --data-only --column-inserts --no-owner --file data.psql meeple_mover
