package subcategory

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
	repositoriesSubCategory "github.com/malikabdulaziz/tmn-backend/repositories/subcategory"
	webSubCategory "github.com/malikabdulaziz/tmn-backend/web/subcategory"
	"github.com/xuri/excelize/v2"
)

type ServiceSubCategoryImpl struct {
	DB                               *sql.DB
	RepositorySubCategoryInterface repositoriesSubCategory.RepositorySubCategoryInterface
}

func NewServiceSubCategoryImpl(
	db *sql.DB,
	repoSubCategory repositoriesSubCategory.RepositorySubCategoryInterface,
) ServiceSubCategoryInterface {
	return &ServiceSubCategoryImpl{
		DB:                               db,
		RepositorySubCategoryInterface: repoSubCategory,
	}
}

func (s *ServiceSubCategoryImpl) Create(ctx context.Context, request webSubCategory.CreateSubCategoryRequest) webSubCategory.SubCategoryResponse {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	subCategory := models.SubCategory{Name: request.Name}
	created, err := s.RepositorySubCategoryInterface.Create(ctx, tx, subCategory)
	helpers.PanicIfError(err)
	return subCategoryModelToResponse(created)
}

func (s *ServiceSubCategoryImpl) FindAll(ctx context.Context, request webSubCategory.SubCategoryRequestFindAll) ([]webSubCategory.SubCategoryResponse, int) {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	list, err := s.RepositorySubCategoryInterface.FindAll(ctx, tx, request.GetTake(), request.GetSkip(), request.GetOrderBy(), request.GetOrderDirection(), request.GetSearch())
	helpers.PanicIfError(err)
	total, err := s.RepositorySubCategoryInterface.CountAll(ctx, tx, request.GetSearch())
	helpers.PanicIfError(err)

	responses := make([]webSubCategory.SubCategoryResponse, len(list))
	for i, c := range list {
		responses[i] = subCategoryModelToResponse(c)
	}
	return responses, total
}

func (s *ServiceSubCategoryImpl) FindById(ctx context.Context, id int) webSubCategory.SubCategoryResponse {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	subCategory, err := s.RepositorySubCategoryInterface.FindById(ctx, tx, id)
	if err == sql.ErrNoRows {
		panic(exceptions.NewNotFoundError("sub category not found"))
	}
	helpers.PanicIfError(err)
	return subCategoryModelToResponse(subCategory)
}

func (s *ServiceSubCategoryImpl) Update(ctx context.Context, request webSubCategory.UpdateSubCategoryRequest, id int) webSubCategory.SubCategoryResponse {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	existing, err := s.RepositorySubCategoryInterface.FindById(ctx, tx, id)
	if err == sql.ErrNoRows {
		panic(exceptions.NewNotFoundError("sub category not found"))
	}
	helpers.PanicIfError(err)

	existing.Name = request.Name
	updated, err := s.RepositorySubCategoryInterface.Update(ctx, tx, existing)
	helpers.PanicIfError(err)
	return subCategoryModelToResponse(updated)
}

func (s *ServiceSubCategoryImpl) Delete(ctx context.Context, id int) {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	_, err = s.RepositorySubCategoryInterface.FindById(ctx, tx, id)
	if err == sql.ErrNoRows {
		panic(exceptions.NewNotFoundError("sub category not found"))
	}
	helpers.PanicIfError(err)
	err = s.RepositorySubCategoryInterface.Delete(ctx, tx, id)
	helpers.PanicIfError(err)
}

func (s *ServiceSubCategoryImpl) Import(ctx context.Context, fileBytes []byte, fileType string) []webSubCategory.SubCategoryResponse {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	var rows [][]string
	switch strings.ToLower(fileType) {
	case "xlsx":
		rows, err = scParseXLSX(fileBytes)
		helpers.PanicIfError(err)
	case "csv":
		rows, err = scParseCSV(fileBytes)
		helpers.PanicIfError(err)
	default:
		panic(exceptions.NewBadRequestError("Unsupported file type. Use xlsx or csv."))
	}

	if len(rows) < 2 {
		panic(exceptions.NewBadRequestError("File must contain a header row and at least one data row."))
	}

	header := rows[0]
	colMap := scMapHeaderColumns(header)

	if _, exists := colMap["name"]; !exists {
		panic(exceptions.NewBadRequestError("Missing required column: name"))
	}

	var responses []webSubCategory.SubCategoryResponse
	for _, row := range rows[1:] {
		name := scGetColValue(row, colMap, "name")
		if name == "" {
			continue
		}

		existing, err := s.RepositorySubCategoryInterface.FindByName(ctx, tx, name)
		if err == sql.ErrNoRows {
			subCategory := models.SubCategory{Name: name}
			created, err := s.RepositorySubCategoryInterface.Create(ctx, tx, subCategory)
			helpers.PanicIfError(err)
			responses = append(responses, subCategoryModelToResponse(created))
		} else {
			helpers.PanicIfError(err)
			responses = append(responses, subCategoryModelToResponse(existing))
		}
	}

	return responses
}

func (s *ServiceSubCategoryImpl) Export(ctx context.Context, search string) ([]byte, error) {
	tx, err := s.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer helpers.CommitOrRollback(tx)

	list, err := s.RepositorySubCategoryInterface.FindAll(ctx, tx, 100000, 0, "name", "ASC", search)
	if err != nil {
		return nil, err
	}

	return buildSubCategoryExcel(list)
}

func subCategoryModelToResponse(c models.SubCategory) webSubCategory.SubCategoryResponse {
	return webSubCategory.SubCategoryResponse{
		Id:        c.Id,
		Name:      c.Name,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}

// --- Import helpers ---

func scParseXLSX(fileBytes []byte) ([][]string, error) {
	f, err := excelize.OpenReader(bytes.NewReader(fileBytes))
	if err != nil {
		return nil, err
	}
	defer f.Close()
	sheetName := f.GetSheetName(0)
	return f.GetRows(sheetName)
}

func scParseCSV(fileBytes []byte) ([][]string, error) {
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

func scMapHeaderColumns(header []string) map[string]int {
	colMap := make(map[string]int)
	for i, h := range header {
		normalized := strings.ToLower(strings.TrimSpace(h))
		if normalized == "name" {
			colMap["name"] = i
		}
	}
	return colMap
}

func scGetColValue(row []string, colMap map[string]int, key string) string {
	idx, exists := colMap[key]
	if !exists || idx >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[idx])
}

// --- Export helpers ---

func buildSubCategoryExcel(list []models.SubCategory) ([]byte, error) {
	f := excelize.NewFile()
	const sheet = "Sheet1"

	headers := []string{"Name"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = f.SetCellValue(sheet, cell, h)
	}

	for rowIdx, subCategory := range list {
		cell, _ := excelize.CoordinatesToCellName(1, rowIdx+2)
		_ = f.SetCellValue(sheet, cell, subCategory.Name)
	}

	_ = f.SetSheetName(sheet, "Sub Categories")

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
