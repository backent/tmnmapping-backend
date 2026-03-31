package auth

import (
	"context"
	"database/sql"
	"os"
	"strconv"
	"time"

	"github.com/malikabdulaziz/tmn-backend/exceptions"
	"github.com/malikabdulaziz/tmn-backend/helpers"
	repositoriesAuth "github.com/malikabdulaziz/tmn-backend/repositories/auth"
	repositoriesUser "github.com/malikabdulaziz/tmn-backend/repositories/user"
	webAuth "github.com/malikabdulaziz/tmn-backend/web/auth"
)

const rememberMeDuration = 48 * time.Hour

type ServiceAuthImpl struct {
	*sql.DB
	repositoriesAuth.RepositoryAuthInterface
	repositoriesUser.RepositoryUserInterface
	defaultDuration time.Duration
}

func NewServiceAuthImpl(db *sql.DB, repositoriesAuth repositoriesAuth.RepositoryAuthInterface, repositoriesUser repositoriesUser.RepositoryUserInterface) ServiceAuthInterface {
	tokenLifeTime, err := strconv.Atoi(os.Getenv("APP_TOKEN_EXPIRE_IN_SEC"))
	if err != nil {
		tokenLifeTime = 3600 // default 1 hour
	}

	return &ServiceAuthImpl{
		DB:                      db,
		RepositoryAuthInterface: repositoriesAuth,
		RepositoryUserInterface: repositoriesUser,
		defaultDuration:         time.Second * time.Duration(tokenLifeTime),
	}
}

func (implementation *ServiceAuthImpl) Login(ctx context.Context, username, password, ipAddress string, remember bool) (webAuth.LoginResponse, string, int) {
	tx, err := implementation.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	user, err := implementation.RepositoryUserInterface.FindByUsername(ctx, tx, username)
	if err != nil {
		panic(exceptions.NewBadRequestError("invalid credentials"))
	}

	if !helpers.CheckPassword(password, user.Password) {
		panic(exceptions.NewBadRequestError("invalid credentials"))
	}

	// Record login log
	err = implementation.RepositoryUserInterface.CreateLoginLog(ctx, tx, user.Id, ipAddress)
	helpers.PanicIfError(err)

	// Choose token duration: 2 days when remember=true, default from config otherwise
	duration := implementation.defaultDuration
	if remember {
		duration = rememberMeDuration
	}

	stringUserId := strconv.Itoa(user.Id)
	token, err := implementation.RepositoryAuthInterface.Issue(stringUserId, duration)
	helpers.PanicIfError(err)

	return webAuth.LoginResponse{
		User: webAuth.UserResponse{
			Id:       user.Id,
			Username: user.Username,
			Name:     user.Name,
			Role:     user.Role,
		},
	}, token, int(duration.Seconds())
}

