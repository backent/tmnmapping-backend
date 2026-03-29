package poi

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/malikabdulaziz/tmn-backend/exceptions"
	"github.com/malikabdulaziz/tmn-backend/helpers"
	"github.com/malikabdulaziz/tmn-backend/models"
	repositoriesBranch "github.com/malikabdulaziz/tmn-backend/repositories/branch"
	repositoriesCategory "github.com/malikabdulaziz/tmn-backend/repositories/category"
	repositoriesMotherBrand "github.com/malikabdulaziz/tmn-backend/repositories/motherbrand"
	repositoriesPOI "github.com/malikabdulaziz/tmn-backend/repositories/poi"
	repositoriesPOIPoint "github.com/malikabdulaziz/tmn-backend/repositories/poipoint"
	repositoriesSubCategory "github.com/malikabdulaziz/tmn-backend/repositories/subcategory"
	webPOI "github.com/malikabdulaziz/tmn-backend/web/poi"
	"github.com/xuri/excelize/v2"
)

// Color palette for auto-assigning colors during import
var colorPalette = []string{
	"#1976D2", "#424242", "#FF6F00", "#E91E63",
	"#388E3C", "#C2185B", "#7B1FA2", "#0097A7",
	"#0288D1", "#00796B", "#F57C00", "#D32F2F",
	"#5D4037", "#455A64", "#303F9F", "#C62828",
}

type ServicePOIImpl struct {
	DB                             *sql.DB
	RepositoryPOIInterface         repositoriesPOI.RepositoryPOIInterface
	RepositoryPOIPointInterface    repositoriesPOIPoint.RepositoryPOIPointInterface
	RepositoryCategoryInterface    repositoriesCategory.RepositoryCategoryInterface
	RepositorySubCategoryInterface repositoriesSubCategory.RepositorySubCategoryInterface
	RepositoryMotherBrandInterface repositoriesMotherBrand.RepositoryMotherBrandInterface
	RepositoryBranchInterface      repositoriesBranch.RepositoryBranchInterface
}

func NewServicePOIImpl(
	db *sql.DB,
	repositoryPOI repositoriesPOI.RepositoryPOIInterface,
	repositoryPOIPoint repositoriesPOIPoint.RepositoryPOIPointInterface,
	repoCategory repositoriesCategory.RepositoryCategoryInterface,
	repoSubCategory repositoriesSubCategory.RepositorySubCategoryInterface,
	repoMotherBrand repositoriesMotherBrand.RepositoryMotherBrandInterface,
	repoBranch repositoriesBranch.RepositoryBranchInterface,
) ServicePOIInterface {
	return &ServicePOIImpl{
		DB:                             db,
		RepositoryPOIInterface:         repositoryPOI,
		RepositoryPOIPointInterface:    repositoryPOIPoint,
		RepositoryCategoryInterface:    repoCategory,
		RepositorySubCategoryInterface: repoSubCategory,
		RepositoryMotherBrandInterface: repoMotherBrand,
		RepositoryBranchInterface:      repoBranch,
	}
}

// Create creates a new POI with links to existing points
func (service *ServicePOIImpl) Create(ctx context.Context, request webPOI.CreatePOIRequest) webPOI.POIResponse {
	tx, err := service.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	// Validate that all point IDs exist
	service.validatePointIds(ctx, tx, request.PointIds)

	poi := models.POI{
		Brand: request.Brand,
		Color: request.Color,
	}

	createdPOI, err := service.RepositoryPOIInterface.Create(ctx, tx, poi, request.PointIds)
	helpers.PanicIfError(err)

	return service.poiModelToResponse(createdPOI)
}

// FindAll retrieves all POIs with their points, with pagination
func (service *ServicePOIImpl) FindAll(ctx context.Context, request webPOI.POIRequestFindAll) ([]webPOI.POIResponse, int) {
	tx, err := service.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	search := request.GetSearch()

	pois, err := service.RepositoryPOIInterface.FindAll(ctx, tx, request.GetTake(), request.GetSkip(), request.GetOrderBy(), request.GetOrderDirection(), search)
	helpers.PanicIfError(err)

	total, err := service.RepositoryPOIInterface.CountAll(ctx, tx, search)
	helpers.PanicIfError(err)

	responses := make([]webPOI.POIResponse, len(pois))
	for i, poi := range pois {
		responses[i] = service.poiModelToResponse(poi)
	}

	return responses, total
}

// FindById retrieves a POI by ID with its points
func (service *ServicePOIImpl) FindById(ctx context.Context, id int) webPOI.POIResponse {
	tx, err := service.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	poi, err := service.RepositoryPOIInterface.FindById(ctx, tx, id)
	if err == sql.ErrNoRows {
		panic(exceptions.NewNotFoundError("POI not found"))
	}
	helpers.PanicIfError(err)

	return service.poiModelToResponse(poi)
}

// Update updates a POI and replaces its point links
func (service *ServicePOIImpl) Update(ctx context.Context, request webPOI.UpdatePOIRequest, id int) webPOI.POIResponse {
	tx, err := service.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	existingPOI, err := service.RepositoryPOIInterface.FindById(ctx, tx, id)
	if err == sql.ErrNoRows {
		panic(exceptions.NewNotFoundError("POI not found"))
	}
	helpers.PanicIfError(err)

	// Validate that all point IDs exist
	service.validatePointIds(ctx, tx, request.PointIds)

	existingPOI.Brand = request.Brand
	existingPOI.Color = request.Color

	updatedPOI, err := service.RepositoryPOIInterface.Update(ctx, tx, existingPOI, request.PointIds)
	helpers.PanicIfError(err)

	return service.poiModelToResponse(updatedPOI)
}

// Delete deletes a POI (cascade removes junction links, NOT the points themselves)
func (service *ServicePOIImpl) Delete(ctx context.Context, id int) {
	tx, err := service.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	_, err = service.RepositoryPOIInterface.FindById(ctx, tx, id)
	if err == sql.ErrNoRows {
		panic(exceptions.NewNotFoundError("POI not found"))
	}
	helpers.PanicIfError(err)

	err = service.RepositoryPOIInterface.Delete(ctx, tx, id)
	helpers.PanicIfError(err)
}

// Import parses an xlsx or csv file and creates POIs grouped by brand, creating points as needed
func (service *ServicePOIImpl) Import(ctx context.Context, fileBytes []byte, fileType string) []webPOI.POIResponse {
	tx, err := service.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	var rows [][]string

	switch strings.ToLower(fileType) {
	case "xlsx":
		rows, err = parseXLSX(fileBytes)
		helpers.PanicIfError(err)
	case "csv":
		rows, err = parseCSV(fileBytes)
		helpers.PanicIfError(err)
	default:
		panic(exceptions.NewBadRequestError("Unsupported file type. Use xlsx or csv."))
	}

	if len(rows) < 2 {
		panic(exceptions.NewBadRequestError("File must contain a header row and at least one data row."))
	}

	// Parse header to find column indices
	header := rows[0]
	colMap := mapHeaderColumns(header)

	// Validate required columns
	requiredCols := []string{"brand", "coordinate"}
	for _, col := range requiredCols {
		if _, exists := colMap[col]; !exists {
			panic(exceptions.NewBadRequestError(fmt.Sprintf("Missing required column: %s", col)))
		}
	}

	// Group rows by Brand, creating standalone points for each row
	type brandGroup struct {
		brand    string
		pointIds []int
	}
	brandOrder := []string{}
	groups := map[string]*brandGroup{}

	for _, row := range rows[1:] {
		brandVal := getColValue(row, colMap, "brand")
		if brandVal == "" {
			continue
		}

		poiName := getColValue(row, colMap, "poi_name")
		address := getColValue(row, colMap, "address")

		// Check if a point with the same name and address already exists
		var pointId int
		existing, err := service.RepositoryPOIPointInterface.FindByNameAndAddress(ctx, tx, poiName, address)
		if err == nil {
			pointId = existing.Id
		} else if err == sql.ErrNoRows {
			// Resolve metadata IDs via find-or-create
			categoryId := service.findOrCreateCategory(ctx, tx, getColValue(row, colMap, "category"))
			subCategoryId := service.findOrCreateSubCategory(ctx, tx, getColValue(row, colMap, "sub_category"))
			motherBrandId := service.findOrCreateMotherBrand(ctx, tx, getColValue(row, colMap, "mother_brand"))
			branchId := service.findOrCreateBranch(ctx, tx, getColValue(row, colMap, "branch"))

			// Create a new standalone point
			lat, lng := parseCoordinate(getColValue(row, colMap, "coordinate"))
			point := models.POIPoint{
				POIName:       poiName,
				Address:       address,
				Latitude:      lat,
				Longitude:     lng,
				CategoryId:    categoryId,
				SubCategoryId: subCategoryId,
				MotherBrandId: motherBrandId,
				BranchId:      branchId,
			}
			createdPoint, createErr := service.RepositoryPOIPointInterface.Create(ctx, tx, point)
			helpers.PanicIfError(createErr)
			pointId = createdPoint.Id
		} else {
			helpers.PanicIfError(err)
		}

		if _, exists := groups[brandVal]; !exists {
			groups[brandVal] = &brandGroup{brand: brandVal}
			brandOrder = append(brandOrder, brandVal)
		}
		groups[brandVal].pointIds = append(groups[brandVal].pointIds, pointId)
	}

	// Replace: delete any existing POIs whose brand matches the imported brands
	existing, err := service.RepositoryPOIInterface.FindByBrands(ctx, tx, brandOrder)
	helpers.PanicIfError(err)
	for _, existingPOI := range existing {
		err = service.RepositoryPOIInterface.DeletePointLinksByPOIId(ctx, tx, existingPOI.Id)
		helpers.PanicIfError(err)
		err = service.RepositoryPOIInterface.Delete(ctx, tx, existingPOI.Id)
		helpers.PanicIfError(err)
	}

	// Create fresh POIs from the imported data
	var responses []webPOI.POIResponse
	for i, brandKey := range brandOrder {
		group := groups[brandKey]
		color := colorPalette[i%len(colorPalette)]

		poi := models.POI{
			Brand: group.brand,
			Color: color,
		}

		createdPOI, err := service.RepositoryPOIInterface.Create(ctx, tx, poi, group.pointIds)
		helpers.PanicIfError(err)

		responses = append(responses, service.poiModelToResponse(createdPOI))
	}

	return responses
}

// Export generates an xlsx file with all POIs flattened
func (service *ServicePOIImpl) Export(ctx context.Context, search string) ([]byte, error) {
	tx, err := service.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer helpers.CommitOrRollback(tx)

	pois, err := service.RepositoryPOIInterface.FindAllFlat(ctx, tx, search)
	if err != nil {
		return nil, err
	}

	return buildPOIExcel(pois)
}

// validatePointIds ensures all point IDs exist
func (service *ServicePOIImpl) validatePointIds(ctx context.Context, tx *sql.Tx, pointIds []int) {
	for _, pid := range pointIds {
		_, err := service.RepositoryPOIPointInterface.FindById(ctx, tx, pid)
		if err == sql.ErrNoRows {
			panic(exceptions.NewBadRequest("POI point not found"))
		}
		helpers.PanicIfError(err)
	}
}

// findOrCreateCategory resolves a category name to an ID (find or create)
func (service *ServicePOIImpl) findOrCreateCategory(ctx context.Context, tx *sql.Tx, name string) *int {
	if name == "" {
		return nil
	}
	cat, err := service.RepositoryCategoryInterface.FindByName(ctx, tx, name)
	if err == sql.ErrNoRows {
		cat, err = service.RepositoryCategoryInterface.Create(ctx, tx, models.Category{Name: name})
		helpers.PanicIfError(err)
	} else {
		helpers.PanicIfError(err)
	}
	id := cat.Id
	return &id
}

// findOrCreateSubCategory resolves a sub_category name to an ID (find or create)
func (service *ServicePOIImpl) findOrCreateSubCategory(ctx context.Context, tx *sql.Tx, name string) *int {
	if name == "" {
		return nil
	}
	sc, err := service.RepositorySubCategoryInterface.FindByName(ctx, tx, name)
	if err == sql.ErrNoRows {
		sc, err = service.RepositorySubCategoryInterface.Create(ctx, tx, models.SubCategory{Name: name})
		helpers.PanicIfError(err)
	} else {
		helpers.PanicIfError(err)
	}
	id := sc.Id
	return &id
}

// findOrCreateMotherBrand resolves a mother_brand name to an ID (find or create)
func (service *ServicePOIImpl) findOrCreateMotherBrand(ctx context.Context, tx *sql.Tx, name string) *int {
	if name == "" {
		return nil
	}
	mb, err := service.RepositoryMotherBrandInterface.FindByName(ctx, tx, name)
	if err == sql.ErrNoRows {
		mb, err = service.RepositoryMotherBrandInterface.Create(ctx, tx, models.MotherBrand{Name: name})
		helpers.PanicIfError(err)
	} else {
		helpers.PanicIfError(err)
	}
	id := mb.Id
	return &id
}

// findOrCreateBranch resolves a branch name to an ID (find or create)
func (service *ServicePOIImpl) findOrCreateBranch(ctx context.Context, tx *sql.Tx, name string) *int {
	if name == "" {
		return nil
	}
	br, err := service.RepositoryBranchInterface.FindByName(ctx, tx, name)
	if err == sql.ErrNoRows {
		br, err = service.RepositoryBranchInterface.Create(ctx, tx, models.Branch{Name: name})
		helpers.PanicIfError(err)
	} else {
		helpers.PanicIfError(err)
	}
	id := br.Id
	return &id
}

// Helper function to convert model to response
func (service *ServicePOIImpl) poiModelToResponse(poi models.POI) webPOI.POIResponse {
	points := make([]webPOI.POIPointResponse, len(poi.Points))
	for i, point := range poi.Points {
		points[i] = webPOI.POIPointResponse{
			Id:          point.Id,
			POIName:     point.POIName,
			Address:     point.Address,
			Latitude:    point.Latitude,
			Longitude:   point.Longitude,
			Category:    point.CategoryName,
			SubCategory: point.SubCategoryName,
			MotherBrand: point.MotherBrandName,
			Branch:      point.BranchName,
			CreatedAt:   point.CreatedAt,
			UpdatedAt:   point.UpdatedAt,
		}
	}

	return webPOI.POIResponse{
		Id:        poi.Id,
		Brand:     poi.Brand,
		Color:     poi.Color,
		Points:    points,
		CreatedAt: poi.CreatedAt,
		UpdatedAt: poi.UpdatedAt,
	}
}

// --- Import helpers ---

func parseXLSX(fileBytes []byte) ([][]string, error) {
	f, err := excelize.OpenReader(bytes.NewReader(fileBytes))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	sheetName := f.GetSheetName(0)
	return f.GetRows(sheetName)
}

func parseCSV(fileBytes []byte) ([][]string, error) {
	reader := csv.NewReader(bytes.NewReader(fileBytes))
	var rows [][]string
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		rows = append(rows, record)
	}
	return rows, nil
}

// mapHeaderColumns maps normalized column names to their indices
func mapHeaderColumns(header []string) map[string]int {
	colMap := make(map[string]int)
	for i, h := range header {
		normalized := strings.ToLower(strings.TrimSpace(h))
		normalized = strings.ReplaceAll(normalized, "-", "_")
		normalized = strings.ReplaceAll(normalized, " ", "_")

		switch normalized {
		case "category":
			colMap["category"] = i
		case "sub_category", "subcategory":
			colMap["sub_category"] = i
		case "mother_brand", "motherbrand":
			colMap["mother_brand"] = i
		case "brand":
			colMap["brand"] = i
		case "branch":
			colMap["branch"] = i
		case "poi_name", "poiname":
			colMap["poi_name"] = i
		case "address":
			colMap["address"] = i
		case "coordinate", "coordinates":
			colMap["coordinate"] = i
		}
	}
	return colMap
}

func getColValue(row []string, colMap map[string]int, key string) string {
	idx, exists := colMap[key]
	if !exists || idx >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[idx])
}

// parseCoordinate parses a coordinate string like "-6.226203670181947, 106.79693887621839"
func parseCoordinate(coord string) (float64, float64) {
	if coord == "" {
		return 0, 0
	}

	parts := strings.SplitN(coord, ",", 2)
	if len(parts) != 2 {
		return 0, 0
	}

	lat, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	if err != nil {
		return 0, 0
	}

	lng, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	if err != nil {
		return 0, 0
	}

	return lat, lng
}

// --- Export helpers ---

func mustCell(col, row int) string {
	s, _ := excelize.CoordinatesToCellName(col, row)
	return s
}

func buildPOIExcel(pois []models.POI) ([]byte, error) {
	f := excelize.NewFile()
	const sheet = "Sheet1"
	sheetName := "POI Data"

	headers := []string{"Category", "Sub-Category", "Mother Brand", "Brand", "Branch", "POI Name", "Address", "Coordinate"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = f.SetCellValue(sheet, cell, h)
	}

	rowIdx := 2
	for _, poi := range pois {
		for _, point := range poi.Points {
			coordinate := ""
			if point.Latitude != 0 || point.Longitude != 0 {
				coordinate = fmt.Sprintf("%f, %f", point.Latitude, point.Longitude)
			}

			_ = f.SetCellValue(sheet, mustCell(1, rowIdx), point.CategoryName)
			_ = f.SetCellValue(sheet, mustCell(2, rowIdx), point.SubCategoryName)
			_ = f.SetCellValue(sheet, mustCell(3, rowIdx), point.MotherBrandName)
			_ = f.SetCellValue(sheet, mustCell(4, rowIdx), poi.Brand)
			_ = f.SetCellValue(sheet, mustCell(5, rowIdx), point.BranchName)
			_ = f.SetCellValue(sheet, mustCell(6, rowIdx), point.POIName)
			_ = f.SetCellValue(sheet, mustCell(7, rowIdx), point.Address)
			_ = f.SetCellValue(sheet, mustCell(8, rowIdx), coordinate)
			rowIdx++
		}
	}

	_ = f.SetSheetName(sheet, sheetName)

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
