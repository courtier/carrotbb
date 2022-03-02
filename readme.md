# carrotbb
bulletin board

## goals
- basic bulletin board implementation
- readable, clean code
- extensive test coverage
- explanatory comments
- multiple database backends with a shared interface
    - supported: json (very slow, for prototyping only)
    - planned: postgresql
    - nice to have: mongodb, sqlite

## immediate todos
- postgresql backend
- ability to delete posts, comments, accounts
    - deletable users (deleted field, change name to deleted)
- store session token hashes in database
- sort comments by their creation date
- switch from xid to uuid

## long term todos
- csrf tokens
- more tests
- css
- logging

## nice to haves for the future
- image embeds
- quoting other comments
- moderation system
- cache for served content
- docker file

## setting up
- postgres:
    - `createuser --interactive`
    - `psql`
    - `ALTER USER user WITH PASSWORD 'password';`
    - `\q`
    - `createdb carrotbb`
- fill the `.env` by looking at `exampledotenv.txt`
- `go run .` or `go build .` then `./carrotbb`