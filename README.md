meeple_mover
============

Web service backing [goboard](https://github.com/rkbodenner/goboard), an app for faster multiplayer boardgame setup.

## Setup

1. Fetch dependencies

    Everything you need is imported in `meeple_mover.go`, so `go get` it.

2. Install PostgreSQL

3. Create the database and schema

    `pg_restore schema.psql`

## Running in Heroku

### Create Heroku app
`heroku create -b https://github.com/kr/heroku-buildpack-go.git`

### Tell it where to find goboard, for CORS purposes
`heroku config:set MEEPLE_MOVER_ORIGIN_URL=http://example.com`

### Initialize data
Copy the contents of your local DB to Heroku:

`heroku pg:push $LOCAL_DB_NAME $HEROKU_DB_NAME::$HEROKU_DB_COLOR`

Where:
* LOCAL_DB_NAME is the name of your local Postgres database for meeple_mover (default: meeple_mover)
* HEROKU_APP_NAME is your DB's name on Heroku
* HEROKU_DB_COLOR is the "color" of the DB on Heroku

You may have to `heroku pg:reset $HEROKU_DB_NAME::$HEROKU_DB_COLOR`, which blows away all your data on the Heroku DB.

### Deploy
`git push heroku master`

## Testing
### Unit tests
Run `go test` in each subpackage directory.

### Test database
It's useful to create a test database with fixture data in order to run integration tests on the [goboard](https://github.com/rkbodenner/goboard) web front-end for meeple_mover.

1. `psql -f schema-test.psql`
2. `psql -f data.psql`

Start meeple_mover with this database by setting an environment variable:
   MEEPLE_MOVER_DB_NAME go run meeple_mover.go
