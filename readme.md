# carrotbb
bulletin board

## goals
- basic bulletin board implementation
- readable, clean code
- high test coverage
- multiple database backends with a shared interface
    - supported: json (very slow, prototyping only)
    - planned: postgresql
    - nice to have: mongodb, sqlite

## immediate todos
- error page template
- move all errors.New into global declarations, making them comparable
- deletable users (deleted field, change name to deleted)
- return to referer once logged in

## long term todos
- csrf tokens
- store session token hashes in database
- more tests
- css

## nice to haves for the future
- image embeds
- moderation system
- cache for served content