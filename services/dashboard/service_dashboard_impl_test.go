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

// TestGetLOIReport_PersonOrderPreserved verifies that multiple persons
// appear in the same order as the repository returns them.
// Retargeted from the acquisition report when that feed was retired on 2026-09-23:
// the behaviour under test is buildReport's ordering, which LOI still uses.
func TestGetLOIReport_PersonOrderPreserved(t *testing.T) {
	db, sqlMock := testutil.NewMockDB(t)
	repoDash := &mocks.MockRepositoryDashboard{}
	svc := newDashboardService(db, repoDash)

	table := models.LetterOfIntentTable

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

	report := svc.GetLOIReport(context.Background(), "", "", "") // pic, dateFrom, dateTo all empty → no filter

	assert.Len(t, report.ByPersonType, 3)
	assert.Equal(t, "Charlie", report.ByPersonType[0].Person)
	assert.Equal(t, "Alice", report.ByPersonType[1].Person)
	assert.Equal(t, "Bob", report.ByPersonType[2].Person)

	repoDash.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}
