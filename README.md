meeple_mover
============

Web service backing goboard, an app for faster multiplayer boardgame setup.

## Setup

1. go get

    Everything you need is imported in `meeple_mover.go`, so `go get` it.

2. Install PostgreSQL

3. Create the database and schema

    `pg_restore schema.psql`

## Running in Heroku

1. Create Heroku app

    `heroku create -b https://github.com/kr/heroku-buildpack-go.git`

2. Tell it where to find goboard, for CORS purposes

    `heroku config:set MEEPLE_MOVER_ORIGIN_URL=http://example.com`

3. `git push heroku master`
