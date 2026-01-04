package exceptions

type Unauthorized struct {
	Error string
}

func NewUnAuthorized(error string) Unauthorized {
	return Unauthorized{Error: error}
}

