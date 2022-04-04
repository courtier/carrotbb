# carrotbb
bulletin board

## goals
- basic bulletin board implementation
- readable, clean code
- extensive test coverage
- multiple database backends with a shared interface
    - supported: json (very slow, for prototyping only), postgresql

## immediate todos
- ability to delete posts, comments, accounts
    - deletable users (deleted field, change name to deleted)
- handle the logging of errors in a single function?
    - generate error template
    - zap the error

## long term todos
- csrf tokens
- css
- paging posts and comments

## nice to haves for the future
- image embeds
- quoting other comments
- moderation system

## notes
- forked xid to work with pgx without any hiccups
    - https://github.com/courtier/xid
        - todo: needs an array type

## setting up
- no docker
    - postgres:
        - `createuser --interactive`
        - `psql`
        - `ALTER USER user WITH PASSWORD 'password';`
        - `\q`
        - `createdb carrotbb`
    - fill the `.env` by looking at `exampledotenv.txt`
    - `go run .` or `go build .` then `./carrotbb`
- docker
    - coming soon