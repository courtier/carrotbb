package database

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/rs/xid"
)

const CREATE_POSTS_TABLE = `CREATE TABLE IF NOT EXISTS posts (
	title			text,
	content			text,
	poster_id		bytea,
	id				bytea,
	comment_ids		bytea[],
	date_created	timestamp
);`

const CREATE_COMMENTS_TABLE = `CREATE TABLE IF NOT EXISTS comments (
	content			text,
	post_id			bytea,
	poster_id		bytea,
	id				bytea,
	date_created	timestamp
);`

const CREATE_USERS_TABLE = `CREATE TABLE IF NOT EXISTS users (
	name			text,
	id				bytea,
	password		text,
	date_joined		timestamp
);`

const CREATE_SESSIONS_TABLE = `CREATE TABLE IF NOT EXISTS sessions (
	hash			bytea,
	user_id			bytea,
	expiry			timestamp
);`

var createTables = []string{CREATE_POSTS_TABLE, CREATE_COMMENTS_TABLE, CREATE_USERS_TABLE, CREATE_SESSIONS_TABLE}

type PostgresDatabase struct {
	pool *pgxpool.Pool
}

func ConnectPostgres() (db *PostgresDatabase, err error) {
	config, err := pgxpool.ParseConfig(os.Getenv("POSTGRES_URL"))
	if err != nil {
		return nil, err
	}
	config.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		for _, statement := range createTables {
			_, err := conn.Exec(ctx, statement)
			if err != nil {
				log.Fatalln(err)
				return err
			}
		}
		return nil
	}
	pool, err := pgxpool.ConnectConfig(context.Background(), config)
	return &PostgresDatabase{pool}, err
}

func (p *PostgresDatabase) Disconnect() (err error) {
	p.pool.Close()
	return
}

func (p *PostgresDatabase) AddPost(title, content string, posterID xid.ID) (id xid.ID, err error) {
	return
}

func (p *PostgresDatabase) AddComment(content string, postID, posterID xid.ID) (id xid.ID, err error) {
	return
}

func (p *PostgresDatabase) AddUser(name, password string) (id xid.ID, err error) {
	return
}

func (p *PostgresDatabase) GetPost(id xid.ID) (post Post, err error) {
	return
}

func (p *PostgresDatabase) GetComment(id xid.ID) (comment Comment, err error) {
	return
}

func (p *PostgresDatabase) GetUser(id xid.ID) (user User, err error) {
	return
}

func (p *PostgresDatabase) FindUserByName(name string) (user User, err error) {
	return
}

func (p *PostgresDatabase) AllPosts() (posts []Post, err error) {
	return
}

func (p *PostgresDatabase) AllCommentsUnderPost(postID xid.ID) (comments []Comment, err error) {
	return
}

func (p *PostgresDatabase) GetPostPageData(postID xid.ID) (post Post, poster User, comments map[Comment]User, err error) {
	return
}
