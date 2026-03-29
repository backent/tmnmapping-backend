package category

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
	repositoriesCategory "github.com/malikabdulaziz/tmn-backend/repositories/category"
	webCategory "github.com/malikabdulaziz/tmn-backend/web/category"
	"github.com/xuri/excelize/v2"
)

type ServiceCategoryImpl struct {
	DB                          *sql.DB
	RepositoryCategoryInterface repositoriesCategory.RepositoryCategoryInterface
}

func NewServiceCategoryImpl(
	db *sql.DB,
	repoCategory repositoriesCategory.RepositoryCategoryInterface,
) ServiceCategoryInterface {
	return &ServiceCategoryImpl{
		DB:                          db,
		RepositoryCategoryInterface: repoCategory,
	}
}

func (s *ServiceCategoryImpl) Create(ctx context.Context, request webCategory.CreateCategoryRequest) webCategory.CategoryResponse {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	category := models.Category{Name: request.Name}
	created, err := s.RepositoryCategoryInterface.Create(ctx, tx, category)
	helpers.PanicIfError(err)
	return categoryModelToResponse(created)
}

func (s *ServiceCategoryImpl) FindAll(ctx context.Context, request webCategory.CategoryRequestFindAll) ([]webCategory.CategoryResponse, int) {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	list, err := s.RepositoryCategoryInterface.FindAll(ctx, tx, request.GetTake(), request.GetSkip(), request.GetOrderBy(), request.GetOrderDirection(), request.GetSearch())
	helpers.PanicIfError(err)
	total, err := s.RepositoryCategoryInterface.CountAll(ctx, tx, request.GetSearch())
	helpers.PanicIfError(err)

	responses := make([]webCategory.CategoryResponse, len(list))
	for i, c := range list {
		responses[i] = categoryModelToResponse(c)
	}
	return responses, total
}

func (s *ServiceCategoryImpl) FindById(ctx context.Context, id int) webCategory.CategoryResponse {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	category, err := s.RepositoryCategoryInterface.FindById(ctx, tx, id)
	if err == sql.ErrNoRows {
		panic(exceptions.NewNotFoundError("category not found"))
	}
	helpers.PanicIfError(err)
	return categoryModelToResponse(category)
}

func (s *ServiceCategoryImpl) Update(ctx context.Context, request webCategory.UpdateCategoryRequest, id int) webCategory.CategoryResponse {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	existing, err := s.RepositoryCategoryInterface.FindById(ctx, tx, id)
	if err == sql.ErrNoRows {
		panic(exceptions.NewNotFoundError("category not found"))
	}
	helpers.PanicIfError(err)

	existing.Name = request.Name
	updated, err := s.RepositoryCategoryInterface.Update(ctx, tx, existing)
	helpers.PanicIfError(err)
	return categoryModelToResponse(updated)
}

func (s *ServiceCategoryImpl) Delete(ctx context.Context, id int) {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	_, err = s.RepositoryCategoryInterface.FindById(ctx, tx, id)
	if err == sql.ErrNoRows {
		panic(exceptions.NewNotFoundError("category not found"))
	}
	helpers.PanicIfError(err)
	err = s.RepositoryCategoryInterface.Delete(ctx, tx, id)
	helpers.PanicIfError(err)
}

func (s *ServiceCategoryImpl) Import(ctx context.Context, fileBytes []byte, fileType string) []webCategory.CategoryResponse {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	var rows [][]string
	switch strings.ToLower(fileType) {
	case "xlsx":
		rows, err = catParseXLSX(fileBytes)
		helpers.PanicIfError(err)
	case "csv":
		rows, err = catParseCSV(fileBytes)
		helpers.PanicIfError(err)
	default:
		panic(exceptions.NewBadRequestError("Unsupported file type. Use xlsx or csv."))
	}

	if len(rows) < 2 {
		panic(exceptions.NewBadRequestError("File must contain a header row and at least one data row."))
	}

	header := rows[0]
	colMap := catMapHeaderColumns(header)

	if _, exists := colMap["name"]; !exists {
		panic(exceptions.NewBadRequestError("Missing required column: name"))
	}

	var responses []webCategory.CategoryResponse
	for _, row := range rows[1:] {
		name := catGetColValue(row, colMap, "name")
		if name == "" {
			continue
		}

		existing, err := s.RepositoryCategoryInterface.FindByName(ctx, tx, name)
		if err == sql.ErrNoRows {
			category := models.Category{Name: name}
			created, err := s.RepositoryCategoryInterface.Create(ctx, tx, category)
			helpers.PanicIfError(err)
			responses = append(responses, categoryModelToResponse(created))
		} else {
			helpers.PanicIfError(err)
			responses = append(responses, categoryModelToResponse(existing))
		}
	}

	return responses
}

func (s *ServiceCategoryImpl) Export(ctx context.Context, search string) ([]byte, error) {
	tx, err := s.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer helpers.CommitOrRollback(tx)

	list, err := s.RepositoryCategoryInterface.FindAll(ctx, tx, 100000, 0, "name", "ASC", search)
	if err != nil {
		return nil, err
	}

	return buildCategoryExcel(list)
}

func categoryModelToResponse(c models.Category) webCategory.CategoryResponse {
	return webCategory.CategoryResponse{
		Id:        c.Id,
		Name:      c.Name,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}

// --- Import helpers ---

func catParseXLSX(fileBytes []byte) ([][]string, error) {
	f, err := excelize.OpenReader(bytes.NewReader(fileBytes))
	if err != nil {
		return nil, err
	}
	defer f.Close()
	sheetName := f.GetSheetName(0)
	return f.GetRows(sheetName)
}

func catParseCSV(fileBytes []byte) ([][]string, error) {
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

func catMapHeaderColumns(header []string) map[string]int {
	colMap := make(map[string]int)
	for i, h := range header {
		normalized := strings.ToLower(strings.TrimSpace(h))
		if normalized == "name" {
			colMap["name"] = i
		}
	}
	return colMap
}

func catGetColValue(row []string, colMap map[string]int, key string) string {
	idx, exists := colMap[key]
	if !exists || idx >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[idx])
}

// --- Export helpers ---

func buildCategoryExcel(list []models.Category) ([]byte, error) {
	f := excelize.NewFile()
	const sheet = "Sheet1"

	headers := []string{"Name"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = f.SetCellValue(sheet, cell, h)
	}

	for rowIdx, category := range list {
		cell, _ := excelize.CoordinatesToCellName(1, rowIdx+2)
		_ = f.SetCellValue(sheet, cell, category.Name)
	}

	_ = f.SetSheetName(sheet, "Categories")

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
