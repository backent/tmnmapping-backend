package auth

import (
	"context"

	webAuth "github.com/malikabdulaziz/tmn-backend/web/auth"
)

type ServiceAuthInterface interface {
	Login(ctx context.Context, username, password string) (webAuth.LoginResponse, string)
}

