package building

import (
	"context"
	"database/sql"
	"sort"
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

	buildings, err := service.RepositoryBuildingInterface.FindAll(ctx, tx, request.GetTake(), request.GetSkip(), request.GetOrderBy(), request.GetOrderDirection(), request.GetSearch(), request.GetBuildingStatus(), request.GetSellable(), request.GetConnectivity(), request.GetResourceType(), request.GetCompetitorLocation(), request.GetCbdArea())
	helpers.PanicIfError(err)

	total, err := service.RepositoryBuildingInterface.CountAll(ctx, tx, request.GetSearch(), request.GetBuildingStatus(), request.GetSellable(), request.GetConnectivity(), request.GetResourceType(), request.GetCompetitorLocation(), request.GetCbdArea())
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

	// Fetch acquisitions from ERP
	erpAcquisitions, err := service.ERPClient.FetchAcquisitions()
	if err != nil {
		service.Logger.WithError(err).Error("Failed to fetch acquisitions from ERP")
		return err
	}

	service.Logger.WithField("count", len(erpAcquisitions)).Info("Fetched acquisitions from ERP")

	// Handle duplicate acquisitions: sort by modified timestamp (descending) and group by building_project
	// Keep only the most recent acquisition for each building_project
	acquisitionMap := make(map[string]string) // building_project -> status

	// Sort acquisitions by modified timestamp (descending)
	sort.Slice(erpAcquisitions, func(i, j int) bool {
		timeI, errI := time.Parse("2006-01-02 15:04:05.999999", erpAcquisitions[i].Modified)
		timeJ, errJ := time.Parse("2006-01-02 15:04:05.999999", erpAcquisitions[j].Modified)

		// If parsing fails, treat as older
		if errI != nil {
			return false
		}
		if errJ != nil {
			return true
		}

		return timeI.After(timeJ)
	})

	// Create map of building_project -> status (taking first/latest for each project)
	for _, acquisition := range erpAcquisitions {
		if acquisition.BuildingProject != "" {
			// Only add if not already in map (since sorted, first one is latest)
			if _, exists := acquisitionMap[acquisition.BuildingProject]; !exists {
				acquisitionMap[acquisition.BuildingProject] = acquisition.Status
			}
		}
	}

	service.Logger.WithField("unique_projects", len(acquisitionMap)).Info("Processed acquisitions (deduplicated)")

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
			// Get building status from acquisition map
			buildingStatus := ""
			if erpBuilding.BuildingProject != "" {
				if status, exists := acquisitionMap[erpBuilding.BuildingProject]; exists {
					buildingStatus = status
				}
			}

			// Convert ERP image fields to JSON array format
			images := []models.BuildingImage{}
			if erpBuilding.FrontSidePhoto != "" {
				images = append(images, models.BuildingImage{Name: "front_side", Path: erpBuilding.FrontSidePhoto})
			}
			if erpBuilding.BackSidePhoto != "" {
				images = append(images, models.BuildingImage{Name: "back_side", Path: erpBuilding.BackSidePhoto})
			}
			if erpBuilding.LeftSidePhoto != "" {
				images = append(images, models.BuildingImage{Name: "left_side", Path: erpBuilding.LeftSidePhoto})
			}
			if erpBuilding.RightSidePhoto != "" {
				images = append(images, models.BuildingImage{Name: "right_side", Path: erpBuilding.RightSidePhoto})
			}

			// Create new building
			newBuilding := models.Building{
				ExternalBuildingId: erpBuilding.BuildingId,
				IrisCode:           erpBuilding.IrisCode,
				Name:               erpBuilding.BuildingName,
				ProjectName:        erpBuilding.BuildingProject,
				Audience:           erpBuilding.AudienceActual,
				Impression:         erpBuilding.AudienceProjection,
				CbdArea:            erpBuilding.CbdArea,
				BuildingStatus:     buildingStatus,
				CompetitorLocation: erpBuilding.CompetitorPresence != 0,
				Images:             images,
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
			// Get building status from acquisition map
			buildingStatus := ""
			if erpBuilding.BuildingProject != "" {
				if status, exists := acquisitionMap[erpBuilding.BuildingProject]; exists {
					buildingStatus = status
				}
			}

			// Convert ERP image fields to JSON array format
			images := []models.BuildingImage{}
			if erpBuilding.FrontSidePhoto != "" {
				images = append(images, models.BuildingImage{Name: "front_side", Path: erpBuilding.FrontSidePhoto})
			}
			if erpBuilding.BackSidePhoto != "" {
				images = append(images, models.BuildingImage{Name: "back_side", Path: erpBuilding.BackSidePhoto})
			}
			if erpBuilding.LeftSidePhoto != "" {
				images = append(images, models.BuildingImage{Name: "left_side", Path: erpBuilding.LeftSidePhoto})
			}
			if erpBuilding.RightSidePhoto != "" {
				images = append(images, models.BuildingImage{Name: "right_side", Path: erpBuilding.RightSidePhoto})
			}

			// Update existing building (ERP fields only)
			existingBuilding.ExternalBuildingId = erpBuilding.BuildingId
			existingBuilding.IrisCode = erpBuilding.IrisCode
			existingBuilding.Name = erpBuilding.BuildingName
			existingBuilding.ProjectName = erpBuilding.BuildingProject
			existingBuilding.Audience = erpBuilding.AudienceActual
			existingBuilding.Impression = erpBuilding.AudienceProjection
			existingBuilding.CbdArea = erpBuilding.CbdArea
			existingBuilding.BuildingStatus = buildingStatus
			existingBuilding.CompetitorLocation = erpBuilding.CompetitorPresence != 0
			existingBuilding.Images = images
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

// GetFilterOptions returns distinct values for filter dropdowns
func (service *ServiceBuildingImpl) GetFilterOptions(ctx context.Context) map[string][]string {
	tx, err := service.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	filterOptions := make(map[string][]string)

	// Get distinct values for each filter field
	columns := []string{"building_status", "sellable", "connectivity", "resource_type", "cbd_area"}

	for _, column := range columns {
		values, err := service.RepositoryBuildingInterface.GetDistinctValues(ctx, tx, column)
		if err != nil {
			service.Logger.WithError(err).WithField("column", column).Error("Failed to get distinct values")
			continue
		}
		filterOptions[column] = values
	}

	return filterOptions
}
