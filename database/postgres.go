package database

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/rs/xid"
)

const CREATE_POSTS_TABLE = `CREATE TABLE IF NOT EXISTS posts (
	title			text,
	content			text,
	poster_id		bytea,
	id				bytea PRIMARY KEY,
	comment_ids		bytea[],
	date_created	timestamp
);`

const CREATE_COMMENTS_TABLE = `CREATE TABLE IF NOT EXISTS comments (
	content			text,
	post_id			bytea,
	poster_id		bytea,
	id				bytea PRIMARY KEY,
	date_created	timestamp
);`

const CREATE_USERS_TABLE = `CREATE TABLE IF NOT EXISTS users (
	name			text UNIQUE,
	id				bytea,
	password		text PRIMARY KEY,
	date_joined		timestamp
);`

const CREATE_SESSIONS_TABLE = `CREATE TABLE IF NOT EXISTS sessions (
	hash			bytea PRIMARY KEY,
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
	id = xid.New()
	_, err = p.pool.Exec(context.Background(),
		`INSERT INTO posts(title, content, poster_id, id, comment_ids, date_created)
	VALUES ($1, $2, $3, $3, $4, $5, $6)
	ON CONFLICT DO NOTHING`, title, content, posterID.Bytes(), id.Bytes(), [][]byte{}, time.Now())
	return
}

func (p *PostgresDatabase) AddComment(content string, postID, posterID xid.ID) (id xid.ID, err error) {
	id = xid.New()
	_, err = p.pool.Exec(context.Background(),
		`INSERT INTO posts(content, post_id, poster_id, id, date_created)
	VALUES ($1, $2, $3, $3, $4, $5)
	ON CONFLICT DO NOTHING`, content, postID.Bytes(), posterID.Bytes(), id.Bytes(), time.Now())
	return
}

func (p *PostgresDatabase) AddUser(name, password string) (id xid.ID, err error) {
	id = xid.New()
	_, err = p.pool.Exec(context.Background(),
		`INSERT INTO posts(name, id, password, date_joined)
	VALUES ($1, $2, $3, $3, $4)
	ON CONFLICT DO NOTHING`, name, id.Bytes(), password, time.Now())
	return
}

func (p *PostgresDatabase) GetPost(id xid.ID) (post Post, err error) {
	var idBytes []byte
	var posterIDBytes []byte
	var commentIDBytesArray [][]byte
	err = p.pool.QueryRow(context.Background(),
		`SELECT * FROM posts WHERE id=$1`, id.Bytes()).
		Scan(post.Title, post.Content, posterIDBytes, idBytes, commentIDBytesArray, post.DateCreated)
	if err != nil {
		return
	}
	if post.ID, err = xid.FromBytes(idBytes); err != nil {
		return
	}
	if post.PosterID, err = xid.FromBytes(posterIDBytes); err != nil {
		return
	}
	if len(commentIDBytesArray) > 0 {
		post.CommentIDs = make([]xid.ID, len(commentIDBytesArray))
		for i, cIDBytes := range commentIDBytesArray {
			if post.CommentIDs[i], err = xid.FromBytes(cIDBytes); err != nil {
				return
			}
		}
	}
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
