# carrotbb
bulletin board

## goals
- basic bulletin board implementation
- readable, clean code
- extensive test coverage
- explanatory comments
- multiple database backends with a shared interface
    - supported: json (very slow, prototyping only)
    - planned: postgresql
    - nice to have: mongodb, sqlite

## immediate todos
- postgresql backend
- error page template
- move all errors.New into global declarations, making them comparable
- ability to delete posts, comments, accounts
    - deletable users (deleted field, change name to deleted)
- return to referer once logged in
- stop refreshing sessions on every request

## long term todos
- csrf tokens
- store session token hashes in database
- more tests
- css

## nice to haves for the future
- image embeds
- moderation system
- cache for served content