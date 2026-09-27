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

func (m *MockRepositoryUser) Create(ctx context.Context, tx *sql.Tx, user models.User) (models.User, error) {
	args := m.Called(ctx, tx, user)
	return args.Get(0).(models.User), args.Error(1)
}

func (m *MockRepositoryUser) FindAll(ctx context.Context, tx *sql.Tx, take int, skip int, orderBy string, orderDirection string, search string) ([]models.User, error) {
	args := m.Called(ctx, tx, take, skip, orderBy, orderDirection, search)
	return args.Get(0).([]models.User), args.Error(1)
}

func (m *MockRepositoryUser) CountAll(ctx context.Context, tx *sql.Tx, search string) (int, error) {
	args := m.Called(ctx, tx, search)
	return args.Int(0), args.Error(1)
}

func (m *MockRepositoryUser) Update(ctx context.Context, tx *sql.Tx, user models.User) (models.User, error) {
	args := m.Called(ctx, tx, user)
	return args.Get(0).(models.User), args.Error(1)
}

func (m *MockRepositoryUser) UpdatePassword(ctx context.Context, tx *sql.Tx, id int, hashedPassword string) error {
	args := m.Called(ctx, tx, id, hashedPassword)
	return args.Error(0)
}

func (m *MockRepositoryUser) Delete(ctx context.Context, tx *sql.Tx, id int) error {
	args := m.Called(ctx, tx, id)
	return args.Error(0)
}

func (m *MockRepositoryUser) CountByRole(ctx context.Context, tx *sql.Tx, role string) (int, error) {
	args := m.Called(ctx, tx, role)
	return args.Int(0), args.Error(1)
}

func (m *MockRepositoryUser) FindById(ctx context.Context, tx *sql.Tx, id int) (models.User, error) {
	args := m.Called(ctx, tx, id)
	return args.Get(0).(models.User), args.Error(1)
}

func (m *MockRepositoryUser) FindByUsername(ctx context.Context, tx *sql.Tx, username string) (models.User, error) {
	args := m.Called(ctx, tx, username)
	return args.Get(0).(models.User), args.Error(1)
}

func (m *MockRepositoryUser) CreateLoginLog(ctx context.Context, tx *sql.Tx, userId int, ipAddress string) error {
	args := m.Called(ctx, tx, userId, ipAddress)
	return args.Error(0)
}

func (m *MockRepositoryUser) FindLastLoginByUserId(ctx context.Context, tx *sql.Tx, userId int) (models.UserLoginLog, error) {
	args := m.Called(ctx, tx, userId)
	return args.Get(0).(models.UserLoginLog), args.Error(1)
}

func (m *MockRepositoryUser) FindByRole(ctx context.Context, tx *sql.Tx, role string) ([]models.User, error) {
	args := m.Called(ctx, tx, role)
	return args.Get(0).([]models.User), args.Error(1)
}

func (m *MockRepositoryUser) CanCreateOnBehalfOf(ctx context.Context, tx *sql.Tx, actorUserId int, ownerUserId int) (bool, error) {
	args := m.Called(ctx, tx, actorUserId, ownerUserId)
	return args.Bool(0), args.Error(1)
}
