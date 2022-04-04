package database

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
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

//go:embed create_tables.sql
var createTables string

type PostgresDatabase struct {
	pool *pgxpool.Pool
}

func ConnectPostgres() (db *PostgresDatabase, err error) {
	config, err := pgxpool.ParseConfig(buildPostgresURL())
	if err != nil {
		return nil, err
	}
	config.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		_, err := conn.Exec(ctx, createTables)
		if err != nil {
			log.Fatalln(err)
			return err
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

func (p *PostgresDatabase) GetPost(id xid.ID) (post Post, err error) {
	err = p.pool.QueryRow(context.Background(),
		`SELECT * FROM posts WHERE id=$1`, id).
		Scan(&post.Title, &post.Content, &post.PosterID, &post.ID, &post.CommentIDs, &post.DateCreated)
	if err == pgx.ErrNoRows {
		err = ErrNoPostFoundByID
	}
	return
}

func (p *PostgresDatabase) GetComment(id xid.ID) (comment Comment, err error) {
	err = p.pool.QueryRow(context.Background(),
		`SELECT * FROM comments WHERE id=$1`, id).
		Scan(&comment.Content, &comment.PostID, &comment.PosterID, &comment.ID, &comment.DateCreated)
	if err == pgx.ErrNoRows {
		err = ErrNoCommentFoundByID
	}
	return
}

func (p *PostgresDatabase) GetUser(id xid.ID) (user User, err error) {
	err = p.pool.QueryRow(context.Background(),
		`SELECT * FROM users WHERE id=$1`, id).
		Scan(&user.Name, &user.ID, &user.Password, &user.DateJoined)
	if err == pgx.ErrNoRows {
		err = ErrNoUserFoundByID
	}
	return
}

func (p *PostgresDatabase) FindUserByName(name string) (user User, err error) {
	err = p.pool.QueryRow(context.Background(),
		`SELECT * FROM users WHERE name=$1`, name).
		Scan(&user.Name, &user.ID, &user.Password, &user.DateJoined)
	if err == pgx.ErrNoRows {
		err = ErrNoUserFoundByName
	}
	return
}

// TODO: paging
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
			log.Println(err)
			return
		}
		posts = append(posts, post)
	}
	return
}

func (p *PostgresDatabase) PagePosts(start, end int) (posts []Post, err error) {
	rows, err := p.pool.Query(context.TODO(), "SELECT * FROM posts ORDER BY date_created DESC LIMIT $1 OFFSET $2", end-start, start)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var post Post
		err = rows.Scan(&post.Title, &post.Content, &post.PosterID, &post.ID, &post.CommentIDs, &post.DateCreated)
		if err != nil {
			log.Println(err)
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
	batch.Queue(`SELECT * FROM comments WHERE id = ANY ($1) ORDER BY date_created ASC`, post.CommentIDs)
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
	users = make(map[xid.ID]User)
	commenterBatch := &pgx.Batch{}
	for _, comment := range comments {
		commenterBatch.Queue(`SELECT * FROM users WHERE id=$1`, comment.PosterID)
	}
	commenterBR := p.pool.SendBatch(context.Background(), commenterBatch)
	defer commenterBR.Close()
	for _, comment := range comments {
		var user User
		if err = commenterBR.QueryRow().Scan(&user.Name, &user.ID, &user.Password, &user.DateJoined); err != nil {
			return
		}
		users[comment.ID] = user
	}
	return
}

func buildPostgresURL() string {
	return fmt.Sprintf("postgres://%s:%s@localhost:5432/%s", os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASSWORD"), os.Getenv("POSTGRES_DB"))
}
