package building

import (
	"context"
	"database/sql"
	"time"

	"github.com/malikabdulaziz/tmn-backend/exceptions"
	"github.com/malikabdulaziz/tmn-backend/helpers"
	"github.com/malikabdulaziz/tmn-backend/models"
	repositoriesBuilding "github.com/malikabdulaziz/tmn-backend/repositories/building"
	"github.com/malikabdulaziz/tmn-backend/services/erp"
	webBuilding "github.com/malikabdulaziz/tmn-backend/web/building"
	"github.com/sirupsen/logrus"
)

type ServiceBuildingImpl struct {
	DB                          *sql.DB
	RepositoryBuildingInterface repositoriesBuilding.RepositoryBuildingInterface
	ERPClient                   *erp.ERPClient
	Logger                      *logrus.Logger
}

func NewServiceBuildingImpl(
	db *sql.DB,
	repositoryBuilding repositoriesBuilding.RepositoryBuildingInterface,
	erpClient *erp.ERPClient,
	logger *logrus.Logger,
) ServiceBuildingInterface {
	return &ServiceBuildingImpl{
		DB:                          db,
		RepositoryBuildingInterface: repositoryBuilding,
		ERPClient:                   erpClient,
		Logger:                      logger,
	}
}

// FindById retrieves a building by ID
func (service *ServiceBuildingImpl) FindById(ctx context.Context, id int) webBuilding.BuildingResponse {
	tx, err := service.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	building, err := service.RepositoryBuildingInterface.FindById(ctx, tx, id)
	if err == sql.ErrNoRows {
		panic(exceptions.NewNotFoundError("building not found"))
	}
	helpers.PanicIfError(err)

	return webBuilding.BuildingModelToBuildingResponse(building)
}

// FindAll retrieves all buildings with pagination
func (service *ServiceBuildingImpl) FindAll(ctx context.Context, request webBuilding.BuildingRequestFindAll) ([]webBuilding.BuildingResponse, int) {
	tx, err := service.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	buildings, err := service.RepositoryBuildingInterface.FindAll(ctx, tx, request.GetTake(), request.GetSkip(), request.GetOrderBy(), request.GetOrderDirection())
	helpers.PanicIfError(err)

	total, err := service.RepositoryBuildingInterface.CountAll(ctx, tx)
	helpers.PanicIfError(err)

	return webBuilding.BuildingModelsToListBuildingResponse(buildings), total
}

// Update updates user-editable fields only
func (service *ServiceBuildingImpl) Update(ctx context.Context, request webBuilding.UpdateBuildingRequest, id int) webBuilding.BuildingResponse {
	tx, err := service.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	// Verify building exists
	existingBuilding, err := service.RepositoryBuildingInterface.FindById(ctx, tx, id)
	if err == sql.ErrNoRows {
		panic(exceptions.NewNotFoundError("building not found"))
	}
	helpers.PanicIfError(err)

	// Update only user-editable fields
	existingBuilding.Sellable = request.Sellable
	existingBuilding.Connectivity = request.Connectivity
	existingBuilding.ResourceType = request.ResourceType

	building, err := service.RepositoryBuildingInterface.Update(ctx, tx, existingBuilding)
	helpers.PanicIfError(err)

	return webBuilding.BuildingModelToBuildingResponse(building)
}

// SyncFromERP fetches buildings from ERP and syncs them to the database
func (service *ServiceBuildingImpl) SyncFromERP(ctx context.Context) error {
	service.Logger.Info("Starting building sync from ERP")

	// Fetch buildings from ERP
	erpBuildings, err := service.ERPClient.FetchBuildings()
	if err != nil {
		service.Logger.WithError(err).Error("Failed to fetch buildings from ERP")
		return err
	}

	service.Logger.WithField("count", len(erpBuildings)).Info("Fetched buildings from ERP")

	// Sync each building
	syncedCount := 0
	createdCount := 0
	updatedCount := 0

	for _, erpBuilding := range erpBuildings {
		tx, err := service.DB.Begin()
		if err != nil {
			service.Logger.WithError(err).Error("Failed to start transaction")
			continue
		}

		// Check if building exists by external ID
		existingBuilding, err := service.RepositoryBuildingInterface.FindByExternalId(ctx, tx, erpBuilding.BuildingId)

		if err == sql.ErrNoRows {
			// Create new building
			newBuilding := models.Building{
				ExternalBuildingId: erpBuilding.BuildingId,
				IrisCode:           erpBuilding.IrisCode,
				Name:               erpBuilding.BuildingName,
				ProjectName:        erpBuilding.BuildingProject,
				Audience:           erpBuilding.AudienceActual,
				Impression:         erpBuilding.AudienceProjection,
				CbdArea:            erpBuilding.CbdArea,
				BuildingStatus:     erpBuilding.Eligible,
				CompetitorLocation: erpBuilding.CompetitorPresence != 0,
				SyncedAt:           time.Now().Format(time.RFC3339),
			}

			_, err = service.RepositoryBuildingInterface.Create(ctx, tx, newBuilding)
			if err != nil {
				service.Logger.WithError(err).WithField("building_id", erpBuilding.BuildingId).Error("Failed to create building")
				tx.Rollback()
				continue
			}

			createdCount++
		} else if err == nil {
			// Update existing building (ERP fields only)
			existingBuilding.ExternalBuildingId = erpBuilding.BuildingId
			existingBuilding.IrisCode = erpBuilding.IrisCode
			existingBuilding.Name = erpBuilding.BuildingName
			existingBuilding.ProjectName = erpBuilding.BuildingProject
			existingBuilding.Audience = erpBuilding.AudienceActual
			existingBuilding.Impression = erpBuilding.AudienceProjection
			existingBuilding.CbdArea = erpBuilding.CbdArea
			existingBuilding.BuildingStatus = erpBuilding.Eligible
			existingBuilding.CompetitorLocation = erpBuilding.CompetitorPresence != 0
			existingBuilding.SyncedAt = time.Now().Format(time.RFC3339)

			_, err = service.RepositoryBuildingInterface.UpdateFromSync(ctx, tx, existingBuilding)
			if err != nil {
				service.Logger.WithError(err).WithField("building_id", erpBuilding.BuildingId).Error("Failed to update building")
				tx.Rollback()
				continue
			}

			updatedCount++
		} else {
			service.Logger.WithError(err).WithField("building_id", erpBuilding.BuildingId).Error("Failed to check building existence")
			tx.Rollback()
			continue
		}

		err = tx.Commit()
		if err != nil {
			service.Logger.WithError(err).Error("Failed to commit transaction")
			continue
		}

		syncedCount++
	}

	service.Logger.WithFields(logrus.Fields{
		"synced":  syncedCount,
		"created": createdCount,
		"updated": updatedCount,
	}).Info("Building sync completed")

	return nil
}
