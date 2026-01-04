package auth

type RepositoryAuthInterface interface {
	Issue(payload string) (string, error)
	Validate(tokenString string) (int, bool)
}

