package building

import (
	"context"
	"database/sql"
	"encoding/json"
	"math"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/malikabdulaziz/tmn-backend/exceptions"
	"github.com/malikabdulaziz/tmn-backend/helpers"
	"github.com/malikabdulaziz/tmn-backend/models"
	repositoriesBuilding "github.com/malikabdulaziz/tmn-backend/repositories/building"
	repositoriesBuildingProject "github.com/malikabdulaziz/tmn-backend/repositories/buildingproject"
	repositoriesPOI "github.com/malikabdulaziz/tmn-backend/repositories/poi"
	"github.com/malikabdulaziz/tmn-backend/services/erp"
	webBuilding "github.com/malikabdulaziz/tmn-backend/web/building"
	"github.com/sirupsen/logrus"
	"github.com/xuri/excelize/v2"
)

const (
	// maxWorkers defines the number of concurrent workers for building sync
	maxWorkers = 10
)

type ServiceBuildingImpl struct {
	DB                          *sql.DB
	RepositoryBuildingInterface repositoriesBuilding.RepositoryBuildingInterface
	RepositoryPOIInterface      repositoriesPOI.RepositoryPOIInterface
	// Needed by the spreadsheet import, which resolves a Project ID IRIS to a
	// project and creates an empty one when the code is not known yet.
	RepositoryBuildingProject repositoriesBuildingProject.RepositoryBuildingProjectInterface
	ERPClient                 *erp.ERPClient
	Logger                    *logrus.Logger
}

// syncCounters holds thread-safe counters for sync operations
type syncCounters struct {
	mu           sync.Mutex
	syncedCount  int
	createdCount int
	updatedCount int
	errorCount   int
	errors       []errorInfo
}

// errorInfo holds error information for a specific building
type errorInfo struct {
	buildingID   string
	buildingName string
	error        error
}

// incrementSynced atomically increments the synced counter
func (c *syncCounters) incrementSynced() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.syncedCount++
}

// incrementCreated atomically increments the created counter
func (c *syncCounters) incrementCreated() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.createdCount++
}

// incrementUpdated atomically increments the updated counter
func (c *syncCounters) incrementUpdated() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.updatedCount++
}

// addError atomically adds an error to the error list
func (c *syncCounters) addError(buildingID, buildingName string, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.errorCount++
	c.errors = append(c.errors, errorInfo{
		buildingID:   buildingID,
		buildingName: buildingName,
		error:        err,
	})
}

func NewServiceBuildingImpl(
	db *sql.DB,
	repositoryBuilding repositoriesBuilding.RepositoryBuildingInterface,
	repositoryBuildingProject repositoriesBuildingProject.RepositoryBuildingProjectInterface,
	repositoryPOI repositoriesPOI.RepositoryPOIInterface,
	erpClient *erp.ERPClient,
	logger *logrus.Logger,
) ServiceBuildingInterface {
	return &ServiceBuildingImpl{
		DB:                          db,
		RepositoryBuildingInterface: repositoryBuilding,
		RepositoryBuildingProject:   repositoryBuildingProject,
		RepositoryPOIInterface:      repositoryPOI,
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

	buildings, err := service.RepositoryBuildingInterface.FindAll(ctx, tx, request.GetTake(), request.GetSkip(), request.GetOrderBy(), request.GetOrderDirection(), request.GetSearch(), request.GetBuildingStatus(), request.GetSellable(), request.GetConnectivity(), request.GetResourceType(), request.GetCompetitorLocation(), request.GetCbdArea(), request.GetSubdistrict(), request.GetCitytown(), request.GetProvince(), request.GetGradeResource(), request.GetBuildingType(), request.GetExcludeIds())
	helpers.PanicIfError(err)

	total, err := service.RepositoryBuildingInterface.CountAll(ctx, tx, request.GetSearch(), request.GetBuildingStatus(), request.GetSellable(), request.GetConnectivity(), request.GetResourceType(), request.GetCompetitorLocation(), request.GetCbdArea(), request.GetSubdistrict(), request.GetCitytown(), request.GetProvince(), request.GetGradeResource(), request.GetBuildingType(), request.GetExcludeIds())
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

// calculateLcdPresenceStatus calculates the LCD presence status based on competitor fields and workflow state.
// Returns: "TMN", "Competitor", "CoExist", "Opportunity", or empty string for unmatched/contradictory data.
// Note: "BAST Signed" + competitor_exclusive=1 is logically contradictory and falls through to "".
func calculateLcdPresenceStatus(competitorPresence, competitorExclusive bool, workflowState string) string {
	workflowStateNormalized := strings.ToLower(strings.TrimSpace(workflowState))
	isBastSigned := workflowStateNormalized == "bast signed"

	if isBastSigned {
		if !competitorPresence && !competitorExclusive {
			return "TMN"
		}
		if competitorPresence && !competitorExclusive {
			return "CoExist"
		}
		return ""
	}

	if competitorPresence || competitorExclusive {
		return "Competitor"
	}
	return "Opportunity"
}

// trimERPBuilding strips surrounding whitespace from every text field ERP supplies.
//
// It is deliberately applied to all of them rather than only the field known to be
// dirty: whitespace in a name breaks a search, in a citytown breaks a map filter, and
// in an IRIS code breaks the price import's lookup. None of those are worth finding
// one at a time.
func trimERPBuilding(b *erp.ERPBuilding) {
	b.BuildingId = strings.TrimSpace(b.BuildingId)
	b.IrisCode = strings.TrimSpace(b.IrisCode)
	b.BuildingName = strings.TrimSpace(b.BuildingName)
	b.BuildingProject = strings.TrimSpace(b.BuildingProject)
	b.CbdArea = strings.TrimSpace(b.CbdArea)
	b.Subdistrict = strings.TrimSpace(b.Subdistrict)
	b.Citytown = strings.TrimSpace(b.Citytown)
	b.Province = strings.TrimSpace(b.Province)
	b.GradeResource = strings.TrimSpace(b.GradeResource)
	b.BuildingType = strings.TrimSpace(b.BuildingType)
}

// SyncFromERP refreshes building PHOTOS from ERP, and nothing else.
//
// Buildings are maintained here now, by spreadsheet. This feed used to write every
// column on a timer; left that way it would quietly revert people's uploads between
// runs, because it ran against a table they had started editing by hand.
//
// Photos are the exception: they are file paths served from ERP, and moving them
// in-house is separate work. Until that happens the sync keeps them current and
// touches `images` and `synced_at` only.
//
// A building ERP knows about but this database does not is SKIPPED, not created. The
// spreadsheet owns which buildings exist; a feed that could still add rows would mean
// two sources of truth for the roster, which is exactly what the cutover removes.
//
// Acquisitions and building proposals are no longer fetched. They supplied
// building_status and a screen count, both of which now come from the spreadsheet.
func (service *ServiceBuildingImpl) SyncFromERP(ctx context.Context) error {
	service.Logger.Info("Starting building photo sync from ERP")

	erpBuildings, err := service.ERPClient.FetchBuildings()
	if err != nil {
		service.Logger.WithError(err).Error("Failed to fetch buildings from ERP")

		return err
	}

	service.Logger.WithField("count", len(erpBuildings)).Info("Fetched buildings from ERP")

	tx, err := service.DB.Begin()
	if err != nil {
		return err
	}
	defer helpers.CommitOrRollback(tx)

	var updated, unchanged, skipped, failed int

	for _, erpBuilding := range erpBuildings {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		trimERPBuilding(&erpBuilding)
		if erpBuilding.BuildingId == "" {
			skipped++

			continue
		}

		existing, err := service.RepositoryBuildingInterface.FindByExternalId(ctx, tx, erpBuilding.BuildingId)
		if err != nil {
			// Not ours to create -- the spreadsheet decides which buildings exist.
			skipped++

			continue
		}

		images := erpImages(erpBuilding)
		if sameImages(existing.Images, images) {
			unchanged++

			continue
		}

		encoded, err := json.Marshal(images)
		if err != nil {
			failed++

			continue
		}

		if len(images) == 0 {
			encoded = nil
		}

		if err := service.RepositoryBuildingInterface.UpdateImagesFromSync(ctx, tx, existing.Id, string(encoded)); err != nil {
			service.Logger.WithFields(logrus.Fields{
				"building_id": erpBuilding.BuildingId,
				"error":       err.Error(),
			}).Error("Failed to update building photos")
			failed++

			continue
		}

		updated++
	}

	service.Logger.WithFields(logrus.Fields{
		"updated":   updated,
		"unchanged": unchanged,
		"skipped":   skipped,
		"failed":    failed,
		"total":     len(erpBuildings),
	}).Info("Building photo sync completed")

	return nil
}

// erpImages collects the four photo paths ERP supplies, dropping the empty ones.
func erpImages(b erp.ERPBuilding) []models.BuildingImage {
	images := []models.BuildingImage{}
	for _, photo := range []struct {
		name string
		path string
	}{
		{"front", b.FrontSidePhoto},
		{"back", b.BackSidePhoto},
		{"left", b.LeftSidePhoto},
		{"right_side", b.RightSidePhoto},
	} {
		if photo.path != "" {
			images = append(images, models.BuildingImage{Name: photo.name, Path: photo.path})
		}
	}

	return images
}

// sameImages avoids writing a row whose photos have not moved. Without it every run
// would bump updated_at on all 3,747 buildings and make "when did this last change"
// meaningless on a table people now edit by hand.
func sameImages(a, b []models.BuildingImage) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i].Name != b[i].Name || a[i].Path != b[i].Path {
			return false
		}
	}

	return true
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

	var latPtr *float64
	var lngPtr *float64
	var radiusPtr *int
	var poiPoints []struct {
		Lat float64
		Lng float64
	}
	var polygonPoints []struct {
		Lat float64
		Lng float64
	}

	// Polygon filter takes priority: when set, use only polygon for spatial filter
	if polygonStr := request.GetPolygon(); polygonStr != "" {
		var parsed []struct {
			Lat float64 `json:"lat"`
			Lng float64 `json:"lng"`
		}
		if err := json.Unmarshal([]byte(polygonStr), &parsed); err == nil && len(parsed) >= 3 {
			polygonPoints = make([]struct {
				Lat float64
				Lng float64
			}, len(parsed))
			for i, p := range parsed {
				polygonPoints[i] = struct {
					Lat float64
					Lng float64
				}{Lat: p.Lat, Lng: p.Lng}
			}
		}
	}

	// When polygon is not set, use POI / lat-lng / radius
	if len(polygonPoints) == 0 {
		if poiIdStr := request.GetPOIId(); poiIdStr != "" {
			for _, idStr := range strings.Split(poiIdStr, ",") {
				idStr = strings.TrimSpace(idStr)
				poiId, err := strconv.Atoi(idStr)
				if err != nil || poiId <= 0 {
					continue
				}
				poi, err := service.RepositoryPOIInterface.FindById(ctx, tx, poiId)
				if err != nil || len(poi.Points) == 0 {
					continue
				}
				for _, point := range poi.Points {
					poiPoints = append(poiPoints, struct {
						Lat float64
						Lng float64
					}{
						Lat: point.Latitude,
						Lng: point.Longitude,
					})
				}
			}
		}

		if len(poiPoints) == 0 {
			if latStr := request.GetLat(); latStr != "" {
				if lat, err := strconv.ParseFloat(latStr, 64); err == nil {
					latPtr = &lat
				}
			}
			if lngStr := request.GetLng(); lngStr != "" {
				if lng, err := strconv.ParseFloat(lngStr, 64); err == nil {
					lngPtr = &lng
				}
			}
		}

		if radiusStr := request.GetRadius(); radiusStr != "" {
			if radius, err := strconv.Atoi(radiusStr); err == nil && radius > 0 {
				radiusPtr = &radius
			}
		}
	}

	// Parse optional map bounds (viewport); only apply when all four are valid and min < max
	var minLatPtr, maxLatPtr, minLngPtr, maxLngPtr *float64
	if minLatStr := request.GetMinLat(); minLatStr != "" {
		if v, err := strconv.ParseFloat(minLatStr, 64); err == nil {
			minLatPtr = &v
		}
	}
	if maxLatStr := request.GetMaxLat(); maxLatStr != "" {
		if v, err := strconv.ParseFloat(maxLatStr, 64); err == nil {
			maxLatPtr = &v
		}
	}
	if minLngStr := request.GetMinLng(); minLngStr != "" {
		if v, err := strconv.ParseFloat(minLngStr, 64); err == nil {
			minLngPtr = &v
		}
	}
	if maxLngStr := request.GetMaxLng(); maxLngStr != "" {
		if v, err := strconv.ParseFloat(maxLngStr, 64); err == nil {
			maxLngPtr = &v
		}
	}
	if minLatPtr != nil && maxLatPtr != nil && minLngPtr != nil && maxLngPtr != nil {
		if *minLatPtr > *maxLatPtr || *minLngPtr > *maxLngPtr {
			minLatPtr, maxLatPtr, minLngPtr, maxLngPtr = nil, nil, nil, nil
		}
	}

	// Data: buildings in view (with bounds when provided)
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
		request.GetSalesPackageIds(),
		request.GetBuildingRestrictionIds(),
		latPtr,
		lngPtr,
		radiusPtr,
		poiPoints,
		polygonPoints,
		minLatPtr,
		maxLatPtr,
		minLngPtr,
		maxLngPtr,
	)
	helpers.PanicIfError(err)

	// Totals: counts by building_type for the full filter set (no bounds), so totals are not scoped to viewport
	var buildingsForTotals []models.Building
	if minLatPtr != nil && maxLatPtr != nil && minLngPtr != nil && maxLngPtr != nil {
		buildingsForTotals, err = service.RepositoryBuildingInterface.FindAllForMapping(
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
			request.GetSalesPackageIds(),
			request.GetBuildingRestrictionIds(),
			latPtr,
			lngPtr,
			radiusPtr,
			poiPoints,
			polygonPoints,
			nil, nil, nil, nil,
		)
		helpers.PanicIfError(err)
	} else {
		buildingsForTotals = buildings
	}

	// Use a map for dynamic totals - count all building types (from full filter set, no bounds)
	totalsMap := make(map[string]int)
	for _, building := range buildingsForTotals {
		buildingType := building.BuildingType
		if buildingType == "" {
			buildingType = "Other"
		}
		key := strings.ToLower(buildingType)
		totalsMap[key]++
	}

	// Convert to mapping response (Data = buildings in view)
	mappingBuildings := make([]webBuilding.MappingBuildingResponse, 0, len(buildings))

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
			Id:                 building.Id,
			ExternalBuildingId: building.ExternalBuildingId,
			Name:               building.Name,
			BuildingType:       building.BuildingType,
			GradeResource:      building.GradeResource,
			CompletionYear:     building.CompletionYear,
			Subdistrict:        building.Subdistrict,
			Citytown:           building.Citytown,
			Province:           building.Province,
			Address:            address,
			BuildingStatus:     building.BuildingStatus,
			Sellable:           building.Sellable,
			Connectivity:       building.Connectivity,
			Latitude:           building.Latitude,
			Longitude:          building.Longitude,
			LcdPresenceStatus:  building.LcdPresenceStatus,
			Images:             images,
		}

		mappingBuildings = append(mappingBuildings, mappingBuilding)
	}

	return webBuilding.MappingBuildingsResponse{
		Data:   mappingBuildings,
		Totals: totalsMap,
	}
}

// ExportForMapping returns Excel file bytes for the given building IDs
func (service *ServiceBuildingImpl) ExportForMapping(ctx context.Context, ids []int) ([]byte, error) {
	tx, err := service.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer helpers.CommitOrRollback(tx)

	buildings, err := service.RepositoryBuildingInterface.FindByIds(ctx, tx, ids)
	if err != nil {
		return nil, err
	}

	f := excelize.NewFile()
	sheetName := "Buildings"
	const sheet = "Sheet1"
	headers := []string{"ID", "Name", "Building Type", "Grade", "Completion Year", "Subdistrict", "City", "Province", "Address", "Status", "Sellable", "Connectivity", "Latitude", "Longitude", "LCD Presence"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = f.SetCellValue(sheet, cell, h)
	}
	for row, building := range buildings {
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
		address := strings.Join(addressParts, ", ")
		rowIdx := row + 2
		_ = f.SetCellValue(sheet, mustCell(1, rowIdx), building.Id)
		_ = f.SetCellValue(sheet, mustCell(2, rowIdx), building.Name)
		_ = f.SetCellValue(sheet, mustCell(3, rowIdx), building.BuildingType)
		_ = f.SetCellValue(sheet, mustCell(4, rowIdx), building.GradeResource)
		_ = f.SetCellValue(sheet, mustCell(5, rowIdx), building.CompletionYear)
		_ = f.SetCellValue(sheet, mustCell(6, rowIdx), building.Subdistrict)
		_ = f.SetCellValue(sheet, mustCell(7, rowIdx), building.Citytown)
		_ = f.SetCellValue(sheet, mustCell(8, rowIdx), building.Province)
		_ = f.SetCellValue(sheet, mustCell(9, rowIdx), address)
		_ = f.SetCellValue(sheet, mustCell(10, rowIdx), building.BuildingStatus)
		_ = f.SetCellValue(sheet, mustCell(11, rowIdx), building.Sellable)
		_ = f.SetCellValue(sheet, mustCell(12, rowIdx), building.Connectivity)
		_ = f.SetCellValue(sheet, mustCell(13, rowIdx), building.Latitude)
		_ = f.SetCellValue(sheet, mustCell(14, rowIdx), building.Longitude)
		_ = f.SetCellValue(sheet, mustCell(15, rowIdx), building.LcdPresenceStatus)
	}
	if sheetName != sheet {
		_ = f.SetSheetName(sheet, sheetName)
	}
	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func mustCell(col, row int) string {
	s, _ := excelize.CoordinatesToCellName(col, row)
	return s
}

// ExportForMappingWithFilters returns Excel bytes for all buildings matching the request (no bounds).
func (service *ServiceBuildingImpl) ExportForMappingWithFilters(ctx context.Context, request webBuilding.MappingBuildingRequest) ([]byte, error) {
	resp := service.FindAllForMapping(ctx, request)
	return buildExcelFromMappingBuildings(resp.Data)
}

// GetLCDPresenceSummary returns building counts and percentages grouped by city and LCD presence status
func (service *ServiceBuildingImpl) GetLCDPresenceSummary(ctx context.Context) webBuilding.LCDPresenceSummaryResponse {
	tx, err := service.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	rawRows, err := service.RepositoryBuildingInterface.GetLCDPresenceSummary(ctx, tx)
	helpers.PanicIfError(err)

	type cityData struct {
		byStatus map[string]int
		total    int
	}
	cityMap := make(map[string]*cityData)
	cityOrder := []string{}

	for _, row := range rawRows {
		if _, exists := cityMap[row.Citytown]; !exists {
			cityMap[row.Citytown] = &cityData{byStatus: make(map[string]int)}
			cityOrder = append(cityOrder, row.Citytown)
		}
		cityMap[row.Citytown].byStatus[row.LcdPresenceStatus] += row.Count
		cityMap[row.Citytown].total += row.Count
	}

	sort.Strings(cityOrder)

	summaries := make([]webBuilding.LCDPresenceCitySummary, 0, len(cityOrder))
	grandTotal := 0
	grandByStatus := make(map[string]int)

	for _, city := range cityOrder {
		cd := cityMap[city]
		percentages := make(map[string]float64, len(cd.byStatus))
		for status, count := range cd.byStatus {
			if cd.total > 0 {
				percentages[status] = math.Round(float64(count) / float64(cd.total) * 100)
			}
		}
		summaries = append(summaries, webBuilding.LCDPresenceCitySummary{
			Citytown:    city,
			Total:       cd.total,
			ByStatus:    cd.byStatus,
			Percentages: percentages,
		})
		grandTotal += cd.total
		for status, count := range cd.byStatus {
			grandByStatus[status] += count
		}
	}

	grandPercentages := make(map[string]float64, len(grandByStatus))
	for status, count := range grandByStatus {
		if grandTotal > 0 {
			grandPercentages[status] = math.Round(float64(count) / float64(grandTotal) * 100)
		}
	}

	return webBuilding.LCDPresenceSummaryResponse{
		Data: summaries,
		Totals: webBuilding.LCDPresenceTotals{
			Total:       grandTotal,
			ByStatus:    grandByStatus,
			Percentages: grandPercentages,
		},
	}
}

// FindAllDropdown retrieves all buildings with only id, name, and building_type for dropdown use
func (service *ServiceBuildingImpl) FindAllDropdown(ctx context.Context) []webBuilding.BuildingDropdownResponse {
	tx, err := service.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	buildings, err := service.RepositoryBuildingInterface.FindAllDropdown(ctx, tx)
	helpers.PanicIfError(err)

	responses := make([]webBuilding.BuildingDropdownResponse, 0, len(buildings))
	for _, b := range buildings {
		responses = append(responses, webBuilding.BuildingDropdownResponse{
			Id:           b.Id,
			Name:         b.Name,
			BuildingType: b.BuildingType,
		})
	}
	return responses
}

func buildExcelFromMappingBuildings(data []webBuilding.MappingBuildingResponse) ([]byte, error) {
	f := excelize.NewFile()
	sheetName := "Buildings"
	const sheet = "Sheet1"
	headers := []string{"Building ID", "Name", "Building Type", "Grade", "Completion Year", "Subdistrict", "City", "Province", "Address", "Status", "Sellable", "Connectivity", "Latitude", "Longitude", "LCD Presence"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = f.SetCellValue(sheet, cell, h)
	}
	for row, b := range data {
		rowIdx := row + 2
		_ = f.SetCellValue(sheet, mustCell(1, rowIdx), b.ExternalBuildingId)
		_ = f.SetCellValue(sheet, mustCell(2, rowIdx), b.Name)
		_ = f.SetCellValue(sheet, mustCell(3, rowIdx), b.BuildingType)
		_ = f.SetCellValue(sheet, mustCell(4, rowIdx), b.GradeResource)
		_ = f.SetCellValue(sheet, mustCell(5, rowIdx), b.CompletionYear)
		_ = f.SetCellValue(sheet, mustCell(6, rowIdx), b.Subdistrict)
		_ = f.SetCellValue(sheet, mustCell(7, rowIdx), b.Citytown)
		_ = f.SetCellValue(sheet, mustCell(8, rowIdx), b.Province)
		_ = f.SetCellValue(sheet, mustCell(9, rowIdx), b.Address)
		_ = f.SetCellValue(sheet, mustCell(10, rowIdx), b.BuildingStatus)
		_ = f.SetCellValue(sheet, mustCell(11, rowIdx), b.Sellable)
		_ = f.SetCellValue(sheet, mustCell(12, rowIdx), b.Connectivity)
		_ = f.SetCellValue(sheet, mustCell(13, rowIdx), b.Latitude)
		_ = f.SetCellValue(sheet, mustCell(14, rowIdx), b.Longitude)
		_ = f.SetCellValue(sheet, mustCell(15, rowIdx), b.LcdPresenceStatus)
	}
	if sheetName != sheet {
		_ = f.SetSheetName(sheet, sheetName)
	}
	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
