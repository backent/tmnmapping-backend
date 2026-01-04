package user

import (
	"context"
	"database/sql"

	"github.com/malikabdulaziz/tmn-backend/models"
)

type RepositoryUserInterface interface {
	FindById(ctx context.Context, tx *sql.Tx, id int) (models.User, error)
	FindByUsername(ctx context.Context, tx *sql.Tx, username string) (models.User, error)
}

