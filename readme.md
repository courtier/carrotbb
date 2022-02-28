# carrotbb
bulletin board

## goals
- basic bulletin board implementation
- readable, clean code
- high test coverage

## immediate todos
- error page template
- move all errors.New into global declarations, making them comparable
- deletable users (deleted field, change name to deleted)
- migrate to gorm

## long term todos
- csrf tokens
- store session token hashes in database
- more tests

## nice to haves for the future
- image embeds
- moderation
- cache for served content