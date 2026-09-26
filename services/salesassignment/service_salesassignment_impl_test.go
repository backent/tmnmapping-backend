package salesassignment_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/malikabdulaziz/tmn-backend/models"
	service "github.com/malikabdulaziz/tmn-backend/services/salesassignment"
	"github.com/malikabdulaziz/tmn-backend/spreadsheets"
	"github.com/malikabdulaziz/tmn-backend/testutil"
	"github.com/malikabdulaziz/tmn-backend/testutil/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type deps struct {
	assignment *mocks.MockRepositorySalesAssignment
	customer   *mocks.MockRepositoryCustomer
	brand      *mocks.MockRepositoryBrand
	user       *mocks.MockRepositoryUser
}

func newService(t *testing.T) (service.ServiceSalesAssignmentInterface, deps, func()) {
	t.Helper()

	db, sqlMock := testutil.NewMockDB(t)
	d := deps{
		assignment: &mocks.MockRepositorySalesAssignment{},
		customer:   &mocks.MockRepositoryCustomer{},
		brand:      &mocks.MockRepositoryBrand{},
		user:       &mocks.MockRepositoryUser{},
	}

	sqlMock.ExpectBegin()
	sqlMock.ExpectCommit()

	svc := service.NewServiceSalesAssignmentImpl(db, d.assignment, d.customer, d.brand, d.user)

	return svc, d, func() { assert.NoError(t, sqlMock.ExpectationsWereMet()) }
}

func csv(lines ...string) []byte {
	out := ""
	for _, line := range lines {
		out += line + "\n"
	}

	return []byte(out)
}

// stubResolves makes CUST-001 / BRND-001 / sales1 all resolve consistently.
func (d deps) stubResolves() {
	d.customer.On("FindByCode", mock.Anything, mock.Anything, "CUST-001").
		Return(models.Customer{Id: 1, Code: "CUST-001"}, nil)
	d.brand.On("FindByCode", mock.Anything, mock.Anything, "BRND-001").
		Return(models.Brand{Id: 10, Code: "BRND-001", CustomerId: 1, CustomerCode: "CUST-001"}, nil)
	d.user.On("FindByUsername", mock.Anything, mock.Anything, "sales1").
		Return(testutil.NewUser(5, "sales1", "secret123", models.RoleSales), nil)
}

func TestTemplate_RoundTripsThroughTheImporter(t *testing.T) {
	svc, _, _ := newService(t)

	fileBytes, err := svc.Template()
	require.NoError(t, err)

	rows, err := spreadsheets.ParseSpreadsheet(fileBytes, "xlsx")
	require.NoError(t, err)

	colMap := spreadsheets.MapHeaderColumns(rows[0], service.TemplateColumns)
	assert.Empty(t, spreadsheets.MissingRequiredColumns(colMap, service.TemplateColumns))
}

func TestImport_CreatesAssignment(t *testing.T) {
	svc, d, assertMock := newService(t)
	d.stubResolves()

	d.assignment.On("FindByCustomerAndBrand", mock.Anything, mock.Anything, 1, 10).
		Return(models.SalesAssignment{}, sql.ErrNoRows)

	var captured models.SalesAssignment
	d.assignment.On("Create", mock.Anything, mock.Anything, mock.AnythingOfType("models.SalesAssignment")).
		Run(func(args mock.Arguments) { captured = args.Get(2).(models.SalesAssignment) }).
		Return(models.SalesAssignment{Id: 1}, nil)

	result := svc.Import(context.Background(), csv(
		"Customer Code,Brand Code,Sales Username,Status,Registration Date,Expiry Date",
		"CUST-001,BRND-001,sales1,active,2026-01-15,2026-12-31",
	), "csv")

	assert.True(t, result.Imported)
	assert.Equal(t, 1, result.Created)
	assert.Equal(t, 5, captured.SalesUserId)
	assert.Equal(t, "2026-01-15", captured.RegistrationDate)
	assertMock()
}

// Re-uploading a pair reassigns it: the table allows one PIC per customer+brand.
func TestImport_ReassignsExistingPair(t *testing.T) {
	svc, d, assertMock := newService(t)
	d.stubResolves()

	d.assignment.On("FindByCustomerAndBrand", mock.Anything, mock.Anything, 1, 10).
		Return(models.SalesAssignment{Id: 42}, nil)
	d.assignment.On("Update", mock.Anything, mock.Anything, mock.AnythingOfType("models.SalesAssignment")).
		Return(models.SalesAssignment{Id: 42}, nil)

	result := svc.Import(context.Background(), csv(
		"Customer Code,Brand Code,Sales Username",
		"CUST-001,BRND-001,sales1",
	), "csv")

	assert.True(t, result.Imported)
	assert.Equal(t, 1, result.Updated)
	assert.Equal(t, 0, result.Created)
	assertMock()
}

// A brand under a different customer is the mistake most likely to slip past a
// human reading the spreadsheet.
func TestImport_RejectsBrandBelongingToAnotherCustomer(t *testing.T) {
	svc, d, assertMock := newService(t)

	d.customer.On("FindByCode", mock.Anything, mock.Anything, "CUST-001").
		Return(models.Customer{Id: 1, Code: "CUST-001"}, nil)
	d.brand.On("FindByCode", mock.Anything, mock.Anything, "BRND-999").
		Return(models.Brand{Id: 99, Code: "BRND-999", CustomerId: 2, CustomerCode: "CUST-002"}, nil)

	result := svc.Import(context.Background(), csv(
		"Customer Code,Brand Code,Sales Username",
		"CUST-001,BRND-999,sales1",
	), "csv")

	assert.False(t, result.Imported)
	require.Len(t, result.Errors, 1)
	assert.Contains(t, result.Errors[0].Message, "belongs to CUST-002")
	d.assignment.AssertNotCalled(t, "Create", mock.Anything, mock.Anything, mock.Anything)
	assertMock()
}

func TestImport_RejectsUnknownUsername(t *testing.T) {
	svc, d, assertMock := newService(t)

	d.customer.On("FindByCode", mock.Anything, mock.Anything, "CUST-001").
		Return(models.Customer{Id: 1, Code: "CUST-001"}, nil)
	d.brand.On("FindByCode", mock.Anything, mock.Anything, "BRND-001").
		Return(models.Brand{Id: 10, CustomerId: 1}, nil)
	d.user.On("FindByUsername", mock.Anything, mock.Anything, "ghost").
		Return(models.User{}, sql.ErrNoRows)

	result := svc.Import(context.Background(), csv(
		"Customer Code,Brand Code,Sales Username",
		"CUST-001,BRND-001,ghost",
	), "csv")

	assert.False(t, result.Imported)
	require.Len(t, result.Errors, 1)
	assert.Contains(t, result.Errors[0].Message, "No user with this username")
	assertMock()
}

// Excel writes a date cell as a timestamp, not as YYYY-MM-DD.
func TestImport_AcceptsExcelStyleDates(t *testing.T) {
	svc, d, assertMock := newService(t)
	d.stubResolves()

	d.assignment.On("FindByCustomerAndBrand", mock.Anything, mock.Anything, 1, 10).
		Return(models.SalesAssignment{}, sql.ErrNoRows)

	var captured models.SalesAssignment
	d.assignment.On("Create", mock.Anything, mock.Anything, mock.AnythingOfType("models.SalesAssignment")).
		Run(func(args mock.Arguments) { captured = args.Get(2).(models.SalesAssignment) }).
		Return(models.SalesAssignment{Id: 1}, nil)

	result := svc.Import(context.Background(), csv(
		"Customer Code,Brand Code,Sales Username,Registration Date",
		"CUST-001,BRND-001,sales1,2026-01-15 00:00:00",
	), "csv")

	assert.True(t, result.Imported)
	assert.Equal(t, "2026-01-15", captured.RegistrationDate)
	assertMock()
}

func TestImport_RejectsExpiryBeforeRegistration(t *testing.T) {
	svc, d, assertMock := newService(t)
	d.stubResolves()

	result := svc.Import(context.Background(), csv(
		"Customer Code,Brand Code,Sales Username,Registration Date,Expiry Date",
		"CUST-001,BRND-001,sales1,2026-12-31,2026-01-01",
	), "csv")

	assert.False(t, result.Imported)
	require.Len(t, result.Errors, 1)
	assert.Contains(t, result.Errors[0].Message, "cannot be earlier than")
	assertMock()
}

// One PIC per brand, so the same pair twice in one file is a mistake, not an update.
func TestImport_RejectsDuplicatePairWithinOneFile(t *testing.T) {
	svc, d, assertMock := newService(t)
	d.stubResolves()

	d.assignment.On("FindByCustomerAndBrand", mock.Anything, mock.Anything, 1, 10).
		Return(models.SalesAssignment{}, sql.ErrNoRows)

	result := svc.Import(context.Background(), csv(
		"Customer Code,Brand Code,Sales Username",
		"CUST-001,BRND-001,sales1",
		"CUST-001,BRND-001,sales1",
	), "csv")

	assert.False(t, result.Imported)
	require.Len(t, result.Errors, 1)
	assert.Contains(t, result.Errors[0].Message, "only one sales PIC")
	assertMock()
}
