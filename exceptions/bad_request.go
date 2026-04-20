package exceptions

type BadRequestError struct {
	Error  string
	Extras interface{}
}

func NewBadRequestError(error string) BadRequestError {
	return BadRequestError{Error: error}
}

// Alias for consistency
func NewBadRequest(error string) BadRequestError {
	return BadRequestError{Error: error}
}

// NewBadRequestWithExtras returns a BadRequestError carrying structured
// detail that the panic handler will surface as WebResponse.Extras.
func NewBadRequestWithExtras(error string, extras interface{}) BadRequestError {
	return BadRequestError{Error: error, Extras: extras}
}

