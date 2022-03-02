package database

import (
	"context"
	"os"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/rs/xid"
)

type PostgresDatabase struct {
	pool *pgxpool.Pool
}

func ConnectPostgres() (db *PostgresDatabase, err error) {
	config, err := pgxpool.ParseConfig(os.Getenv("POSTGRES_URL"))
	if err != nil {
		return nil, err
	}
	config.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		// create tables IF NOT EXIST here
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
