# carrotbb
bulletin board

# notes
- multiple database backends with one shared frontend,
which is a interface that handles things such as rules (for usernames, content etc.) then passes them onto the backing database, like http middleware
- for auth we will use jwt, and we will have a expired token list in the database
- templating will be used to render pages
- once the basics are done we will start building more advanced things on top, like embedding images, having an admin system, etc.

## nice to haves for the future
- error handling by having a separate error page where we display the error in a nice manner using templates

## database backends
- json
- postgresql (planned)
- sqlite (probably never, but nice to have)

## todo
- jwt expire list in the database