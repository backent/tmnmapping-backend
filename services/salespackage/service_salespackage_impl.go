package salespackage

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"strings"

	"github.com/malikabdulaziz/tmn-backend/exceptions"
	"github.com/malikabdulaziz/tmn-backend/helpers"
	"github.com/malikabdulaziz/tmn-backend/models"
	repositoriesBuilding "github.com/malikabdulaziz/tmn-backend/repositories/building"
	repositoriesSalesPackage "github.com/malikabdulaziz/tmn-backend/repositories/salespackage"
	webSalesPackage "github.com/malikabdulaziz/tmn-backend/web/salespackage"
	"github.com/xuri/excelize/v2"
)

type ServiceSalesPackageImpl struct {
	DB                            *sql.DB
	RepositorySalesPackageInterface repositoriesSalesPackage.RepositorySalesPackageInterface
	RepositoryBuildingInterface     repositoriesBuilding.RepositoryBuildingInterface
}

func NewServiceSalesPackageImpl(
	db *sql.DB,
	repoSalesPackage repositoriesSalesPackage.RepositorySalesPackageInterface,
	repoBuilding repositoriesBuilding.RepositoryBuildingInterface,
) ServiceSalesPackageInterface {
	return &ServiceSalesPackageImpl{
		DB:                            db,
		RepositorySalesPackageInterface: repoSalesPackage,
		RepositoryBuildingInterface:     repoBuilding,
	}
}

// validateBuildingIdsErr ensures all building ids exist; panics with BadRequest if any invalid
func (s *ServiceSalesPackageImpl) validateBuildingIdsErr(ctx context.Context, tx *sql.Tx, buildingIds []int) {
	for _, bid := range buildingIds {
		_, err := s.RepositoryBuildingInterface.FindById(ctx, tx, bid)
		if err == sql.ErrNoRows {
			panic(exceptions.NewBadRequest("building not found"))
		}
		helpers.PanicIfError(err)
	}
}

// Create creates a new sales package with building links
func (s *ServiceSalesPackageImpl) Create(ctx context.Context, request webSalesPackage.CreateSalesPackageRequest) webSalesPackage.SalesPackageResponse {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	s.validateBuildingIdsErr(ctx, tx, request.BuildingIds)

	pkg := models.SalesPackage{Name: request.Name}
	created, err := s.RepositorySalesPackageInterface.Create(ctx, tx, pkg, request.BuildingIds)
	helpers.PanicIfError(err)
	return s.modelToResponse(created)
}

// FindAll retrieves all sales packages with pagination
func (s *ServiceSalesPackageImpl) FindAll(ctx context.Context, request webSalesPackage.SalesPackageRequestFindAll) ([]webSalesPackage.SalesPackageResponse, int) {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	list, err := s.RepositorySalesPackageInterface.FindAll(ctx, tx, request.GetTake(), request.GetSkip(), request.GetOrderBy(), request.GetOrderDirection())
	helpers.PanicIfError(err)
	total, err := s.RepositorySalesPackageInterface.CountAll(ctx, tx)
	helpers.PanicIfError(err)

	responses := make([]webSalesPackage.SalesPackageResponse, len(list))
	for i, p := range list {
		responses[i] = s.modelToResponse(p)
	}
	return responses, total
}

// FindById retrieves a sales package by ID
func (s *ServiceSalesPackageImpl) FindById(ctx context.Context, id int) webSalesPackage.SalesPackageResponse {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	pkg, err := s.RepositorySalesPackageInterface.FindById(ctx, tx, id)
	if err == sql.ErrNoRows {
		panic(exceptions.NewNotFoundError("sales package not found"))
	}
	helpers.PanicIfError(err)
	return s.modelToResponse(pkg)
}

// Update updates a sales package and replaces building links
func (s *ServiceSalesPackageImpl) Update(ctx context.Context, request webSalesPackage.UpdateSalesPackageRequest, id int) webSalesPackage.SalesPackageResponse {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	existing, err := s.RepositorySalesPackageInterface.FindById(ctx, tx, id)
	if err == sql.ErrNoRows {
		panic(exceptions.NewNotFoundError("sales package not found"))
	}
	helpers.PanicIfError(err)

	s.validateBuildingIdsErr(ctx, tx, request.BuildingIds)

	existing.Name = request.Name
	updated, err := s.RepositorySalesPackageInterface.Update(ctx, tx, existing, request.BuildingIds)
	helpers.PanicIfError(err)
	return s.modelToResponse(updated)
}

// Delete deletes a sales package
func (s *ServiceSalesPackageImpl) Delete(ctx context.Context, id int) {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	_, err = s.RepositorySalesPackageInterface.FindById(ctx, tx, id)
	if err == sql.ErrNoRows {
		panic(exceptions.NewNotFoundError("sales package not found"))
	}
	helpers.PanicIfError(err)
	err = s.RepositorySalesPackageInterface.Delete(ctx, tx, id)
	helpers.PanicIfError(err)
}

// Import parses an xlsx or csv file and creates/replaces sales packages
func (s *ServiceSalesPackageImpl) Import(ctx context.Context, fileBytes []byte, fileType string) []webSalesPackage.SalesPackageResponse {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	var rows [][]string
	switch strings.ToLower(fileType) {
	case "xlsx":
		rows, err = spParseXLSX(fileBytes)
		helpers.PanicIfError(err)
	case "csv":
		rows, err = spParseCSV(fileBytes)
		helpers.PanicIfError(err)
	default:
		panic(exceptions.NewBadRequestError("Unsupported file type. Use xlsx or csv."))
	}

	if len(rows) < 2 {
		panic(exceptions.NewBadRequestError("File must contain a header row and at least one data row."))
	}

	header := rows[0]
	colMap := spMapHeaderColumns(header)

	if _, exists := colMap["name"]; !exists {
		panic(exceptions.NewBadRequestError("Missing required column: name"))
	}
	if _, exists := colMap["building_name"]; !exists {
		panic(exceptions.NewBadRequestError("Missing required column: building_name"))
	}

	// Load all buildings for name->id resolution
	allBuildings, err := s.RepositoryBuildingInterface.FindAllDropdown(ctx, tx)
	helpers.PanicIfError(err)
	buildingNameMap := make(map[string]int)
	for _, b := range allBuildings {
		buildingNameMap[strings.TrimSpace(strings.ToLower(b.Name))] = b.Id
	}

	// Group rows by package name (each row has one building)
	type importGroup struct {
		name        string
		buildingIds []int
	}
	nameOrder := []string{}
	groups := map[string]*importGroup{}

	for _, row := range rows[1:] {
		name := spGetColValue(row, colMap, "name")
		if name == "" {
			continue
		}
		buildingName := strings.TrimSpace(strings.ToLower(spGetColValue(row, colMap, "building_name")))

		if _, exists := groups[name]; !exists {
			groups[name] = &importGroup{name: name}
			nameOrder = append(nameOrder, name)
		}

		if buildingName != "" {
			bid, ok := buildingNameMap[buildingName]
			if !ok {
				panic(exceptions.NewBadRequestError(fmt.Sprintf("Building not found: %s", buildingName)))
			}
			groups[name].buildingIds = append(groups[name].buildingIds, bid)
		}
	}

	// Delete existing sales packages with matching names
	existing, err := s.RepositorySalesPackageInterface.FindByNames(ctx, tx, nameOrder)
	helpers.PanicIfError(err)
	for _, ep := range existing {
		err = s.RepositorySalesPackageInterface.DeleteBuildingLinksBySalesPackageId(ctx, tx, ep.Id)
		helpers.PanicIfError(err)
		err = s.RepositorySalesPackageInterface.Delete(ctx, tx, ep.Id)
		helpers.PanicIfError(err)
	}

	// Create fresh sales packages
	var responses []webSalesPackage.SalesPackageResponse
	for _, name := range nameOrder {
		group := groups[name]
		pkg := models.SalesPackage{Name: group.name}
		created, err := s.RepositorySalesPackageInterface.Create(ctx, tx, pkg, group.buildingIds)
		helpers.PanicIfError(err)
		responses = append(responses, s.modelToResponse(created))
	}

	return responses
}

// Export generates an xlsx file with all sales packages
func (s *ServiceSalesPackageImpl) Export(ctx context.Context, search string) ([]byte, error) {
	tx, err := s.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer helpers.CommitOrRollback(tx)

	packages, err := s.RepositorySalesPackageInterface.FindAllFlat(ctx, tx, search)
	if err != nil {
		return nil, err
	}

	return buildSalesPackageExcel(packages)
}

func (s *ServiceSalesPackageImpl) modelToResponse(p models.SalesPackage) webSalesPackage.SalesPackageResponse {
	buildings := make([]webSalesPackage.BuildingRefResponse, len(p.Buildings))
	for i, b := range p.Buildings {
		buildings[i] = webSalesPackage.BuildingRefResponse{Id: b.Id, Name: b.Name}
	}
	return webSalesPackage.SalesPackageResponse{
		Id:        p.Id,
		Name:      p.Name,
		Buildings: buildings,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}

// --- Import helpers ---

func spParseXLSX(fileBytes []byte) ([][]string, error) {
	f, err := excelize.OpenReader(bytes.NewReader(fileBytes))
	if err != nil {
		return nil, err
	}
	defer f.Close()
	sheetName := f.GetSheetName(0)
	return f.GetRows(sheetName)
}

func spParseCSV(fileBytes []byte) ([][]string, error) {
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

func spMapHeaderColumns(header []string) map[string]int {
	colMap := make(map[string]int)
	for i, h := range header {
		normalized := strings.ToLower(strings.TrimSpace(h))
		normalized = strings.ReplaceAll(normalized, "-", "_")
		normalized = strings.ReplaceAll(normalized, " ", "_")
		switch normalized {
		case "name":
			colMap["name"] = i
		case "building_name", "buildingname", "building_names", "buildingnames", "buildings":
			colMap["building_name"] = i
		}
	}
	return colMap
}

func spGetColValue(row []string, colMap map[string]int, key string) string {
	idx, exists := colMap[key]
	if !exists || idx >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[idx])
}

// --- Export helpers ---

func spMustCell(col, row int) string {
	s, _ := excelize.CoordinatesToCellName(col, row)
	return s
}

func buildSalesPackageExcel(packages []models.SalesPackage) ([]byte, error) {
	f := excelize.NewFile()
	const sheet = "Sheet1"

	headers := []string{"Name", "Building Name"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = f.SetCellValue(sheet, cell, h)
	}

	rowIdx := 2
	for _, pkg := range packages {
		if len(pkg.Buildings) == 0 {
			_ = f.SetCellValue(sheet, spMustCell(1, rowIdx), pkg.Name)
			_ = f.SetCellValue(sheet, spMustCell(2, rowIdx), "")
			rowIdx++
		} else {
			for _, b := range pkg.Buildings {
				_ = f.SetCellValue(sheet, spMustCell(1, rowIdx), pkg.Name)
				_ = f.SetCellValue(sheet, spMustCell(2, rowIdx), b.Name)
				rowIdx++
			}
		}
	}

	_ = f.SetSheetName(sheet, "Sales Packages")

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
