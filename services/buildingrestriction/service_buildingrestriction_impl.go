package buildingrestriction

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
	repositoriesBuildingRestriction "github.com/malikabdulaziz/tmn-backend/repositories/buildingrestriction"
	webBuildingRestriction "github.com/malikabdulaziz/tmn-backend/web/buildingrestriction"
	"github.com/xuri/excelize/v2"
)

type ServiceBuildingRestrictionImpl struct {
	DB                                    *sql.DB
	RepositoryBuildingRestrictionInterface repositoriesBuildingRestriction.RepositoryBuildingRestrictionInterface
	RepositoryBuildingInterface           repositoriesBuilding.RepositoryBuildingInterface
}

func NewServiceBuildingRestrictionImpl(
	db *sql.DB,
	repoBuildingRestriction repositoriesBuildingRestriction.RepositoryBuildingRestrictionInterface,
	repoBuilding repositoriesBuilding.RepositoryBuildingInterface,
) ServiceBuildingRestrictionInterface {
	return &ServiceBuildingRestrictionImpl{
		DB:                                    db,
		RepositoryBuildingRestrictionInterface: repoBuildingRestriction,
		RepositoryBuildingInterface:           repoBuilding,
	}
}

// validateBuildingIdsErr ensures all building ids exist; panics with BadRequest if any invalid
func (s *ServiceBuildingRestrictionImpl) validateBuildingIdsErr(ctx context.Context, tx *sql.Tx, buildingIds []int) {
	for _, bid := range buildingIds {
		_, err := s.RepositoryBuildingInterface.FindById(ctx, tx, bid)
		if err == sql.ErrNoRows {
			panic(exceptions.NewBadRequest("building not found"))
		}
		helpers.PanicIfError(err)
	}
}

// Create creates a new building restriction with building links
func (s *ServiceBuildingRestrictionImpl) Create(ctx context.Context, request webBuildingRestriction.CreateBuildingRestrictionRequest) webBuildingRestriction.BuildingRestrictionResponse {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	s.validateBuildingIdsErr(ctx, tx, request.BuildingIds)

	restriction := models.BuildingRestriction{Name: request.Name}
	created, err := s.RepositoryBuildingRestrictionInterface.Create(ctx, tx, restriction, request.BuildingIds)
	helpers.PanicIfError(err)
	return s.modelToResponse(created)
}

// FindAll retrieves all building restrictions with pagination
func (s *ServiceBuildingRestrictionImpl) FindAll(ctx context.Context, request webBuildingRestriction.BuildingRestrictionRequestFindAll) ([]webBuildingRestriction.BuildingRestrictionResponse, int) {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	list, err := s.RepositoryBuildingRestrictionInterface.FindAll(ctx, tx, request.GetTake(), request.GetSkip(), request.GetOrderBy(), request.GetOrderDirection())
	helpers.PanicIfError(err)
	total, err := s.RepositoryBuildingRestrictionInterface.CountAll(ctx, tx)
	helpers.PanicIfError(err)

	responses := make([]webBuildingRestriction.BuildingRestrictionResponse, len(list))
	for i, r := range list {
		responses[i] = s.modelToResponse(r)
	}
	return responses, total
}

// FindById retrieves a building restriction by ID
func (s *ServiceBuildingRestrictionImpl) FindById(ctx context.Context, id int) webBuildingRestriction.BuildingRestrictionResponse {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	restriction, err := s.RepositoryBuildingRestrictionInterface.FindById(ctx, tx, id)
	if err == sql.ErrNoRows {
		panic(exceptions.NewNotFoundError("building restriction not found"))
	}
	helpers.PanicIfError(err)
	return s.modelToResponse(restriction)
}

// Update updates a building restriction and replaces building links
func (s *ServiceBuildingRestrictionImpl) Update(ctx context.Context, request webBuildingRestriction.UpdateBuildingRestrictionRequest, id int) webBuildingRestriction.BuildingRestrictionResponse {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	existing, err := s.RepositoryBuildingRestrictionInterface.FindById(ctx, tx, id)
	if err == sql.ErrNoRows {
		panic(exceptions.NewNotFoundError("building restriction not found"))
	}
	helpers.PanicIfError(err)

	s.validateBuildingIdsErr(ctx, tx, request.BuildingIds)

	existing.Name = request.Name
	updated, err := s.RepositoryBuildingRestrictionInterface.Update(ctx, tx, existing, request.BuildingIds)
	helpers.PanicIfError(err)
	return s.modelToResponse(updated)
}

// Delete deletes a building restriction
func (s *ServiceBuildingRestrictionImpl) Delete(ctx context.Context, id int) {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	_, err = s.RepositoryBuildingRestrictionInterface.FindById(ctx, tx, id)
	if err == sql.ErrNoRows {
		panic(exceptions.NewNotFoundError("building restriction not found"))
	}
	helpers.PanicIfError(err)
	err = s.RepositoryBuildingRestrictionInterface.Delete(ctx, tx, id)
	helpers.PanicIfError(err)
}

// Import parses an xlsx or csv file and creates/replaces building restrictions
func (s *ServiceBuildingRestrictionImpl) Import(ctx context.Context, fileBytes []byte, fileType string) []webBuildingRestriction.BuildingRestrictionResponse {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	var rows [][]string
	switch strings.ToLower(fileType) {
	case "xlsx":
		rows, err = brParseXLSX(fileBytes)
		helpers.PanicIfError(err)
	case "csv":
		rows, err = brParseCSV(fileBytes)
		helpers.PanicIfError(err)
	default:
		panic(exceptions.NewBadRequestError("Unsupported file type. Use xlsx or csv."))
	}

	if len(rows) < 2 {
		panic(exceptions.NewBadRequestError("File must contain a header row and at least one data row."))
	}

	header := rows[0]
	colMap := brMapHeaderColumns(header)

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

	// Group rows by restriction name (each row has one building)
	type importGroup struct {
		name        string
		buildingIds []int
	}
	nameOrder := []string{}
	groups := map[string]*importGroup{}

	for _, row := range rows[1:] {
		name := brGetColValue(row, colMap, "name")
		if name == "" {
			continue
		}
		buildingName := strings.TrimSpace(strings.ToLower(brGetColValue(row, colMap, "building_name")))

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

	// Delete existing building restrictions with matching names
	existing, err := s.RepositoryBuildingRestrictionInterface.FindByNames(ctx, tx, nameOrder)
	helpers.PanicIfError(err)
	for _, er := range existing {
		err = s.RepositoryBuildingRestrictionInterface.DeleteBuildingLinksByBuildingRestrictionId(ctx, tx, er.Id)
		helpers.PanicIfError(err)
		err = s.RepositoryBuildingRestrictionInterface.Delete(ctx, tx, er.Id)
		helpers.PanicIfError(err)
	}

	// Create fresh building restrictions
	var responses []webBuildingRestriction.BuildingRestrictionResponse
	for _, name := range nameOrder {
		group := groups[name]
		restriction := models.BuildingRestriction{Name: group.name}
		created, err := s.RepositoryBuildingRestrictionInterface.Create(ctx, tx, restriction, group.buildingIds)
		helpers.PanicIfError(err)
		responses = append(responses, s.modelToResponse(created))
	}

	return responses
}

// Export generates an xlsx file with all building restrictions
func (s *ServiceBuildingRestrictionImpl) Export(ctx context.Context, search string) ([]byte, error) {
	tx, err := s.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer helpers.CommitOrRollback(tx)

	restrictions, err := s.RepositoryBuildingRestrictionInterface.FindAllFlat(ctx, tx, search)
	if err != nil {
		return nil, err
	}

	return buildBuildingRestrictionExcel(restrictions)
}

func (s *ServiceBuildingRestrictionImpl) modelToResponse(r models.BuildingRestriction) webBuildingRestriction.BuildingRestrictionResponse {
	buildings := make([]webBuildingRestriction.BuildingRefResponse, len(r.Buildings))
	for i, b := range r.Buildings {
		buildings[i] = webBuildingRestriction.BuildingRefResponse{Id: b.Id, Name: b.Name}
	}
	return webBuildingRestriction.BuildingRestrictionResponse{
		Id:        r.Id,
		Name:      r.Name,
		Buildings: buildings,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
}

// --- Import helpers ---

func brParseXLSX(fileBytes []byte) ([][]string, error) {
	f, err := excelize.OpenReader(bytes.NewReader(fileBytes))
	if err != nil {
		return nil, err
	}
	defer f.Close()
	sheetName := f.GetSheetName(0)
	return f.GetRows(sheetName)
}

func brParseCSV(fileBytes []byte) ([][]string, error) {
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

func brMapHeaderColumns(header []string) map[string]int {
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

func brGetColValue(row []string, colMap map[string]int, key string) string {
	idx, exists := colMap[key]
	if !exists || idx >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[idx])
}

// --- Export helpers ---

func brMustCell(col, row int) string {
	s, _ := excelize.CoordinatesToCellName(col, row)
	return s
}

func buildBuildingRestrictionExcel(restrictions []models.BuildingRestriction) ([]byte, error) {
	f := excelize.NewFile()
	const sheet = "Sheet1"

	headers := []string{"Name", "Building Name"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = f.SetCellValue(sheet, cell, h)
	}

	rowIdx := 2
	for _, restriction := range restrictions {
		if len(restriction.Buildings) == 0 {
			_ = f.SetCellValue(sheet, brMustCell(1, rowIdx), restriction.Name)
			_ = f.SetCellValue(sheet, brMustCell(2, rowIdx), "")
			rowIdx++
		} else {
			for _, b := range restriction.Buildings {
				_ = f.SetCellValue(sheet, brMustCell(1, rowIdx), restriction.Name)
				_ = f.SetCellValue(sheet, brMustCell(2, rowIdx), b.Name)
				rowIdx++
			}
		}
	}

	_ = f.SetSheetName(sheet, "Building Restrictions")

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
