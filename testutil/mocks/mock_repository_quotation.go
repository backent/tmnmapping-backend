package mocks

import (
	"context"
	"database/sql"

	"github.com/malikabdulaziz/tmn-backend/models"
	repositoriesQuotation "github.com/malikabdulaziz/tmn-backend/repositories/quotation"
	"github.com/stretchr/testify/mock"
)

// MockRepositoryQuotation implements repositories/quotation.RepositoryQuotationInterface
type MockRepositoryQuotation struct {
	mock.Mock
}

func (m *MockRepositoryQuotation) Create(ctx context.Context, tx *sql.Tx, q models.Quotation) (models.Quotation, error) {
	args := m.Called(ctx, tx, q)
	return args.Get(0).(models.Quotation), args.Error(1)
}

func (m *MockRepositoryQuotation) FindAll(ctx context.Context, tx *sql.Tx, take int, skip int, orderBy string, orderDirection string, scope repositoriesQuotation.ListScope) ([]models.Quotation, error) {
	args := m.Called(ctx, tx, take, skip, orderBy, orderDirection, scope)
	return args.Get(0).([]models.Quotation), args.Error(1)
}

func (m *MockRepositoryQuotation) CountAll(ctx context.Context, tx *sql.Tx, scope repositoriesQuotation.ListScope) (int, error) {
	args := m.Called(ctx, tx, scope)
	return args.Int(0), args.Error(1)
}

func (m *MockRepositoryQuotation) FindById(ctx context.Context, tx *sql.Tx, id int) (models.Quotation, error) {
	args := m.Called(ctx, tx, id)
	return args.Get(0).(models.Quotation), args.Error(1)
}

func (m *MockRepositoryQuotation) UpdateDraft(ctx context.Context, tx *sql.Tx, q models.Quotation) error {
	args := m.Called(ctx, tx, q)
	return args.Error(0)
}

func (m *MockRepositoryQuotation) UpdatePricingAndStatus(ctx context.Context, tx *sql.Tx, q models.Quotation) error {
	args := m.Called(ctx, tx, q)
	return args.Error(0)
}

func (m *MockRepositoryQuotation) Delete(ctx context.Context, tx *sql.Tx, id int) error {
	args := m.Called(ctx, tx, id)
	return args.Error(0)
}

func (m *MockRepositoryQuotation) ReplaceSelections(ctx context.Context, tx *sql.Tx, quotationId int, selections []models.QuotationSelection) error {
	args := m.Called(ctx, tx, quotationId, selections)
	return args.Error(0)
}

func (m *MockRepositoryQuotation) FindSelections(ctx context.Context, tx *sql.Tx, quotationId int) ([]models.QuotationSelection, error) {
	args := m.Called(ctx, tx, quotationId)
	return args.Get(0).([]models.QuotationSelection), args.Error(1)
}

func (m *MockRepositoryQuotation) CreateVersionSnapshot(ctx context.Context, tx *sql.Tx, quotationId int, version int, snapshot []byte, approverUserId int) error {
	args := m.Called(ctx, tx, quotationId, version, snapshot, approverUserId)
	return args.Error(0)
}

func (m *MockRepositoryQuotation) CreateApproval(ctx context.Context, tx *sql.Tx, a models.QuotationApproval) error {
	args := m.Called(ctx, tx, a)
	return args.Error(0)
}

func (m *MockRepositoryQuotation) FindApprovals(ctx context.Context, tx *sql.Tx, quotationId int) ([]models.QuotationApproval, error) {
	args := m.Called(ctx, tx, quotationId)
	return args.Get(0).([]models.QuotationApproval), args.Error(1)
}

func (m *MockRepositoryQuotation) NextQuoteNumber(ctx context.Context, tx *sql.Tx, year int) (string, error) {
	args := m.Called(ctx, tx, year)
	return args.String(0), args.Error(1)
}

func (m *MockRepositoryQuotation) CountByStatusForOwner(ctx context.Context, tx *sql.Tx, ownerUserId int) (map[string]int, error) {
	args := m.Called(ctx, tx, ownerUserId)
	return args.Get(0).(map[string]int), args.Error(1)
}
