package account

import (
	"context"
	"database/sql"
	"errors"
	"github.com/go-kit/log"
)

var RepoError = errors.New("Unable to handle request")

type repo struct {
	db     *sql.DB
	logger log.Logger
}

func (r repo) CreateUser(ctx context.Context, user User) error {
	sql := `INSERT INTO users (id, email, password) 
		VALUES($1, $2, $3)`

	if user.Email == "" || user.Password == "" {
		return RepoError
	}

	_, err := r.db.Exec(sql, user.ID, user.Email, user.Password)
	if err != nil {
		return err
	}
	return nil
}

func (r repo) Getuser(ctx context.Context, ID string) (string, error) {
	var email string
	err := r.db.QueryRow(`SELECT email FROM users WHERE id=$1`, ID).Scan(&email)
	if err != nil {
		return "", RepoError
	}

	return email, nil
}

func NewRepo(db *sql.DB, logger log.Logger) Repository {
	_, _ = db.Exec(`CREATE TABLE IF NOT EXISTS users (
    id varchar,
    email varchar,
    password varchar
	)`)

	return &repo{
		db:     db,
		logger: logger,
	}
}
