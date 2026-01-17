package building

import (
	"context"
	"database/sql"
	"sort"
	"strings"
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

	buildings, err := service.RepositoryBuildingInterface.FindAll(ctx, tx, request.GetTake(), request.GetSkip(), request.GetOrderBy(), request.GetOrderDirection(), request.GetSearch(), request.GetBuildingStatus(), request.GetSellable(), request.GetConnectivity(), request.GetResourceType(), request.GetCompetitorLocation(), request.GetCbdArea(), request.GetSubdistrict(), request.GetCitytown(), request.GetProvince(), request.GetGradeResource(), request.GetBuildingType())
	helpers.PanicIfError(err)

	total, err := service.RepositoryBuildingInterface.CountAll(ctx, tx, request.GetSearch(), request.GetBuildingStatus(), request.GetSellable(), request.GetConnectivity(), request.GetResourceType(), request.GetCompetitorLocation(), request.GetCbdArea(), request.GetSubdistrict(), request.GetCitytown(), request.GetProvince(), request.GetGradeResource(), request.GetBuildingType())
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

// calculateLcdPresenceStatus calculates the LCD presence status based on competitor fields, workflow state, and screen count
// Returns: "TMN", "Competitor", "CoExist", "Opportunity", or empty string if insufficient data
func calculateLcdPresenceStatus(competitorPresence, competitorExclusive bool, workflowState string, screenCount int) string {
	// Normalize workflow_state for case-insensitive comparison
	workflowStateNormalized := strings.ToLower(strings.TrimSpace(workflowState))
	isBastSigned := workflowStateNormalized == "bast signed"

	// 1. TMN Check (requires all fields)
	if workflowState != "" {
		if isBastSigned && !competitorPresence && !competitorExclusive {
			return "TMN"
		}
	}

	// 2. Competitor Check (needs competitor + workflow_state)
	// If workflow_state is empty, assume != "BAST Signed" (assumption)
	if workflowState == "" || !isBastSigned {
		if competitorPresence || competitorExclusive {
			return "Competitor"
		}
	}

	// 3. CoExist Check (requires all fields)
	if workflowState != "" && screenCount > 0 {
		if isBastSigned && !competitorExclusive && competitorPresence {
			return "CoExist"
		}
	}

	// 4. Opportunity Check (needs competitor + workflow_state)
	// If workflow_state is empty, assume != "BAST Signed" (assumption)
	if workflowState == "" || !isBastSigned {
		if !competitorPresence && !competitorExclusive {
			return "Opportunity"
		}
	}

	// Default: Return empty string if no conditions match
	return ""
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

	// Fetch building proposals from ERP
	erpBuildingProposals, err := service.ERPClient.FetchBuildingProposals()
	if err != nil {
		service.Logger.WithError(err).Error("Failed to fetch building proposals from ERP")
		return err
	}

	service.Logger.WithField("count", len(erpBuildingProposals)).Info("Fetched building proposals from ERP")

	// Handle duplicate acquisitions: sort by modified timestamp (descending) and group by building_project
	// Keep only the most recent acquisition for each building_project
	workflowStateMap := make(map[string]string) // building_project -> workflow_state

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

	// Create map of building_project -> workflow_state (taking first/latest for each project)
	for _, acquisition := range erpAcquisitions {
		if acquisition.BuildingProject != "" {
			// Only add if not already in map (since sorted, first one is latest)
			if _, exists := workflowStateMap[acquisition.BuildingProject]; !exists {
				workflowStateMap[acquisition.BuildingProject] = acquisition.WorkflowState
			}
		}
	}

	service.Logger.WithField("unique_projects", len(workflowStateMap)).Info("Processed acquisitions (deduplicated)")

	// Handle duplicate building proposals: sort by modified timestamp (descending) and group by building_project
	// Keep only the most recent proposal for each building_project
	screenCountMap := make(map[string]int) // building_project -> number_of_screen

	// Sort building proposals by modified timestamp (descending)
	sort.Slice(erpBuildingProposals, func(i, j int) bool {
		timeI, errI := time.Parse("2006-01-02 15:04:05.999999", erpBuildingProposals[i].Modified)
		timeJ, errJ := time.Parse("2006-01-02 15:04:05.999999", erpBuildingProposals[j].Modified)

		// If parsing fails, treat as older
		if errI != nil {
			return false
		}
		if errJ != nil {
			return true
		}

		return timeI.After(timeJ)
	})

	// Create map of building_project -> number_of_screen (taking first/latest for each project)
	for _, proposal := range erpBuildingProposals {
		if proposal.BuildingProject != "" {
			// Only add if not already in map (since sorted, first one is latest)
			if _, exists := screenCountMap[proposal.BuildingProject]; !exists {
				screenCountMap[proposal.BuildingProject] = proposal.NumberOfScreen
			}
		}
	}

	service.Logger.WithField("unique_projects", len(screenCountMap)).Info("Processed building proposals (deduplicated)")

	// Sync each building
	syncedCount := 0
	createdCount := 0
	updatedCount := 0

	for _, erpBuilding := range erpBuildings {
		// Get workflow state and screen count for logging
		workflowState := ""
		screenCount := 0
		if erpBuilding.BuildingProject != "" {
			if ws, exists := workflowStateMap[erpBuilding.BuildingProject]; exists {
				workflowState = ws
			}
			if count, exists := screenCountMap[erpBuilding.BuildingProject]; exists {
				screenCount = count
			}
		}

		// Calculate LCD presence status for logging
		calculatedStatus := calculateLcdPresenceStatus(
			erpBuilding.CompetitorPresence != 0,
			erpBuilding.CompetitorExclusive != 0,
			workflowState,
			screenCount,
		)

		// Log building data for debugging
		service.Logger.WithFields(logrus.Fields{
			"building_name":        erpBuilding.BuildingName,
			"building_project":     erpBuilding.BuildingProject,
			"screen_count":         screenCount,
			"workflow_state":       workflowState,
			"competitor_presence":  erpBuilding.CompetitorPresence,
			"competitor_exclusive": erpBuilding.CompetitorExclusive,
			"lcd_presence_status":  calculatedStatus,
			"external_building_id": erpBuilding.BuildingId,
		}).Info("Processing building from ERP")

		tx, err := service.DB.Begin()
		if err != nil {
			service.Logger.WithError(err).Error("Failed to start transaction")
			continue
		}

		// Check if building exists by external ID
		existingBuilding, err := service.RepositoryBuildingInterface.FindByExternalId(ctx, tx, erpBuilding.BuildingId)

		if err == sql.ErrNoRows {
			// Use workflow_state as building_status for now (keeping existing logic)
			buildingStatus := workflowState

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
				ExternalBuildingId:  erpBuilding.BuildingId,
				IrisCode:            erpBuilding.IrisCode,
				Name:                erpBuilding.BuildingName,
				ProjectName:         erpBuilding.BuildingProject,
				Audience:            erpBuilding.AudienceActual,
				Impression:          erpBuilding.AudienceProjection,
				CbdArea:             erpBuilding.CbdArea,
				Subdistrict:         erpBuilding.Subdistrict,
				Citytown:            erpBuilding.Citytown,
				Province:            erpBuilding.Province,
				GradeResource:       erpBuilding.GradeResource,
				BuildingType:        erpBuilding.BuildingType,
				CompletionYear:      erpBuilding.CompletionYear,
				Latitude:            erpBuilding.Latitude,
				Longitude:           erpBuilding.Longitude,
				BuildingStatus:      buildingStatus,
				CompetitorLocation:  erpBuilding.CompetitorPresence != 0,
				CompetitorExclusive: erpBuilding.CompetitorExclusive != 0,
				CompetitorPresence:  erpBuilding.CompetitorPresence != 0,
				LcdPresenceStatus:   calculatedStatus,
				Images:              images,
				SyncedAt:            time.Now().Format(time.RFC3339),
			}

			_, err = service.RepositoryBuildingInterface.Create(ctx, tx, newBuilding)
			if err != nil {
				service.Logger.WithError(err).WithField("building_id", erpBuilding.BuildingId).Error("Failed to create building")
				tx.Rollback()
				continue
			}

			createdCount++
		} else if err == nil {
			// Use workflow_state as building_status for now (keeping existing logic)
			buildingStatus := workflowState

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
			existingBuilding.Subdistrict = erpBuilding.Subdistrict
			existingBuilding.Citytown = erpBuilding.Citytown
			existingBuilding.Province = erpBuilding.Province
			existingBuilding.GradeResource = erpBuilding.GradeResource
			existingBuilding.BuildingType = erpBuilding.BuildingType
			existingBuilding.CompletionYear = erpBuilding.CompletionYear
			// Zero-preservation logic: only update latitude/longitude if ERP provides non-zero values
			if erpBuilding.Latitude != 0 {
				existingBuilding.Latitude = erpBuilding.Latitude
			}
			if erpBuilding.Longitude != 0 {
				existingBuilding.Longitude = erpBuilding.Longitude
			}
			existingBuilding.BuildingStatus = buildingStatus
			existingBuilding.CompetitorLocation = erpBuilding.CompetitorPresence != 0
			existingBuilding.CompetitorExclusive = erpBuilding.CompetitorExclusive != 0
			existingBuilding.CompetitorPresence = erpBuilding.CompetitorPresence != 0
			existingBuilding.LcdPresenceStatus = calculatedStatus
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
	columns := []string{"building_status", "sellable", "connectivity", "resource_type", "cbd_area", "subdistrict", "citytown", "province", "grade_resource", "building_type"}

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

// FindAllForMapping retrieves all buildings for mapping with filters
func (service *ServiceBuildingImpl) FindAllForMapping(ctx context.Context, request webBuilding.MappingBuildingRequest) webBuilding.MappingBuildingsResponse {
	tx, err := service.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	buildings, err := service.RepositoryBuildingInterface.FindAllForMapping(
		ctx,
		tx,
		request.GetBuildingType(),
		request.GetBuildingGrade(),
		request.GetYear(),
		request.GetSubdistrict(),
		request.GetProgress(),
		request.GetSellable(),
		request.GetConnectivity(),
		request.GetLCDPresence(),
	)
	helpers.PanicIfError(err)

	// Convert to mapping response and calculate totals
	mappingBuildings := make([]webBuilding.MappingBuildingResponse, 0, len(buildings))
	totalApartment := 0
	totalHotel := 0
	totalOffice := 0
	totalRetail := 0
	totalOthers := 0

	for _, building := range buildings {
		// Convert images
		images := make([]webBuilding.MappingBuildingImageResponse, 0, len(building.Images))
		for _, img := range building.Images {
			images = append(images, webBuilding.MappingBuildingImageResponse{
				Name: img.Name,
				Path: img.Path,
			})
		}

		// Construct address from location fields
		addressParts := []string{}
		if building.Subdistrict != "" {
			addressParts = append(addressParts, building.Subdistrict)
		}
		if building.Citytown != "" {
			addressParts = append(addressParts, building.Citytown)
		}
		if building.Province != "" {
			addressParts = append(addressParts, building.Province)
		}
		address := ""
		if len(addressParts) > 0 {
			address = addressParts[0]
			for i := 1; i < len(addressParts); i++ {
				address += ", " + addressParts[i]
			}
		}

		mappingBuilding := webBuilding.MappingBuildingResponse{
			Id:                building.Id,
			Name:              building.Name,
			BuildingType:      building.BuildingType,
			GradeResource:     building.GradeResource,
			CompletionYear:    building.CompletionYear,
			Subdistrict:       building.Subdistrict,
			Citytown:          building.Citytown,
			Province:          building.Province,
			Address:           address,
			BuildingStatus:    building.BuildingStatus,
			Sellable:          building.Sellable,
			Connectivity:      building.Connectivity,
			Latitude:          building.Latitude,
			Longitude:         building.Longitude,
			LcdPresenceStatus: building.LcdPresenceStatus,
			Images:            images,
		}

		mappingBuildings = append(mappingBuildings, mappingBuilding)

		// Count by building type
		switch building.BuildingType {
		case "Apartment":
			totalApartment++
		case "Hotel":
			totalHotel++
		case "Office":
			totalOffice++
		case "Retail":
			totalRetail++
		default:
			totalOthers++
		}
	}

	return webBuilding.MappingBuildingsResponse{
		Data:           mappingBuildings,
		TotalApartment: totalApartment,
		TotalHotel:     totalHotel,
		TotalOffice:    totalOffice,
		TotalRetail:    totalRetail,
		TotalOthers:    totalOthers,
	}
}
