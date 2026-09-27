package buildingproject

import (
	"context"
	"database/sql"

	"github.com/malikabdulaziz/tmn-backend/models"
)

// ListFilter is the shared filter for list and count, so the two cannot disagree
// about what they are counting.
type ListFilter struct {
	Search       string
	Status       string
	ContractType string
	Pic          string
}

type RepositoryBuildingProjectInterface interface {
	// includeFinance decides whether the landlord money columns are SELECTED at all.
	// It is a query-level gate, not a mapping-level one: a response mapper that drops
	// fields after loading them has still loaded them, and has still sent them
	// through every layer in between.
	FindAll(ctx context.Context, tx *sql.Tx, take int, skip int, orderBy string, orderDirection string, filter ListFilter, includeFinance bool) ([]models.BuildingProject, error)
	CountAll(ctx context.Context, tx *sql.Tx, filter ListFilter) (int, error)
	FindById(ctx context.Context, tx *sql.Tx, id int, includeFinance bool) (models.BuildingProject, error)
	FindByIris(ctx context.Context, tx *sql.Tx, iris string, includeFinance bool) (models.BuildingProject, error)

	Create(ctx context.Context, tx *sql.Tx, project models.BuildingProject) (models.BuildingProject, error)
	Update(ctx context.Context, tx *sql.Tx, project models.BuildingProject) (models.BuildingProject, error)
	Delete(ctx context.Context, tx *sql.Tx, id int) error

	// RecordChanges writes the audit rows. Always called inside the same transaction
	// as the write it describes, so a change cannot be logged for a write that rolled
	// back, nor a write land unlogged.
	RecordChanges(ctx context.Context, tx *sql.Tx, changes []models.BuildingProjectChange) error
	FindChanges(ctx context.Context, tx *sql.Tx, projectId int, take int, skip int) ([]models.BuildingProjectChange, error)
	CountChanges(ctx context.Context, tx *sql.Tx, projectId int) (int, error)
}
