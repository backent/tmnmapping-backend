package buildingrestriction

import (
	"context"
	"database/sql"

	"github.com/malikabdulaziz/tmn-backend/models"
)

type RepositoryBuildingRestrictionInterface interface {
	Create(ctx context.Context, tx *sql.Tx, restriction models.BuildingRestriction, buildingIds []int) (models.BuildingRestriction, error)
	FindAll(ctx context.Context, tx *sql.Tx, take int, skip int, orderBy string, orderDirection string) ([]models.BuildingRestriction, error)
	CountAll(ctx context.Context, tx *sql.Tx) (int, error)
	FindById(ctx context.Context, tx *sql.Tx, id int) (models.BuildingRestriction, error)
	Update(ctx context.Context, tx *sql.Tx, restriction models.BuildingRestriction, buildingIds []int) (models.BuildingRestriction, error)
	DeleteBuildingLinksByBuildingRestrictionId(ctx context.Context, tx *sql.Tx, buildingRestrictionId int) error
	CreateBuildingLink(ctx context.Context, tx *sql.Tx, buildingRestrictionId int, buildingId int) error
	Delete(ctx context.Context, tx *sql.Tx, id int) error
}
