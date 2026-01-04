package exceptions

type BadRequestError struct {
	Error string
}

func NewBadRequestError(error string) BadRequestError {
	return BadRequestError{Error: error}
}

// Alias for consistency
func NewBadRequest(error string) BadRequestError {
	return BadRequestError{Error: error}
}

