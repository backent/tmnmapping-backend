package dashboard_test

import (
	"context"
	"database/sql"
	"io"
	"testing"

	"github.com/malikabdulaziz/tmn-backend/models"
	repositoriesDashboard "github.com/malikabdulaziz/tmn-backend/repositories/dashboard"
	serviceDashboard "github.com/malikabdulaziz/tmn-backend/services/dashboard"
	"github.com/malikabdulaziz/tmn-backend/testutil"
	"github.com/malikabdulaziz/tmn-backend/testutil/mocks"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newDashboardService(db *sql.DB, repoDash *mocks.MockRepositoryDashboard) serviceDashboard.ServiceDashboardInterface {
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	return serviceDashboard.NewServiceDashboardImpl(db, repoDash, logger)
}

// setupDashboardMocks registers the four repository calls that buildReport makes.
func setupDashboardMocks(
	repoDash *mocks.MockRepositoryDashboard,
	table string,
	statusCounts []repositoriesDashboard.StatusCount,
	personTypeRows []repositoriesDashboard.PersonTypeCount,
	personStatusRows []repositoriesDashboard.PersonStatusCount,
	pics []string,
) {
	repoDash.On("GetStatusCounts", mock.Anything, mock.AnythingOfType("*sql.Tx"),
		table, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
	).Return(statusCounts, nil)

	repoDash.On("GetByPersonAndType", mock.Anything, mock.AnythingOfType("*sql.Tx"),
		table, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
	).Return(personTypeRows, nil)

	repoDash.On("GetByPersonAndStatus", mock.Anything, mock.AnythingOfType("*sql.Tx"),
		table, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
	).Return(personStatusRows, nil)

	repoDash.On("GetDistinctPICs", mock.Anything, mock.AnythingOfType("*sql.Tx"), table).
		Return(pics, nil)
}

// --- GetAcquisitionReport ---

func TestGetAcquisitionReport_HappyPath(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repoDash := &mocks.MockRepositoryDashboard{}
	svc := newDashboardService(db, repoDash)

	table := models.AcquisitionTable

	statusCounts := []repositoriesDashboard.StatusCount{
		{WorkflowState: "Open", Count: 10},
		{WorkflowState: "Closed", Count: 5},
	}
	personTypeRows := []repositoriesDashboard.PersonTypeCount{
		{Person: "Alice", BuildingType: "Office", Count: 3},
		{Person: "Alice", BuildingType: "Mall", Count: 2},
		{Person: "Bob", BuildingType: "Office", Count: 7},
	}
	personStatusRows := []repositoriesDashboard.PersonStatusCount{
		{Person: "Alice", WorkflowState: "Open", Count: 4},
		{Person: "Bob", WorkflowState: "Closed", Count: 3},
	}
	pics := []string{"Alice", "Bob"}

	sqlMock.ExpectBegin()
	sqlMock.ExpectCommit()

	setupDashboardMocks(repoDash, table, statusCounts, personTypeRows, personStatusRows, pics)

	report := svc.GetAcquisitionReport(context.Background(), "", "", "") // pic, dateFrom, dateTo all empty → no filter

	// Stats aggregation
	assert.Equal(t, 15, report.Stats.Total)
	assert.Equal(t, 10, report.Stats.ByStatus["Open"])
	assert.Equal(t, 5, report.Stats.ByStatus["Closed"])

	// Person x type — insertion order preserved
	assert.Len(t, report.ByPersonType, 2)
	assert.Equal(t, "Alice", report.ByPersonType[0].Person)
	assert.Equal(t, 3, report.ByPersonType[0].ByType["Office"])
	assert.Equal(t, 2, report.ByPersonType[0].ByType["Mall"])
	assert.Equal(t, "Bob", report.ByPersonType[1].Person)
	assert.Equal(t, 7, report.ByPersonType[1].ByType["Office"])

	// Person x status — insertion order preserved
	assert.Len(t, report.ByPersonStatus, 2)
	assert.Equal(t, "Alice", report.ByPersonStatus[0].Person)
	assert.Equal(t, 4, report.ByPersonStatus[0].ByStatus["Open"])
	assert.Equal(t, "Bob", report.ByPersonStatus[1].Person)
	assert.Equal(t, 3, report.ByPersonStatus[1].ByStatus["Closed"])

	// PICs
	assert.Equal(t, []string{"Alice", "Bob"}, report.PICs)

	repoDash.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

// TestGetAcquisitionReport_EmptyResults verifies that nil PICs from the repo
// are normalized to an empty slice (not nil) in the response.
func TestGetAcquisitionReport_NilPICs(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repoDash := &mocks.MockRepositoryDashboard{}
	svc := newDashboardService(db, repoDash)

	table := models.AcquisitionTable

	sqlMock.ExpectBegin()
	sqlMock.ExpectCommit()

	setupDashboardMocks(repoDash, table,
		[]repositoriesDashboard.StatusCount{},
		[]repositoriesDashboard.PersonTypeCount{},
		[]repositoriesDashboard.PersonStatusCount{},
		nil, // repo returns nil
	)

	report := svc.GetAcquisitionReport(context.Background(), "", "", "") // pic, dateFrom, dateTo all empty → no filter

	// Service normalizes nil PICs to empty slice
	assert.NotNil(t, report.PICs)
	assert.Len(t, report.PICs, 0)
	assert.Equal(t, 0, report.Stats.Total)
	assert.Len(t, report.ByPersonType, 0)
	assert.Len(t, report.ByPersonStatus, 0)

	repoDash.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

// --- GetBuildingProposalReport ---

func TestGetBuildingProposalReport_HappyPath(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repoDash := &mocks.MockRepositoryDashboard{}
	svc := newDashboardService(db, repoDash)

	table := models.BuildingProposalTable

	sqlMock.ExpectBegin()
	sqlMock.ExpectCommit()

	setupDashboardMocks(repoDash, table,
		[]repositoriesDashboard.StatusCount{{WorkflowState: "Approved", Count: 8}},
		[]repositoriesDashboard.PersonTypeCount{},
		[]repositoriesDashboard.PersonStatusCount{},
		[]string{"Carol"},
	)

	report := svc.GetBuildingProposalReport(context.Background(), "", "", "") // pic, dateFrom, dateTo all empty → no filter

	assert.Equal(t, 8, report.Stats.Total)
	assert.Equal(t, 8, report.Stats.ByStatus["Approved"])
	assert.Equal(t, []string{"Carol"}, report.PICs)

	repoDash.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

// --- GetLOIReport ---

func TestGetLOIReport_HappyPath(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repoDash := &mocks.MockRepositoryDashboard{}
	svc := newDashboardService(db, repoDash)

	table := models.LetterOfIntentTable

	sqlMock.ExpectBegin()
	sqlMock.ExpectCommit()

	setupDashboardMocks(repoDash, table,
		[]repositoriesDashboard.StatusCount{
			{WorkflowState: "Signed", Count: 4},
			{WorkflowState: "Pending", Count: 6},
		},
		[]repositoriesDashboard.PersonTypeCount{},
		[]repositoriesDashboard.PersonStatusCount{},
		[]string{"Dave"},
	)

	report := svc.GetLOIReport(context.Background(), "", "", "") // pic, dateFrom, dateTo all empty → no filter

	assert.Equal(t, 10, report.Stats.Total)
	assert.Equal(t, 4, report.Stats.ByStatus["Signed"])
	assert.Equal(t, 6, report.Stats.ByStatus["Pending"])

	repoDash.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

// TestGetAcquisitionReport_PersonOrderPreserved verifies that multiple persons
// appear in the same order as the repository returns them.
func TestGetAcquisitionReport_PersonOrderPreserved(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repoDash := &mocks.MockRepositoryDashboard{}
	svc := newDashboardService(db, repoDash)

	table := models.AcquisitionTable

	// Repo returns persons in Charlie → Alice → Bob order
	personTypeRows := []repositoriesDashboard.PersonTypeCount{
		{Person: "Charlie", BuildingType: "Mall", Count: 1},
		{Person: "Alice", BuildingType: "Office", Count: 2},
		{Person: "Bob", BuildingType: "Mall", Count: 3},
	}

	sqlMock.ExpectBegin()
	sqlMock.ExpectCommit()

	setupDashboardMocks(repoDash, table,
		[]repositoriesDashboard.StatusCount{},
		personTypeRows,
		[]repositoriesDashboard.PersonStatusCount{},
		[]string{},
	)

	report := svc.GetAcquisitionReport(context.Background(), "", "", "") // pic, dateFrom, dateTo all empty → no filter

	assert.Len(t, report.ByPersonType, 3)
	assert.Equal(t, "Charlie", report.ByPersonType[0].Person)
	assert.Equal(t, "Alice", report.ByPersonType[1].Person)
	assert.Equal(t, "Bob", report.ByPersonType[2].Person)

	repoDash.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}
