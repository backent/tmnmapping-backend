package auth

import "time"

type RepositoryAuthInterface interface {
	Issue(payload string, duration time.Duration) (string, error)
	Validate(tokenString string) (int, bool)
}

