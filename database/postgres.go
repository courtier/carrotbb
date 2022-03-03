package database

import (
	"context"
	"errors"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/rs/xid"
)

var (
	ErrMistmatchedRowsAffected = errors.New("errors affected does not match desired number")
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
		// conn.ConnInfo().RegisterDataType(pgtype.DataType{
		// 	Value: xid.ID{},
		// 	Name:  "uuid",
		// 	OID:   pgtype.UUIDOID,
		// })
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
	ct, err := p.pool.Exec(context.Background(),
		`INSERT INTO posts(title, content, poster_id, id, comment_ids, date_created)
	VALUES ($1, $2, $3, $4, $5, $6)
	ON CONFLICT DO NOTHING`, title, content, posterID, id, []xid.ID{}, time.Now())
	if ct.RowsAffected() != 1 {
		err = ErrMistmatchedRowsAffected
	}
	return
}

func (p *PostgresDatabase) AddComment(content string, postID, posterID xid.ID) (id xid.ID, err error) {
	id = xid.New()
	batch := &pgx.Batch{}
	batch.Queue(`INSERT INTO comments(content, post_id, poster_id, id, date_created)
	VALUES ($1, $2, $3, $4, $5)
	ON CONFLICT DO NOTHING`, content, postID, posterID, id, time.Now())
	batch.Queue(`UPDATE posts SET comment_ids = array_append(comment_ids, $1) WHERE id=$2`, id, postID)
	br := p.pool.SendBatch(context.Background(), batch)
	defer br.Close()
	ct, err := br.Exec()
	if err != nil {
		return
	}
	if ct.RowsAffected() != 1 {
		err = ErrMistmatchedRowsAffected
		return
	}
	ct, err = br.Exec()
	if err != nil {
		return
	}
	if ct.RowsAffected() != 1 {
		err = ErrMistmatchedRowsAffected
	}
	return
}

func (p *PostgresDatabase) AddUser(name, password string) (id xid.ID, err error) {
	id = xid.New()
	ct, err := p.pool.Exec(context.Background(),
		`INSERT INTO users(name, id, password, date_joined)
	VALUES ($1, $2, $3, $4)
	ON CONFLICT DO NOTHING`, name, id, password, time.Now())
	if ct.RowsAffected() != 1 {
		err = ErrMistmatchedRowsAffected
	}
	return
}

func (p *PostgresDatabase) AddSession(tokenHash string, userID xid.ID, expiry time.Time) (err error) {
	return
}

func (p *PostgresDatabase) GetPost(id xid.ID) (post Post, err error) {
	err = p.pool.QueryRow(context.Background(),
		`SELECT * FROM posts WHERE id=$1`, id).
		Scan(&post.Title, &post.Content, &post.PosterID, &post.ID, &post.CommentIDs, &post.DateCreated)
	return
}

func (p *PostgresDatabase) GetComment(id xid.ID) (comment Comment, err error) {
	err = p.pool.QueryRow(context.Background(),
		`SELECT * FROM comments WHERE id=$1`, id).
		Scan(&comment.Content, &comment.PostID, &comment.PosterID, &comment.ID, &comment.DateCreated)
	return
}

func (p *PostgresDatabase) GetUser(id xid.ID) (user User, err error) {
	err = p.pool.QueryRow(context.Background(),
		`SELECT * FROM users WHERE id=$1`, id).
		Scan(&user.Name, &user.ID, &user.Password, &user.DateJoined)
	return
}

func (p *PostgresDatabase) FindUserByName(name string) (user User, err error) {
	err = p.pool.QueryRow(context.Background(),
		`SELECT * FROM users WHERE name=$1`, name).
		Scan(&user.Name, &user.ID, &user.Password, &user.DateJoined)
	return
}

func (p *PostgresDatabase) AllPosts() (posts []Post, err error) {
	rows, err := p.pool.Query(context.TODO(), "SELECT * FROM posts")
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var post Post
		err = rows.Scan(&post.Title, &post.Content, &post.PosterID, &post.ID, &post.CommentIDs, &post.DateCreated)
		if err != nil {
			return
		}
		posts = append(posts, post)
	}
	return
}

func (p *PostgresDatabase) GetPostPageData(postID xid.ID) (post Post, poster User, comments []Comment, users map[xid.ID]User, err error) {
	post, err = p.GetPost(postID)
	if err != nil {
		return
	}
	batch := &pgx.Batch{}
	batch.Queue(`SELECT * FROM users WHERE id=$1`, post.PosterID)
	batch.Queue(`SELECT * FROM comments WHERE id IN $1 ORDER BY date_created ASC`, post.CommentIDs)
	br := p.pool.SendBatch(context.Background(), batch)
	defer br.Close()
	err = br.QueryRow().Scan(&poster.Name, &poster.ID, &poster.Password, &poster.DateJoined)
	if err != nil {
		return
	}
	rows, err := br.Query()
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var comment Comment
		err = rows.Scan(&comment.Content, &comment.PostID, &comment.PosterID, &comment.ID, &comment.DateCreated)
		if err != nil {
			return
		}
		comments = append(comments, comment)
	}
	commenterBatch := &pgx.Batch{}
	for _, comment := range comments {
	}
	return
}
