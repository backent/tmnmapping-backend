package building

import (
	"context"
	"database/sql"

	"github.com/malikabdulaziz/tmn-backend/models"
)

type RepositoryBuildingInterface interface {
	Create(ctx context.Context, tx *sql.Tx, building models.Building) (models.Building, error)
	FindById(ctx context.Context, tx *sql.Tx, id int) (models.Building, error)
	FindByExternalId(ctx context.Context, tx *sql.Tx, externalId string) (models.Building, error)
	FindAll(ctx context.Context, tx *sql.Tx, take int, skip int, orderBy string, orderDirection string, search string, buildingStatus string, sellable string, connectivity string, resourceType string, competitorLocation *bool, cbdArea string, subdistrict string, citytown string, province string, gradeResource string, buildingType string) ([]models.Building, error)
	CountAll(ctx context.Context, tx *sql.Tx, search string, buildingStatus string, sellable string, connectivity string, resourceType string, competitorLocation *bool, cbdArea string, subdistrict string, citytown string, province string, gradeResource string, buildingType string) (int, error)
	GetDistinctValues(ctx context.Context, tx *sql.Tx, columnName string) ([]string, error)
	Update(ctx context.Context, tx *sql.Tx, building models.Building) (models.Building, error)
	UpdateFromSync(ctx context.Context, tx *sql.Tx, building models.Building) (models.Building, error)
	FindAllForMapping(ctx context.Context, tx *sql.Tx, buildingType string, buildingGrade string, year string, subdistrict string, progress string, sellable string, connectivity string, lcdPresence string, lat *float64, lng *float64, radius *int) ([]models.Building, error)
}

