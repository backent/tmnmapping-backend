package user

import (
	"context"
	"database/sql"

	"github.com/malikabdulaziz/tmn-backend/models"
)

type RepositoryUserImpl struct {
}

func NewRepositoryUserImpl() RepositoryUserInterface {
	return &RepositoryUserImpl{}
}

func (repository *RepositoryUserImpl) FindById(ctx context.Context, tx *sql.Tx, id int) (models.User, error) {
	SQL := "SELECT id, username, name, email, password, role FROM " + models.UserTable + " WHERE id = $1"
	rows, err := tx.QueryContext(ctx, SQL, id)
	if err != nil {
		return models.User{}, err
	}
	defer rows.Close()

	user := models.NullAbleUser{}
	if rows.Next() {
		err := rows.Scan(&user.Id, &user.Username, &user.Name, &user.Email, &user.Password, &user.Role)
		if err != nil {
			return models.User{}, err
		}
		return models.NullAbleUserToUser(user), nil
	} else {
		return models.User{}, sql.ErrNoRows
	}
}

func (repository *RepositoryUserImpl) FindByUsername(ctx context.Context, tx *sql.Tx, username string) (models.User, error) {
	SQL := "SELECT id, username, name, email, password, role FROM " + models.UserTable + " WHERE username = $1"
	rows, err := tx.QueryContext(ctx, SQL, username)
	if err != nil {
		return models.User{}, err
	}
	defer rows.Close()

	user := models.NullAbleUser{}
	if rows.Next() {
		err := rows.Scan(&user.Id, &user.Username, &user.Name, &user.Email, &user.Password, &user.Role)
		if err != nil {
			return models.User{}, err
		}
		return models.NullAbleUserToUser(user), nil
	} else {
		return models.User{}, sql.ErrNoRows
	}
}

func (repository *RepositoryUserImpl) CreateLoginLog(ctx context.Context, tx *sql.Tx, userId int, ipAddress string) error {
	SQL := "INSERT INTO " + models.UserLoginLogTable + " (user_id, ip_address) VALUES ($1, $2)"
	_, err := tx.ExecContext(ctx, SQL, userId, ipAddress)
	return err
}

func (repository *RepositoryUserImpl) FindLastLoginByUserId(ctx context.Context, tx *sql.Tx, userId int) (models.UserLoginLog, error) {
	SQL := "SELECT id, user_id, logged_in_at, ip_address FROM " + models.UserLoginLogTable + " WHERE user_id = $1 ORDER BY logged_in_at DESC LIMIT 1 OFFSET 1"
	rows, err := tx.QueryContext(ctx, SQL, userId)
	if err != nil {
		return models.UserLoginLog{}, err
	}
	defer rows.Close()

	log := models.NullAbleUserLoginLog{}
	if rows.Next() {
		err := rows.Scan(&log.Id, &log.UserId, &log.LoggedInAt, &log.IPAddress)
		if err != nil {
			return models.UserLoginLog{}, err
		}
		return models.NullAbleUserLoginLogToUserLoginLog(log), nil
	}
	return models.UserLoginLog{}, sql.ErrNoRows
}

