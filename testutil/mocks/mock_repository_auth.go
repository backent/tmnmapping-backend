package mocks

import "github.com/stretchr/testify/mock"

// MockRepositoryAuth implements repositories/auth.RepositoryAuthInterface
type MockRepositoryAuth struct {
	mock.Mock
}

func (m *MockRepositoryAuth) Issue(payload string) (string, error) {
	args := m.Called(payload)
	return args.String(0), args.Error(1)
}

func (m *MockRepositoryAuth) Validate(tokenString string) (int, bool) {
	args := m.Called(tokenString)
	return args.Int(0), args.Bool(1)
}
