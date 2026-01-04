package auth

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/malikabdulaziz/tmn-backend/exceptions"
	"github.com/malikabdulaziz/tmn-backend/helpers"
	repositoriesAuth "github.com/malikabdulaziz/tmn-backend/repositories/auth"
	repositoriesUser "github.com/malikabdulaziz/tmn-backend/repositories/user"
	webAuth "github.com/malikabdulaziz/tmn-backend/web/auth"
)

type ServiceAuthImpl struct {
	*sql.DB
	repositoriesAuth.RepositoryAuthInterface
	repositoriesUser.RepositoryUserInterface
}

func NewServiceAuthImpl(db *sql.DB, repositoriesAuth repositoriesAuth.RepositoryAuthInterface, repositoriesUser repositoriesUser.RepositoryUserInterface) ServiceAuthInterface {
	return &ServiceAuthImpl{
		DB:                      db,
		RepositoryAuthInterface: repositoriesAuth,
		RepositoryUserInterface: repositoriesUser,
	}
}

func (implementation *ServiceAuthImpl) Login(ctx context.Context, username, password string) (webAuth.LoginResponse, string) {
	tx, err := implementation.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	// Pure business logic - find user and verify password
	user, err := implementation.RepositoryUserInterface.FindByUsername(ctx, tx, username)
	if err != nil {
		panic(exceptions.NewBadRequestError("invalid credentials"))
	}

	if !helpers.CheckPassword(password, user.Password) {
		panic(exceptions.NewBadRequestError("invalid credentials"))
	}

	// Generate JWT token
	stringUserId := strconv.Itoa(user.Id)
	token, err := implementation.RepositoryAuthInterface.Issue(stringUserId)
	helpers.PanicIfError(err)

	// Return user data and token separately
	return webAuth.LoginResponse{
		User: webAuth.UserResponse{
			Id:       user.Id,
			Username: user.Username,
			Name:     user.Name,
			Role:     user.Role,
		},
	}, token
}

