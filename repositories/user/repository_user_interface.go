package user

import (
	"context"
	"database/sql"

	"github.com/malikabdulaziz/tmn-backend/models"
)

type RepositoryUserInterface interface {
	Create(ctx context.Context, tx *sql.Tx, user models.User) (models.User, error)
	FindAll(ctx context.Context, tx *sql.Tx, take int, skip int, orderBy string, orderDirection string, search string) ([]models.User, error)
	CountAll(ctx context.Context, tx *sql.Tx, search string) (int, error)
	FindById(ctx context.Context, tx *sql.Tx, id int) (models.User, error)
	FindByUsername(ctx context.Context, tx *sql.Tx, username string) (models.User, error)
	Update(ctx context.Context, tx *sql.Tx, user models.User) (models.User, error)
	UpdatePassword(ctx context.Context, tx *sql.Tx, id int, hashedPassword string) error
	Delete(ctx context.Context, tx *sql.Tx, id int) error
	CountByRole(ctx context.Context, tx *sql.Tx, role string) (int, error)
	FindByRole(ctx context.Context, tx *sql.Tx, role string) ([]models.User, error)
	CanCreateOnBehalfOf(ctx context.Context, tx *sql.Tx, actorUserId int, ownerUserId int) (bool, error)
	CreateLoginLog(ctx context.Context, tx *sql.Tx, userId int, ipAddress string) error
	FindLastLoginByUserId(ctx context.Context, tx *sql.Tx, userId int) (models.UserLoginLog, error)
}
