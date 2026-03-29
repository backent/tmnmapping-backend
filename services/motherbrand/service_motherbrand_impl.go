package motherbrand

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/csv"
	"io"
	"strings"

	"github.com/malikabdulaziz/tmn-backend/exceptions"
	"github.com/malikabdulaziz/tmn-backend/helpers"
	"github.com/malikabdulaziz/tmn-backend/models"
	repositoriesMotherBrand "github.com/malikabdulaziz/tmn-backend/repositories/motherbrand"
	webMotherBrand "github.com/malikabdulaziz/tmn-backend/web/motherbrand"
	"github.com/xuri/excelize/v2"
)

type ServiceMotherBrandImpl struct {
	DB                               *sql.DB
	RepositoryMotherBrandInterface repositoriesMotherBrand.RepositoryMotherBrandInterface
}

func NewServiceMotherBrandImpl(
	db *sql.DB,
	repoMotherBrand repositoriesMotherBrand.RepositoryMotherBrandInterface,
) ServiceMotherBrandInterface {
	return &ServiceMotherBrandImpl{
		DB:                               db,
		RepositoryMotherBrandInterface: repoMotherBrand,
	}
}

func (s *ServiceMotherBrandImpl) Create(ctx context.Context, request webMotherBrand.CreateMotherBrandRequest) webMotherBrand.MotherBrandResponse {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	motherBrand := models.MotherBrand{Name: request.Name}
	created, err := s.RepositoryMotherBrandInterface.Create(ctx, tx, motherBrand)
	helpers.PanicIfError(err)
	return motherBrandModelToResponse(created)
}

func (s *ServiceMotherBrandImpl) FindAll(ctx context.Context, request webMotherBrand.MotherBrandRequestFindAll) ([]webMotherBrand.MotherBrandResponse, int) {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	list, err := s.RepositoryMotherBrandInterface.FindAll(ctx, tx, request.GetTake(), request.GetSkip(), request.GetOrderBy(), request.GetOrderDirection(), request.GetSearch())
	helpers.PanicIfError(err)
	total, err := s.RepositoryMotherBrandInterface.CountAll(ctx, tx, request.GetSearch())
	helpers.PanicIfError(err)

	responses := make([]webMotherBrand.MotherBrandResponse, len(list))
	for i, c := range list {
		responses[i] = motherBrandModelToResponse(c)
	}
	return responses, total
}

func (s *ServiceMotherBrandImpl) FindById(ctx context.Context, id int) webMotherBrand.MotherBrandResponse {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	motherBrand, err := s.RepositoryMotherBrandInterface.FindById(ctx, tx, id)
	if err == sql.ErrNoRows {
		panic(exceptions.NewNotFoundError("mother brand not found"))
	}
	helpers.PanicIfError(err)
	return motherBrandModelToResponse(motherBrand)
}

func (s *ServiceMotherBrandImpl) Update(ctx context.Context, request webMotherBrand.UpdateMotherBrandRequest, id int) webMotherBrand.MotherBrandResponse {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	existing, err := s.RepositoryMotherBrandInterface.FindById(ctx, tx, id)
	if err == sql.ErrNoRows {
		panic(exceptions.NewNotFoundError("mother brand not found"))
	}
	helpers.PanicIfError(err)

	existing.Name = request.Name
	updated, err := s.RepositoryMotherBrandInterface.Update(ctx, tx, existing)
	helpers.PanicIfError(err)
	return motherBrandModelToResponse(updated)
}

func (s *ServiceMotherBrandImpl) Delete(ctx context.Context, id int) {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	_, err = s.RepositoryMotherBrandInterface.FindById(ctx, tx, id)
	if err == sql.ErrNoRows {
		panic(exceptions.NewNotFoundError("mother brand not found"))
	}
	helpers.PanicIfError(err)
	err = s.RepositoryMotherBrandInterface.Delete(ctx, tx, id)
	helpers.PanicIfError(err)
}

func (s *ServiceMotherBrandImpl) Import(ctx context.Context, fileBytes []byte, fileType string) []webMotherBrand.MotherBrandResponse {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	var rows [][]string
	switch strings.ToLower(fileType) {
	case "xlsx":
		rows, err = mbParseXLSX(fileBytes)
		helpers.PanicIfError(err)
	case "csv":
		rows, err = mbParseCSV(fileBytes)
		helpers.PanicIfError(err)
	default:
		panic(exceptions.NewBadRequestError("Unsupported file type. Use xlsx or csv."))
	}

	if len(rows) < 2 {
		panic(exceptions.NewBadRequestError("File must contain a header row and at least one data row."))
	}

	header := rows[0]
	colMap := mbMapHeaderColumns(header)

	if _, exists := colMap["name"]; !exists {
		panic(exceptions.NewBadRequestError("Missing required column: name"))
	}

	var responses []webMotherBrand.MotherBrandResponse
	for _, row := range rows[1:] {
		name := mbGetColValue(row, colMap, "name")
		if name == "" {
			continue
		}

		existing, err := s.RepositoryMotherBrandInterface.FindByName(ctx, tx, name)
		if err == sql.ErrNoRows {
			motherBrand := models.MotherBrand{Name: name}
			created, err := s.RepositoryMotherBrandInterface.Create(ctx, tx, motherBrand)
			helpers.PanicIfError(err)
			responses = append(responses, motherBrandModelToResponse(created))
		} else {
			helpers.PanicIfError(err)
			responses = append(responses, motherBrandModelToResponse(existing))
		}
	}

	return responses
}

func (s *ServiceMotherBrandImpl) Export(ctx context.Context, search string) ([]byte, error) {
	tx, err := s.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer helpers.CommitOrRollback(tx)

	list, err := s.RepositoryMotherBrandInterface.FindAll(ctx, tx, 100000, 0, "name", "ASC", search)
	if err != nil {
		return nil, err
	}

	return buildMotherBrandExcel(list)
}

func motherBrandModelToResponse(c models.MotherBrand) webMotherBrand.MotherBrandResponse {
	return webMotherBrand.MotherBrandResponse{
		Id:        c.Id,
		Name:      c.Name,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}

// --- Import helpers ---

func mbParseXLSX(fileBytes []byte) ([][]string, error) {
	f, err := excelize.OpenReader(bytes.NewReader(fileBytes))
	if err != nil {
		return nil, err
	}
	defer f.Close()
	sheetName := f.GetSheetName(0)
	return f.GetRows(sheetName)
}

func mbParseCSV(fileBytes []byte) ([][]string, error) {
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

func mbMapHeaderColumns(header []string) map[string]int {
	colMap := make(map[string]int)
	for i, h := range header {
		normalized := strings.ToLower(strings.TrimSpace(h))
		if normalized == "name" {
			colMap["name"] = i
		}
	}
	return colMap
}

func mbGetColValue(row []string, colMap map[string]int, key string) string {
	idx, exists := colMap[key]
	if !exists || idx >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[idx])
}

// --- Export helpers ---

func buildMotherBrandExcel(list []models.MotherBrand) ([]byte, error) {
	f := excelize.NewFile()
	const sheet = "Sheet1"

	headers := []string{"Name"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = f.SetCellValue(sheet, cell, h)
	}

	for rowIdx, motherBrand := range list {
		cell, _ := excelize.CoordinatesToCellName(1, rowIdx+2)
		_ = f.SetCellValue(sheet, cell, motherBrand.Name)
	}

	_ = f.SetSheetName(sheet, "MotherBrands")

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
