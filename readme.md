# carrotbb
bulletin board

# notes
- multiple database backends that all implement the same interface
- for auth we will use session tokens
- templating will be used to render pages
- once the basics are done we will start building more advanced things on top, like image embeds, moderating system etc.

## nice to haves for the future
- error handling by having a separate error page where we display the error in a nice manner using templates
- csrf tokens

## database backends
- json
- postgresql (planned)
- mongodb (maybe)
- sqlite (probably never, but nice to have)

## uncertain
- store hashes of session tokens in db, or have all session tokens in memory
    - for now session tokens will live only in memory

## todo
- move all errors.New into global declarations, making them comparable
- unexport unnecessarily exported functions