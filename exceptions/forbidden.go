package exceptions

// ForbiddenError is raised when a request is authenticated but the caller's role
// does not permit the action. Distinct from Unauthorized (403 vs 401) so the
// frontend can tell "log in again" apart from "you may not do this".
type ForbiddenError struct {
	Error string
}

func NewForbidden(error string) ForbiddenError {
	return ForbiddenError{Error: error}
}
