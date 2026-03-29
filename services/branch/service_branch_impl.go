package branch

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
	repositoriesBranch "github.com/malikabdulaziz/tmn-backend/repositories/branch"
	webBranch "github.com/malikabdulaziz/tmn-backend/web/branch"
	"github.com/xuri/excelize/v2"
)

type ServiceBranchImpl struct {
	DB                          *sql.DB
	RepositoryBranchInterface repositoriesBranch.RepositoryBranchInterface
}

func NewServiceBranchImpl(
	db *sql.DB,
	repoBranch repositoriesBranch.RepositoryBranchInterface,
) ServiceBranchInterface {
	return &ServiceBranchImpl{
		DB:                          db,
		RepositoryBranchInterface: repoBranch,
	}
}

func (s *ServiceBranchImpl) Create(ctx context.Context, request webBranch.CreateBranchRequest) webBranch.BranchResponse {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	branch := models.Branch{Name: request.Name}
	created, err := s.RepositoryBranchInterface.Create(ctx, tx, branch)
	helpers.PanicIfError(err)
	return branchModelToResponse(created)
}

func (s *ServiceBranchImpl) FindAll(ctx context.Context, request webBranch.BranchRequestFindAll) ([]webBranch.BranchResponse, int) {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	list, err := s.RepositoryBranchInterface.FindAll(ctx, tx, request.GetTake(), request.GetSkip(), request.GetOrderBy(), request.GetOrderDirection(), request.GetSearch())
	helpers.PanicIfError(err)
	total, err := s.RepositoryBranchInterface.CountAll(ctx, tx, request.GetSearch())
	helpers.PanicIfError(err)

	responses := make([]webBranch.BranchResponse, len(list))
	for i, c := range list {
		responses[i] = branchModelToResponse(c)
	}
	return responses, total
}

func (s *ServiceBranchImpl) FindById(ctx context.Context, id int) webBranch.BranchResponse {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	branch, err := s.RepositoryBranchInterface.FindById(ctx, tx, id)
	if err == sql.ErrNoRows {
		panic(exceptions.NewNotFoundError("branch not found"))
	}
	helpers.PanicIfError(err)
	return branchModelToResponse(branch)
}

func (s *ServiceBranchImpl) Update(ctx context.Context, request webBranch.UpdateBranchRequest, id int) webBranch.BranchResponse {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	existing, err := s.RepositoryBranchInterface.FindById(ctx, tx, id)
	if err == sql.ErrNoRows {
		panic(exceptions.NewNotFoundError("branch not found"))
	}
	helpers.PanicIfError(err)

	existing.Name = request.Name
	updated, err := s.RepositoryBranchInterface.Update(ctx, tx, existing)
	helpers.PanicIfError(err)
	return branchModelToResponse(updated)
}

func (s *ServiceBranchImpl) Delete(ctx context.Context, id int) {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	_, err = s.RepositoryBranchInterface.FindById(ctx, tx, id)
	if err == sql.ErrNoRows {
		panic(exceptions.NewNotFoundError("branch not found"))
	}
	helpers.PanicIfError(err)
	err = s.RepositoryBranchInterface.Delete(ctx, tx, id)
	helpers.PanicIfError(err)
}

func (s *ServiceBranchImpl) Import(ctx context.Context, fileBytes []byte, fileType string) []webBranch.BranchResponse {
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

	var responses []webBranch.BranchResponse
	for _, row := range rows[1:] {
		name := brGetColValue(row, colMap, "name")
		if name == "" {
			continue
		}

		existing, err := s.RepositoryBranchInterface.FindByName(ctx, tx, name)
		if err == sql.ErrNoRows {
			branch := models.Branch{Name: name}
			created, err := s.RepositoryBranchInterface.Create(ctx, tx, branch)
			helpers.PanicIfError(err)
			responses = append(responses, branchModelToResponse(created))
		} else {
			helpers.PanicIfError(err)
			responses = append(responses, branchModelToResponse(existing))
		}
	}

	return responses
}

func (s *ServiceBranchImpl) Export(ctx context.Context, search string) ([]byte, error) {
	tx, err := s.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer helpers.CommitOrRollback(tx)

	list, err := s.RepositoryBranchInterface.FindAll(ctx, tx, 100000, 0, "name", "ASC", search)
	if err != nil {
		return nil, err
	}

	return buildBranchExcel(list)
}

func branchModelToResponse(c models.Branch) webBranch.BranchResponse {
	return webBranch.BranchResponse{
		Id:        c.Id,
		Name:      c.Name,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
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
		if normalized == "name" {
			colMap["name"] = i
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

func buildBranchExcel(list []models.Branch) ([]byte, error) {
	f := excelize.NewFile()
	const sheet = "Sheet1"

	headers := []string{"Name"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = f.SetCellValue(sheet, cell, h)
	}

	for rowIdx, branch := range list {
		cell, _ := excelize.CoordinatesToCellName(1, rowIdx+2)
		_ = f.SetCellValue(sheet, cell, branch.Name)
	}

	_ = f.SetSheetName(sheet, "Branches")

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
