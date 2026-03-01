package mocks

import (
	"context"
	"database/sql"

	repositoriesDashboard "github.com/malikabdulaziz/tmn-backend/repositories/dashboard"
	"github.com/stretchr/testify/mock"
)

// MockRepositoryDashboard implements repositories/dashboard.RepositoryDashboardInterface
type MockRepositoryDashboard struct {
	mock.Mock
}

func (m *MockRepositoryDashboard) GetStatusCounts(ctx context.Context, tx *sql.Tx, table, dedupField, pic, dateFrom, dateTo string) ([]repositoriesDashboard.StatusCount, error) {
	args := m.Called(ctx, tx, table, dedupField, pic, dateFrom, dateTo)
	return args.Get(0).([]repositoriesDashboard.StatusCount), args.Error(1)
}

func (m *MockRepositoryDashboard) GetByPersonAndType(ctx context.Context, tx *sql.Tx, table, dedupField, pic, dateFrom, dateTo string) ([]repositoriesDashboard.PersonTypeCount, error) {
	args := m.Called(ctx, tx, table, dedupField, pic, dateFrom, dateTo)
	return args.Get(0).([]repositoriesDashboard.PersonTypeCount), args.Error(1)
}

func (m *MockRepositoryDashboard) GetByPersonAndStatus(ctx context.Context, tx *sql.Tx, table, dedupField, pic, dateFrom, dateTo string) ([]repositoriesDashboard.PersonStatusCount, error) {
	args := m.Called(ctx, tx, table, dedupField, pic, dateFrom, dateTo)
	return args.Get(0).([]repositoriesDashboard.PersonStatusCount), args.Error(1)
}

func (m *MockRepositoryDashboard) GetDistinctPICs(ctx context.Context, tx *sql.Tx, table string) ([]string, error) {
	args := m.Called(ctx, tx, table)
	return args.Get(0).([]string), args.Error(1)
}
