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
	RepositoryCategoryInterface    repositoriesCategory.RepositoryCategoryInterface
	RepositorySubCategoryInterface repositoriesSubCategory.RepositorySubCategoryInterface
	RepositoryMotherBrandInterface repositoriesMotherBrand.RepositoryMotherBrandInterface
	RepositoryBranchInterface      repositoriesBranch.RepositoryBranchInterface
}

func NewServicePOIImpl(
	db *sql.DB,
	repositoryPOI repositoriesPOI.RepositoryPOIInterface,
	repoCategory repositoriesCategory.RepositoryCategoryInterface,
	repoSubCategory repositoriesSubCategory.RepositorySubCategoryInterface,
	repoMotherBrand repositoriesMotherBrand.RepositoryMotherBrandInterface,
	repoBranch repositoriesBranch.RepositoryBranchInterface,
) ServicePOIInterface {
	return &ServicePOIImpl{
		DB:                             db,
		RepositoryPOIInterface:         repositoryPOI,
		RepositoryCategoryInterface:    repoCategory,
		RepositorySubCategoryInterface: repoSubCategory,
		RepositoryMotherBrandInterface: repoMotherBrand,
		RepositoryBranchInterface:      repoBranch,
	}
}

// Create creates a new POI together with its owned points.
func (service *ServicePOIImpl) Create(ctx context.Context, request webPOI.CreatePOIRequest) webPOI.POIResponse {
	tx, err := service.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	service.validateMetadata(ctx, tx, request.CategoryId, request.SubCategoryId, request.MotherBrandId)
	service.validateBranches(ctx, tx, request.Points)

	poi := models.POI{
		Brand:         request.Brand,
		Color:         request.Color,
		CategoryId:    request.CategoryId,
		SubCategoryId: request.SubCategoryId,
		MotherBrandId: request.MotherBrandId,
	}

	createdPOI, err := service.RepositoryPOIInterface.Create(ctx, tx, poi, pointsFromInputs(request.Points))
	helpers.PanicIfError(err)

	return service.poiModelToResponse(createdPOI)
}

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

// Update updates a POI and replaces its owned points.
func (service *ServicePOIImpl) Update(ctx context.Context, request webPOI.UpdatePOIRequest, id int) webPOI.POIResponse {
	tx, err := service.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	existingPOI, err := service.RepositoryPOIInterface.FindById(ctx, tx, id)
	if err == sql.ErrNoRows {
		panic(exceptions.NewNotFoundError("POI not found"))
	}
	helpers.PanicIfError(err)

	service.validateMetadata(ctx, tx, request.CategoryId, request.SubCategoryId, request.MotherBrandId)
	service.validateBranches(ctx, tx, request.Points)

	existingPOI.Brand = request.Brand
	existingPOI.Color = request.Color
	existingPOI.CategoryId = request.CategoryId
	existingPOI.SubCategoryId = request.SubCategoryId
	existingPOI.MotherBrandId = request.MotherBrandId

	updatedPOI, err := service.RepositoryPOIInterface.Update(ctx, tx, existingPOI, pointsFromInputs(request.Points))
	helpers.PanicIfError(err)

	return service.poiModelToResponse(updatedPOI)
}

// Delete cascades to owned points via the FK on poi_points.poi_id.
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

// Import parses xlsx/csv. Each row is a point; rows are grouped by Brand. The first
// row of each brand sets the POI metadata (category/sub_category/mother_brand). If
// a later row of the same brand disagrees, the whole file is rejected.
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

	header := rows[0]
	colMap := mapHeaderColumns(header)

	for _, col := range []string{"brand", "coordinate"} {
		if _, exists := colMap[col]; !exists {
			panic(exceptions.NewBadRequestError(fmt.Sprintf("Missing required column: %s", col)))
		}
	}

	type brandGroup struct {
		brand           string
		categoryName    string
		subCategoryName string
		motherBrandName string
		firstRow        int
		rows            []int // every Excel row that belongs to this brand group
		points          []models.POIPoint
		seenAt          map[string]int // poi_name|address -> first Excel row
	}
	type duplicateEntry struct {
		Brand   string `json:"brand"`
		POIName string `json:"poi_name"`
		Address string `json:"address"`
		Rows    []int  `json:"rows"`
	}
	type metadataMismatch struct {
		Brand string `json:"brand"`
		Field string `json:"field"`
		Rows  []int  `json:"rows"`
	}

	brandOrder := []string{}
	groups := map[string]*brandGroup{}
	duplicateIndex := map[string]*duplicateEntry{}
	// mismatchedFields[brand] is the set of fields ("category", "sub_category",
	// "mother_brand") where any row in that brand disagreed with another.
	mismatchedFields := map[string]map[string]struct{}{}

	noteMismatch := func(brand, field string) {
		if _, ok := mismatchedFields[brand]; !ok {
			mismatchedFields[brand] = map[string]struct{}{}
		}
		mismatchedFields[brand][field] = struct{}{}
	}

	var lastBrand string
	for i, row := range rows[1:] {
		excelRow := i + 2

		brandVal := getColValue(row, colMap, "brand")
		if brandVal == "" {
			// Inherit brand from the previous non-empty row (handles merged
			// cells and spreadsheets where brand is filled once per group).
			// A fully blank row is still skipped.
			if lastBrand == "" || isRowBlank(row) {
				continue
			}
			brandVal = lastBrand
		} else {
			lastBrand = brandVal
		}

		categoryName := getColValue(row, colMap, "category")
		subCategoryName := getColValue(row, colMap, "sub_category")
		motherBrandName := getColValue(row, colMap, "mother_brand")
		branchName := getColValue(row, colMap, "branch")
		poiName := getColValue(row, colMap, "poi_name")
		address := getColValue(row, colMap, "address")
		lat, lng := parseCoordinate(getColValue(row, colMap, "coordinate"))

		group, exists := groups[brandVal]
		if !exists {
			group = &brandGroup{
				brand:           brandVal,
				categoryName:    categoryName,
				subCategoryName: subCategoryName,
				motherBrandName: motherBrandName,
				firstRow:        excelRow,
				seenAt:          map[string]int{},
			}
			groups[brandVal] = group
			brandOrder = append(brandOrder, brandVal)
		} else {
			if !strings.EqualFold(group.categoryName, categoryName) {
				noteMismatch(brandVal, "category")
			}
			if !strings.EqualFold(group.subCategoryName, subCategoryName) {
				noteMismatch(brandVal, "sub_category")
			}
			if !strings.EqualFold(group.motherBrandName, motherBrandName) {
				noteMismatch(brandVal, "mother_brand")
			}
		}
		group.rows = append(group.rows, excelRow)

		dupKey := strings.ToLower(poiName) + "|" + strings.ToLower(address)
		if firstRow, dup := group.seenAt[dupKey]; dup {
			key := fmt.Sprintf("%s|%s", brandVal, dupKey)
			if entry, ok := duplicateIndex[key]; ok {
				entry.Rows = append(entry.Rows, excelRow)
			} else {
				duplicateIndex[key] = &duplicateEntry{
					Brand:   brandVal,
					POIName: poiName,
					Address: address,
					Rows:    []int{firstRow, excelRow},
				}
			}
			continue
		}
		group.seenAt[dupKey] = excelRow

		branchId := service.findOrCreateBranch(ctx, tx, branchName)
		group.points = append(group.points, models.POIPoint{
			POIName:   poiName,
			Address:   address,
			Latitude:  lat,
			Longitude: lng,
			BranchId:  branchId,
		})
	}

	if len(mismatchedFields) > 0 {
		mismatches := make([]metadataMismatch, 0)
		for _, brand := range brandOrder {
			fields, ok := mismatchedFields[brand]
			if !ok {
				continue
			}
			group := groups[brand]
			for _, field := range []string{"category", "sub_category", "mother_brand"} {
				if _, hit := fields[field]; !hit {
					continue
				}
				rowsCopy := append([]int(nil), group.rows...)
				mismatches = append(mismatches, metadataMismatch{
					Brand: brand,
					Field: field,
					Rows:  rowsCopy,
				})
			}
		}
		panic(exceptions.NewBadRequestWithExtras(
			"Metadata mismatch within a brand group: all rows of the same brand must share the same Category, Sub-Category, and Mother Brand. Please review the rows below.",
			map[string]interface{}{"mismatches": mismatches},
		))
	}

	if len(duplicateIndex) > 0 {
		duplicates := make([]duplicateEntry, 0, len(duplicateIndex))
		for _, entry := range duplicateIndex {
			duplicates = append(duplicates, *entry)
		}
		panic(exceptions.NewBadRequestWithExtras(
			"Duplicate rows found: the same brand cannot reference the same POI (by name and address) more than once.",
			map[string]interface{}{"duplicates": duplicates},
		))
	}

	// Replace: delete any existing POIs whose brand matches the imported brands.
	existing, err := service.RepositoryPOIInterface.FindByBrands(ctx, tx, brandOrder)
	helpers.PanicIfError(err)
	for _, existingPOI := range existing {
		err = service.RepositoryPOIInterface.Delete(ctx, tx, existingPOI.Id)
		helpers.PanicIfError(err)
	}

	var responses []webPOI.POIResponse
	for i, brandKey := range brandOrder {
		group := groups[brandKey]
		color := colorPalette[i%len(colorPalette)]

		categoryId := service.findOrCreateCategory(ctx, tx, group.categoryName)
		subCategoryId := service.findOrCreateSubCategory(ctx, tx, group.subCategoryName)
		motherBrandId := service.findOrCreateMotherBrand(ctx, tx, group.motherBrandName)

		poi := models.POI{
			Brand:         group.brand,
			Color:         color,
			CategoryId:    categoryId,
			SubCategoryId: subCategoryId,
			MotherBrandId: motherBrandId,
		}

		createdPOI, err := service.RepositoryPOIInterface.Create(ctx, tx, poi, group.points)
		helpers.PanicIfError(err)
		responses = append(responses, service.poiModelToResponse(createdPOI))
	}

	return responses
}

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

// validateMetadata ensures provided category/sub/mother-brand IDs exist.
func (service *ServicePOIImpl) validateMetadata(ctx context.Context, tx *sql.Tx, categoryId, subCategoryId, motherBrandId *int) {
	if categoryId != nil {
		if _, err := service.RepositoryCategoryInterface.FindById(ctx, tx, *categoryId); err != nil {
			if err == sql.ErrNoRows {
				panic(exceptions.NewBadRequest("Category not found"))
			}
			helpers.PanicIfError(err)
		}
	}
	if subCategoryId != nil {
		if _, err := service.RepositorySubCategoryInterface.FindById(ctx, tx, *subCategoryId); err != nil {
			if err == sql.ErrNoRows {
				panic(exceptions.NewBadRequest("Sub-Category not found"))
			}
			helpers.PanicIfError(err)
		}
	}
	if motherBrandId != nil {
		if _, err := service.RepositoryMotherBrandInterface.FindById(ctx, tx, *motherBrandId); err != nil {
			if err == sql.ErrNoRows {
				panic(exceptions.NewBadRequest("Mother Brand not found"))
			}
			helpers.PanicIfError(err)
		}
	}
}

// validateBranches ensures provided branch IDs on each point exist.
func (service *ServicePOIImpl) validateBranches(ctx context.Context, tx *sql.Tx, points []webPOI.POIPointInput) {
	for _, pt := range points {
		if pt.BranchId == nil {
			continue
		}
		if _, err := service.RepositoryBranchInterface.FindById(ctx, tx, *pt.BranchId); err != nil {
			if err == sql.ErrNoRows {
				panic(exceptions.NewBadRequest("Branch not found"))
			}
			helpers.PanicIfError(err)
		}
	}
}

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

func (service *ServicePOIImpl) poiModelToResponse(poi models.POI) webPOI.POIResponse {
	points := make([]webPOI.POIPointResponse, len(poi.Points))
	for i, point := range poi.Points {
		points[i] = webPOI.POIPointResponse{
			Id:        point.Id,
			POIName:   point.POIName,
			Address:   point.Address,
			Latitude:  point.Latitude,
			Longitude: point.Longitude,
			Branch:    point.BranchName,
			BranchId:  point.BranchId,
			CreatedAt: point.CreatedAt,
			UpdatedAt: point.UpdatedAt,
		}
	}

	return webPOI.POIResponse{
		Id:            poi.Id,
		Brand:         poi.Brand,
		Color:         poi.Color,
		Category:      poi.CategoryName,
		SubCategory:   poi.SubCategoryName,
		MotherBrand:   poi.MotherBrandName,
		CategoryId:    poi.CategoryId,
		SubCategoryId: poi.SubCategoryId,
		MotherBrandId: poi.MotherBrandId,
		Points:        points,
		CreatedAt:     poi.CreatedAt,
		UpdatedAt:     poi.UpdatedAt,
	}
}

func pointsFromInputs(inputs []webPOI.POIPointInput) []models.POIPoint {
	out := make([]models.POIPoint, len(inputs))
	for i, in := range inputs {
		out[i] = models.POIPoint{
			POIName:   in.POIName,
			Address:   in.Address,
			Latitude:  in.Latitude,
			Longitude: in.Longitude,
			BranchId:  in.BranchId,
		}
	}
	return out
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

// isRowBlank reports whether every cell in the row is empty after trimming.
func isRowBlank(row []string) bool {
	for _, cell := range row {
		if strings.TrimSpace(cell) != "" {
			return false
		}
	}
	return true
}

func getColValue(row []string, colMap map[string]int, key string) string {
	idx, exists := colMap[key]
	if !exists || idx >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[idx])
}

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

			_ = f.SetCellValue(sheet, mustCell(1, rowIdx), poi.CategoryName)
			_ = f.SetCellValue(sheet, mustCell(2, rowIdx), poi.SubCategoryName)
			_ = f.SetCellValue(sheet, mustCell(3, rowIdx), poi.MotherBrandName)
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
