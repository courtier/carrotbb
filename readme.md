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
- return to referer once logged in
- store session token hashes in database
- config file
    - store json db backing file, postgres user, port, ssl cert folder etc.

## long term todos
- csrf tokens
- more tests
- css
- logging

## nice to haves for the future
- image embeds
- moderation system
- cache for served content