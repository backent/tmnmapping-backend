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

