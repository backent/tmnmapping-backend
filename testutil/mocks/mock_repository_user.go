package mocks

import (
	"context"
	"database/sql"

	"github.com/malikabdulaziz/tmn-backend/models"
	"github.com/stretchr/testify/mock"
)

// MockRepositoryUser implements repositories/user.RepositoryUserInterface
type MockRepositoryUser struct {
	mock.Mock
}

func (m *MockRepositoryUser) FindById(ctx context.Context, tx *sql.Tx, id int) (models.User, error) {
	args := m.Called(ctx, tx, id)
	return args.Get(0).(models.User), args.Error(1)
}

func (m *MockRepositoryUser) FindByUsername(ctx context.Context, tx *sql.Tx, username string) (models.User, error) {
	args := m.Called(ctx, tx, username)
	return args.Get(0).(models.User), args.Error(1)
}
