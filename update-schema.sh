#!/bin/sh
pg_dump --create --schema-only --no-owner --no-privileges --file schema.psql meeple_mover
