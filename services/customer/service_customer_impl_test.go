package customer_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/malikabdulaziz/tmn-backend/exceptions"
	"github.com/malikabdulaziz/tmn-backend/models"
	serviceCustomer "github.com/malikabdulaziz/tmn-backend/services/customer"
	"github.com/malikabdulaziz/tmn-backend/spreadsheets"
	"github.com/malikabdulaziz/tmn-backend/testutil"
	"github.com/malikabdulaziz/tmn-backend/testutil/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newService(t *testing.T, commits bool) (serviceCustomer.ServiceCustomerInterface, *mocks.MockRepositoryCustomer, func()) {
	t.Helper()

	db, sqlMock := testutil.NewMockDB(t)
	repo := &mocks.MockRepositoryCustomer{}

	sqlMock.ExpectBegin()
	if commits {
		sqlMock.ExpectCommit()
	} else {
		sqlMock.ExpectRollback()
	}

	return serviceCustomer.NewServiceCustomerImpl(db, repo), repo, func() {
		assert.NoError(t, sqlMock.ExpectationsWereMet())
	}
}

// csv builds an upload body. CSV and XLSX share the same parsing path once rows are
// extracted, so the import tests use CSV for readability.
func csv(lines ...string) []byte {
	out := ""
	for _, line := range lines {
		out += line + "\n"
	}

	return []byte(out)
}

// ---------------------------------------------------------------------------
// Template
// ---------------------------------------------------------------------------

// The template and the importer read the same column list, so a template that omits
// a required column would mean an upload of that template always fails.
func TestTemplate_RoundTripsThroughTheImporter(t *testing.T) {
	svc, _, _ := newService(t, true)

	fileBytes, err := svc.Template()
	require.NoError(t, err)
	assert.NotEmpty(t, fileBytes)

	rows, err := spreadsheets.ParseSpreadsheet(fileBytes, "xlsx")
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(rows), 2, "template needs a header and an example row")

	colMap := spreadsheets.MapHeaderColumns(rows[0], serviceCustomer.TemplateColumns)
	assert.Empty(t, spreadsheets.MissingRequiredColumns(colMap, serviceCustomer.TemplateColumns),
		"the template must satisfy its own importer")
}

// ---------------------------------------------------------------------------
// Import
// ---------------------------------------------------------------------------

func TestImport_CreatesAndUpdatesByCode(t *testing.T) {
	svc, repo, assertMock := newService(t, true)

	repo.On("FindByCode", mock.Anything, mock.Anything, "CUST-001").
		Return(models.Customer{}, sql.ErrNoRows)
	repo.On("FindByCode", mock.Anything, mock.Anything, "CUST-002").
		Return(models.Customer{Id: 7, Code: "CUST-002"}, nil)
	repo.On("Create", mock.Anything, mock.Anything, mock.AnythingOfType("models.Customer")).
		Return(models.Customer{Id: 1}, nil)
	repo.On("Update", mock.Anything, mock.Anything, mock.AnythingOfType("models.Customer")).
		Return(models.Customer{Id: 7}, nil)

	result := svc.Import(context.Background(), csv(
		"Customer Code,Customer Name,Industry,Status",
		"CUST-001,New Advertiser,Retail,active",
		"CUST-002,Existing Advertiser,Banking,inactive",
	), "csv")

	assert.True(t, result.Imported)
	assert.Equal(t, 2, result.Rows)
	assert.Equal(t, 1, result.Created)
	assert.Equal(t, 1, result.Updated)
	assert.Empty(t, result.Errors)
	assertMock()
}

// Column order is the operator's business, not ours.
func TestImport_ResolvesColumnsByHeaderNotPosition(t *testing.T) {
	svc, repo, assertMock := newService(t, true)

	var captured models.Customer
	repo.On("FindByCode", mock.Anything, mock.Anything, "CUST-009").
		Return(models.Customer{}, sql.ErrNoRows)
	repo.On("Create", mock.Anything, mock.Anything, mock.AnythingOfType("models.Customer")).
		Run(func(args mock.Arguments) { captured = args.Get(2).(models.Customer) }).
		Return(models.Customer{Id: 1}, nil)

	result := svc.Import(context.Background(), csv(
		"Status,Industry,Customer Name,Customer Code",
		"active,Telco,Reordered Co,CUST-009",
	), "csv")

	assert.True(t, result.Imported)
	assert.Equal(t, "CUST-009", captured.Code)
	assert.Equal(t, "Reordered Co", captured.Name)
	assert.Equal(t, "Telco", captured.Industry)
	assertMock()
}

func TestImport_StatusDefaultsToActive(t *testing.T) {
	svc, repo, assertMock := newService(t, true)

	var captured models.Customer
	repo.On("FindByCode", mock.Anything, mock.Anything, "CUST-003").
		Return(models.Customer{}, sql.ErrNoRows)
	repo.On("Create", mock.Anything, mock.Anything, mock.AnythingOfType("models.Customer")).
		Run(func(args mock.Arguments) { captured = args.Get(2).(models.Customer) }).
		Return(models.Customer{Id: 2}, nil)

	result := svc.Import(context.Background(), csv(
		"Customer Code,Customer Name,Industry,Status",
		"CUST-003,No Status Given,,",
	), "csv")

	assert.True(t, result.Imported)
	assert.Equal(t, models.StatusActive, captured.Status)
	assertMock()
}

// The whole point of validating first: a bad row must leave the database untouched.
func TestImport_WritesNothingWhenAnyRowFails(t *testing.T) {
	svc, repo, assertMock := newService(t, true)

	repo.On("FindByCode", mock.Anything, mock.Anything, "CUST-001").
		Return(models.Customer{}, sql.ErrNoRows)

	result := svc.Import(context.Background(), csv(
		"Customer Code,Customer Name,Industry,Status",
		"CUST-001,Good Row,Retail,active",
		",Missing Code,Retail,active",
		"CUST-004,Bad Status,Retail,pending",
	), "csv")

	assert.False(t, result.Imported)
	assert.Equal(t, 0, result.Created)
	assert.Equal(t, 0, result.Updated)
	assert.Len(t, result.Errors, 2, "every problem is reported in one pass")

	repo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything, mock.Anything)
	repo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything, mock.Anything)
	assertMock()
}

// Row numbers must match what the operator sees in Excel: row 1 is the header.
func TestImport_ReportsSpreadsheetRowNumbers(t *testing.T) {
	svc, repo, assertMock := newService(t, true)

	repo.On("FindByCode", mock.Anything, mock.Anything, "CUST-001").
		Return(models.Customer{}, sql.ErrNoRows)

	result := svc.Import(context.Background(), csv(
		"Customer Code,Customer Name",
		"CUST-001,Fine",
		",Broken",
	), "csv")

	require.Len(t, result.Errors, 1)
	assert.Equal(t, 3, result.Errors[0].Row)
	assert.Equal(t, "Customer Code", result.Errors[0].Column)
	assertMock()
}

func TestImport_RejectsDuplicateCodeWithinOneFile(t *testing.T) {
	svc, repo, assertMock := newService(t, true)

	repo.On("FindByCode", mock.Anything, mock.Anything, "CUST-001").
		Return(models.Customer{}, sql.ErrNoRows)

	result := svc.Import(context.Background(), csv(
		"Customer Code,Customer Name",
		"CUST-001,First",
		"CUST-001,Second",
	), "csv")

	assert.False(t, result.Imported)
	require.Len(t, result.Errors, 1)
	assert.Contains(t, result.Errors[0].Message, "row 2")
	assertMock()
}

// Excel leaves formatted-but-empty rows below real data all the time.
func TestImport_SkipsBlankRows(t *testing.T) {
	svc, repo, assertMock := newService(t, true)

	repo.On("FindByCode", mock.Anything, mock.Anything, "CUST-001").
		Return(models.Customer{}, sql.ErrNoRows)
	repo.On("Create", mock.Anything, mock.Anything, mock.AnythingOfType("models.Customer")).
		Return(models.Customer{Id: 1}, nil)

	result := svc.Import(context.Background(), csv(
		"Customer Code,Customer Name,Industry,Status",
		"CUST-001,Only Real Row,Retail,active",
		",,,",
		",,,",
	), "csv")

	assert.True(t, result.Imported)
	assert.Equal(t, 1, result.Rows)
	assert.Equal(t, 1, result.Created)
	assertMock()
}

// A file missing a required column is rejected before any transaction is opened,
// so a malformed upload never touches the database at all.
func TestImport_RejectsFileMissingARequiredColumn(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repo := &mocks.MockRepositoryCustomer{}
	svc := serviceCustomer.NewServiceCustomerImpl(db, repo)

	assert.PanicsWithValue(t,
		exceptions.NewBadRequestError("Missing required column(s): Customer Name"),
		func() {
			svc.Import(context.Background(), csv(
				"Customer Code,Industry",
				"CUST-001,Retail",
			), "csv")
		})

	// No ExpectBegin was registered, and none was needed.
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

// ---------------------------------------------------------------------------
// Delete
// ---------------------------------------------------------------------------

// The FK is ON DELETE RESTRICT, so this must be a readable 400, not a 500.
func TestDelete_RefusesWhileCustomerHasBrands(t *testing.T) {
	svc, repo, assertMock := newService(t, false)

	repo.On("FindById", mock.Anything, mock.Anything, 5).
		Return(models.Customer{Id: 5, Code: "CUST-005"}, nil)
	repo.On("CountBrands", mock.Anything, mock.Anything, 5).Return(2, nil)

	assert.PanicsWithValue(t,
		exceptions.NewBadRequestError("this customer still has brands; delete or reassign them first"),
		func() { svc.Delete(context.Background(), 5) })

	repo.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything, mock.Anything)
	assertMock()
}

func TestDelete_HappyPath(t *testing.T) {
	svc, repo, assertMock := newService(t, true)

	repo.On("FindById", mock.Anything, mock.Anything, 5).Return(models.Customer{Id: 5}, nil)
	repo.On("CountBrands", mock.Anything, mock.Anything, 5).Return(0, nil)
	repo.On("Delete", mock.Anything, mock.Anything, 5).Return(nil)

	svc.Delete(context.Background(), 5)

	repo.AssertExpectations(t)
	assertMock()
}
