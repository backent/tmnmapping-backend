package poipoint

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
	repositoriesPOIPoint "github.com/malikabdulaziz/tmn-backend/repositories/poipoint"
	webPOIPoint "github.com/malikabdulaziz/tmn-backend/web/poipoint"
	"github.com/xuri/excelize/v2"
)

type ServicePOIPointImpl struct {
	DB                          *sql.DB
	RepositoryPOIPointInterface repositoriesPOIPoint.RepositoryPOIPointInterface
}

func NewServicePOIPointImpl(
	db *sql.DB,
	repoPOIPoint repositoriesPOIPoint.RepositoryPOIPointInterface,
) ServicePOIPointInterface {
	return &ServicePOIPointImpl{
		DB:                          db,
		RepositoryPOIPointInterface: repoPOIPoint,
	}
}

// Create creates a new standalone POI point
func (s *ServicePOIPointImpl) Create(ctx context.Context, request webPOIPoint.CreatePOIPointRequest) webPOIPoint.POIPointResponse {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	point := models.POIPoint{
		POIName:     request.POIName,
		Address:     request.Address,
		Latitude:    request.Latitude,
		Longitude:   request.Longitude,
		Category:    request.Category,
		SubCategory: request.SubCategory,
		MotherBrand: request.MotherBrand,
		Branch:      request.Branch,
	}
	created, err := s.RepositoryPOIPointInterface.Create(ctx, tx, point)
	helpers.PanicIfError(err)
	return s.modelToResponse(created)
}

// FindAll retrieves all POI points with pagination
func (s *ServicePOIPointImpl) FindAll(ctx context.Context, request webPOIPoint.POIPointRequestFindAll) ([]webPOIPoint.POIPointResponse, int) {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	search := request.GetSearch()
	list, err := s.RepositoryPOIPointInterface.FindAll(ctx, tx, request.GetTake(), request.GetSkip(), request.GetOrderBy(), request.GetOrderDirection(), search)
	helpers.PanicIfError(err)
	total, err := s.RepositoryPOIPointInterface.CountAll(ctx, tx, search)
	helpers.PanicIfError(err)

	responses := make([]webPOIPoint.POIPointResponse, len(list))
	for i, p := range list {
		responses[i] = s.modelToResponse(p)
	}
	return responses, total
}

// FindById retrieves a POI point by ID
func (s *ServicePOIPointImpl) FindById(ctx context.Context, id int) webPOIPoint.POIPointResponse {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	point, err := s.RepositoryPOIPointInterface.FindById(ctx, tx, id)
	if err == sql.ErrNoRows {
		panic(exceptions.NewNotFoundError("POI point not found"))
	}
	helpers.PanicIfError(err)
	return s.modelToResponse(point)
}

// Update updates a POI point
func (s *ServicePOIPointImpl) Update(ctx context.Context, request webPOIPoint.UpdatePOIPointRequest, id int) webPOIPoint.POIPointResponse {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	existing, err := s.RepositoryPOIPointInterface.FindById(ctx, tx, id)
	if err == sql.ErrNoRows {
		panic(exceptions.NewNotFoundError("POI point not found"))
	}
	helpers.PanicIfError(err)

	existing.POIName = request.POIName
	existing.Address = request.Address
	existing.Latitude = request.Latitude
	existing.Longitude = request.Longitude
	existing.Category = request.Category
	existing.SubCategory = request.SubCategory
	existing.MotherBrand = request.MotherBrand
	existing.Branch = request.Branch

	updated, err := s.RepositoryPOIPointInterface.Update(ctx, tx, existing)
	helpers.PanicIfError(err)
	return s.modelToResponse(updated)
}

// Delete deletes a POI point
func (s *ServicePOIPointImpl) Delete(ctx context.Context, id int) {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	_, err = s.RepositoryPOIPointInterface.FindById(ctx, tx, id)
	if err == sql.ErrNoRows {
		panic(exceptions.NewNotFoundError("POI point not found"))
	}
	helpers.PanicIfError(err)
	err = s.RepositoryPOIPointInterface.Delete(ctx, tx, id)
	helpers.PanicIfError(err)
}

// GetPointUsage returns which POIs use a given point
func (s *ServicePOIPointImpl) GetPointUsage(ctx context.Context, id int) webPOIPoint.POIPointUsageResponse {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	_, err = s.RepositoryPOIPointInterface.FindById(ctx, tx, id)
	if err == sql.ErrNoRows {
		panic(exceptions.NewNotFoundError("POI point not found"))
	}
	helpers.PanicIfError(err)

	refs, err := s.RepositoryPOIPointInterface.FindPOIRefsByPointId(ctx, tx, id)
	helpers.PanicIfError(err)

	poiRefs := make([]webPOIPoint.POIRefResponse, len(refs))
	for i, r := range refs {
		poiRefs[i] = webPOIPoint.POIRefResponse{Id: r.Id, Brand: r.Brand}
	}
	return webPOIPoint.POIPointUsageResponse{POIs: poiRefs}
}

// Import parses an xlsx or csv file and creates POI points
func (s *ServicePOIPointImpl) Import(ctx context.Context, fileBytes []byte, fileType string) []webPOIPoint.POIPointResponse {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	var rows [][]string
	switch strings.ToLower(fileType) {
	case "xlsx":
		rows, err = ppParseXLSX(fileBytes)
		helpers.PanicIfError(err)
	case "csv":
		rows, err = ppParseCSV(fileBytes)
		helpers.PanicIfError(err)
	default:
		panic(exceptions.NewBadRequestError("Unsupported file type. Use xlsx or csv."))
	}

	if len(rows) < 2 {
		panic(exceptions.NewBadRequestError("File must contain a header row and at least one data row."))
	}

	header := rows[0]
	colMap := ppMapHeaderColumns(header)

	requiredCols := []string{"poi_name", "coordinate"}
	for _, col := range requiredCols {
		if _, exists := colMap[col]; !exists {
			panic(exceptions.NewBadRequestError(fmt.Sprintf("Missing required column: %s", col)))
		}
	}

	// Find or create points
	var responses []webPOIPoint.POIPointResponse
	for _, row := range rows[1:] {
		poiName := ppGetColValue(row, colMap, "poi_name")
		if poiName == "" {
			continue
		}

		address := ppGetColValue(row, colMap, "address")

		// Check if a point with the same name and address already exists
		existing, err := s.RepositoryPOIPointInterface.FindByNameAndAddress(ctx, tx, poiName, address)
		if err == nil {
			// Point already exists, reuse it
			responses = append(responses, s.modelToResponse(existing))
			continue
		}
		if err != sql.ErrNoRows {
			helpers.PanicIfError(err)
		}

		// Point does not exist, create a new one
		lat, lng := ppParseCoordinate(ppGetColValue(row, colMap, "coordinate"))

		point := models.POIPoint{
			POIName:     poiName,
			Address:     address,
			Latitude:    lat,
			Longitude:   lng,
			Category:    ppGetColValue(row, colMap, "category"),
			SubCategory: ppGetColValue(row, colMap, "sub_category"),
			MotherBrand: ppGetColValue(row, colMap, "mother_brand"),
			Branch:      ppGetColValue(row, colMap, "branch"),
		}

		created, err := s.RepositoryPOIPointInterface.Create(ctx, tx, point)
		helpers.PanicIfError(err)
		responses = append(responses, s.modelToResponse(created))
	}

	return responses
}

// Export generates an xlsx file with all POI points
func (s *ServicePOIPointImpl) Export(ctx context.Context, search string) ([]byte, error) {
	tx, err := s.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer helpers.CommitOrRollback(tx)

	points, err := s.RepositoryPOIPointInterface.FindAllFlat(ctx, tx, search)
	if err != nil {
		return nil, err
	}

	return buildPOIPointExcel(points)
}

func (s *ServicePOIPointImpl) modelToResponse(p models.POIPoint) webPOIPoint.POIPointResponse {
	pois := make([]webPOIPoint.POIRefResponse, len(p.POIs))
	for i, ref := range p.POIs {
		pois[i] = webPOIPoint.POIRefResponse{Id: ref.Id, Brand: ref.Brand}
	}
	return webPOIPoint.POIPointResponse{
		Id:          p.Id,
		POIName:     p.POIName,
		Address:     p.Address,
		Latitude:    p.Latitude,
		Longitude:   p.Longitude,
		Category:    p.Category,
		SubCategory: p.SubCategory,
		MotherBrand: p.MotherBrand,
		Branch:      p.Branch,
		POIs:        pois,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

// --- Import helpers ---

func ppParseXLSX(fileBytes []byte) ([][]string, error) {
	f, err := excelize.OpenReader(bytes.NewReader(fileBytes))
	if err != nil {
		return nil, err
	}
	defer f.Close()
	sheetName := f.GetSheetName(0)
	return f.GetRows(sheetName)
}

func ppParseCSV(fileBytes []byte) ([][]string, error) {
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

func ppMapHeaderColumns(header []string) map[string]int {
	colMap := make(map[string]int)
	for i, h := range header {
		normalized := strings.ToLower(strings.TrimSpace(h))
		normalized = strings.ReplaceAll(normalized, "-", "_")
		normalized = strings.ReplaceAll(normalized, " ", "_")
		switch normalized {
		case "poi_name", "poiname":
			colMap["poi_name"] = i
		case "address":
			colMap["address"] = i
		case "coordinate", "coordinates":
			colMap["coordinate"] = i
		case "category":
			colMap["category"] = i
		case "sub_category", "subcategory":
			colMap["sub_category"] = i
		case "mother_brand", "motherbrand":
			colMap["mother_brand"] = i
		case "branch":
			colMap["branch"] = i
		}
	}
	return colMap
}

func ppGetColValue(row []string, colMap map[string]int, key string) string {
	idx, exists := colMap[key]
	if !exists || idx >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[idx])
}

func ppParseCoordinate(coord string) (float64, float64) {
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

func ppMustCell(col, row int) string {
	s, _ := excelize.CoordinatesToCellName(col, row)
	return s
}

func buildPOIPointExcel(points []models.POIPoint) ([]byte, error) {
	f := excelize.NewFile()
	const sheet = "Sheet1"

	headers := []string{"POI Name", "Address", "Coordinate", "Category", "Sub-Category", "Mother Brand", "Branch", "Brands"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = f.SetCellValue(sheet, cell, h)
	}

	rowIdx := 2
	for _, point := range points {
		coordinate := ""
		if point.Latitude != 0 || point.Longitude != 0 {
			coordinate = fmt.Sprintf("%f, %f", point.Latitude, point.Longitude)
		}

		var brandNames []string
		for _, ref := range point.POIs {
			brandNames = append(brandNames, ref.Brand)
		}

		_ = f.SetCellValue(sheet, ppMustCell(1, rowIdx), point.POIName)
		_ = f.SetCellValue(sheet, ppMustCell(2, rowIdx), point.Address)
		_ = f.SetCellValue(sheet, ppMustCell(3, rowIdx), coordinate)
		_ = f.SetCellValue(sheet, ppMustCell(4, rowIdx), point.Category)
		_ = f.SetCellValue(sheet, ppMustCell(5, rowIdx), point.SubCategory)
		_ = f.SetCellValue(sheet, ppMustCell(6, rowIdx), point.MotherBrand)
		_ = f.SetCellValue(sheet, ppMustCell(7, rowIdx), point.Branch)
		_ = f.SetCellValue(sheet, ppMustCell(8, rowIdx), strings.Join(brandNames, ", "))
		rowIdx++
	}

	_ = f.SetSheetName(sheet, "POI Points")

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
