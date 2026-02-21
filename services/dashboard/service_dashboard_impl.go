package dashboard

import (
	"context"
	"database/sql"
	"os"

	"github.com/malikabdulaziz/tmn-backend/helpers"
	"github.com/malikabdulaziz/tmn-backend/models"
	repositoriesDashboard "github.com/malikabdulaziz/tmn-backend/repositories/dashboard"
	webDashboard "github.com/malikabdulaziz/tmn-backend/web/dashboard"
	"github.com/sirupsen/logrus"
)

type ServiceDashboardImpl struct {
	DB         *sql.DB
	Repository repositoriesDashboard.RepositoryDashboardInterface
	Logger     *logrus.Logger
}

func NewServiceDashboardImpl(
	db *sql.DB,
	repository repositoriesDashboard.RepositoryDashboardInterface,
	logger *logrus.Logger,
) ServiceDashboardInterface {
	return &ServiceDashboardImpl{
		DB:         db,
		Repository: repository,
		Logger:     logger,
	}
}

// dedupKeyFor returns the configured dedup field for the given resource name,
// falling back to "building_project" if not set.
func dedupKeyFor(envKey string) string {
	if v := os.Getenv(envKey); v != "" {
		return v
	}
	return "building_project"
}

func (s *ServiceDashboardImpl) GetAcquisitionReport(ctx context.Context, pic, month string) webDashboard.DashboardReport {
	dedupField := dedupKeyFor("DASHBOARD_DEDUP_KEY_ACQUISITION")
	return s.buildReport(ctx, models.AcquisitionTable, dedupField, pic, month)
}

func (s *ServiceDashboardImpl) GetBuildingProposalReport(ctx context.Context, pic, month string) webDashboard.DashboardReport {
	dedupField := dedupKeyFor("DASHBOARD_DEDUP_KEY_PROPOSAL")
	return s.buildReport(ctx, models.BuildingProposalTable, dedupField, pic, month)
}

func (s *ServiceDashboardImpl) GetLOIReport(ctx context.Context, pic, month string) webDashboard.DashboardReport {
	dedupField := dedupKeyFor("DASHBOARD_DEDUP_KEY_LOI")
	return s.buildReport(ctx, models.LetterOfIntentTable, dedupField, pic, month)
}

// buildReport runs all 3 queries for the given table and assembles the response.
func (s *ServiceDashboardImpl) buildReport(ctx context.Context, table, dedupField, pic, month string) webDashboard.DashboardReport {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	// 1. Status counts
	statusCounts, err := s.Repository.GetStatusCounts(ctx, tx, table, dedupField, pic, month)
	helpers.PanicIfError(err)

	stats := webDashboard.StatsSummary{
		ByStatus: make(map[string]int),
	}
	for _, sc := range statusCounts {
		stats.ByStatus[sc.WorkflowState] = sc.Count
		stats.Total += sc.Count
	}

	// 2. By person x building type
	personTypeRows, err := s.Repository.GetByPersonAndType(ctx, tx, table, dedupField, pic, month)
	helpers.PanicIfError(err)

	personTypeMap := make(map[string]map[string]int)
	personOrder := []string{}
	for _, row := range personTypeRows {
		if _, exists := personTypeMap[row.Person]; !exists {
			personTypeMap[row.Person] = make(map[string]int)
			personOrder = append(personOrder, row.Person)
		}
		personTypeMap[row.Person][row.BuildingType] = row.Count
	}
	byPersonType := make([]webDashboard.PersonTypeStat, 0, len(personTypeMap))
	for _, person := range personOrder {
		byPersonType = append(byPersonType, webDashboard.PersonTypeStat{
			Person: person,
			ByType: personTypeMap[person],
		})
	}

	// 3. By person x status
	personStatusRows, err := s.Repository.GetByPersonAndStatus(ctx, tx, table, dedupField, pic, month)
	helpers.PanicIfError(err)

	personStatusMap := make(map[string]map[string]int)
	personStatusOrder := []string{}
	for _, row := range personStatusRows {
		if _, exists := personStatusMap[row.Person]; !exists {
			personStatusMap[row.Person] = make(map[string]int)
			personStatusOrder = append(personStatusOrder, row.Person)
		}
		personStatusMap[row.Person][row.WorkflowState] = row.Count
	}
	byPersonStatus := make([]webDashboard.PersonStatusStat, 0, len(personStatusMap))
	for _, person := range personStatusOrder {
		byPersonStatus = append(byPersonStatus, webDashboard.PersonStatusStat{
			Person:   person,
			ByStatus: personStatusMap[person],
		})
	}

	// 4. Distinct PICs for filter dropdown
	pics, err := s.Repository.GetDistinctPICs(ctx, tx, table)
	helpers.PanicIfError(err)
	if pics == nil {
		pics = []string{}
	}

	return webDashboard.DashboardReport{
		Stats:          stats,
		ByPersonType:   byPersonType,
		ByPersonStatus: byPersonStatus,
		PICs:           pics,
	}
}
