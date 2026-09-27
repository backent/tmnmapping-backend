package customer

import (
	"context"
	"database/sql"
	"strconv"
	"strings"

	"github.com/malikabdulaziz/tmn-backend/exceptions"
	"github.com/malikabdulaziz/tmn-backend/helpers"
	"github.com/malikabdulaziz/tmn-backend/models"
	repositoriesCustomer "github.com/malikabdulaziz/tmn-backend/repositories/customer"
	"github.com/malikabdulaziz/tmn-backend/spreadsheets"
	"github.com/malikabdulaziz/tmn-backend/web"
	webCustomer "github.com/malikabdulaziz/tmn-backend/web/customer"
)

type ServiceCustomerImpl struct {
	DB *sql.DB
	repositoriesCustomer.RepositoryCustomerInterface
}

func NewServiceCustomerImpl(db *sql.DB, repo repositoriesCustomer.RepositoryCustomerInterface) ServiceCustomerInterface {
	return &ServiceCustomerImpl{DB: db, RepositoryCustomerInterface: repo}
}

func (s *ServiceCustomerImpl) Create(ctx context.Context, request webCustomer.CreateCustomerRequest) webCustomer.CustomerResponse {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	s.assertCodeFree(ctx, tx, request.Code, 0)

	created, err := s.RepositoryCustomerInterface.Create(ctx, tx, models.Customer{
		Code:     request.Code,
		Name:     request.Name,
		Industry: request.Industry,
		Status:   request.Status,
	})
	helpers.PanicIfError(err)

	return customerToResponse(created)
}

func (s *ServiceCustomerImpl) FindAll(ctx context.Context, request webCustomer.CustomerRequestFindAll) ([]webCustomer.CustomerResponse, int) {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	list, err := s.RepositoryCustomerInterface.FindAll(ctx, tx, request.GetTake(), request.GetSkip(),
		request.GetOrderBy(), request.GetOrderDirection(), request.GetSearch())
	helpers.PanicIfError(err)

	total, err := s.RepositoryCustomerInterface.CountAll(ctx, tx, request.GetSearch())
	helpers.PanicIfError(err)

	responses := make([]webCustomer.CustomerResponse, len(list))
	for i, item := range list {
		responses[i] = customerToResponse(item)
	}

	return responses, total
}

func (s *ServiceCustomerImpl) FindById(ctx context.Context, id int) webCustomer.CustomerResponse {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	found, err := s.RepositoryCustomerInterface.FindById(ctx, tx, id)
	if err == sql.ErrNoRows {
		panic(exceptions.NewNotFoundError("customer not found"))
	}
	helpers.PanicIfError(err)

	return customerToResponse(found)
}

func (s *ServiceCustomerImpl) Update(ctx context.Context, request webCustomer.UpdateCustomerRequest, id int) webCustomer.CustomerResponse {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	existing, err := s.RepositoryCustomerInterface.FindById(ctx, tx, id)
	if err == sql.ErrNoRows {
		panic(exceptions.NewNotFoundError("customer not found"))
	}
	helpers.PanicIfError(err)

	s.assertCodeFree(ctx, tx, request.Code, id)

	existing.Code = request.Code
	existing.Name = request.Name
	existing.Industry = request.Industry
	existing.Status = request.Status

	updated, err := s.RepositoryCustomerInterface.Update(ctx, tx, existing)
	helpers.PanicIfError(err)

	return customerToResponse(updated)
}

func (s *ServiceCustomerImpl) Delete(ctx context.Context, id int) {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	_, err = s.RepositoryCustomerInterface.FindById(ctx, tx, id)
	if err == sql.ErrNoRows {
		panic(exceptions.NewNotFoundError("customer not found"))
	}
	helpers.PanicIfError(err)

	// The FK is ON DELETE RESTRICT. Check first so the operator gets a readable
	// message instead of a foreign-key violation surfacing as a 500.
	brands, err := s.RepositoryCustomerInterface.CountBrands(ctx, tx, id)
	helpers.PanicIfError(err)
	if brands > 0 {
		panic(exceptions.NewBadRequestError("this customer still has brands; delete or reassign them first"))
	}

	helpers.PanicIfError(s.RepositoryCustomerInterface.Delete(ctx, tx, id))
}

func (s *ServiceCustomerImpl) Template() ([]byte, error) {
	return spreadsheets.BuildTemplate("Customers", TemplateColumns)
}

func (s *ServiceCustomerImpl) Export(ctx context.Context, search string) ([]byte, error) {
	tx, err := s.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer helpers.CommitOrRollback(tx)

	list, err := s.RepositoryCustomerInterface.FindAll(ctx, tx, 100000, 0, "code", "ASC", search)
	if err != nil {
		return nil, err
	}

	// Export uses the template's headers, so a downloaded file can be edited and
	// uploaded straight back.
	headers := make([]string, len(TemplateColumns))
	for i, col := range TemplateColumns {
		headers[i] = col.Header
	}

	rows := make([][]interface{}, len(list))
	for i, item := range list {
		rows[i] = []interface{}{item.Code, item.Name, item.Industry, item.Status}
	}

	return spreadsheets.BuildExport("Customers", headers, rows)
}

// Import upserts by Customer Code. Every row is validated first; if any row fails,
// nothing is written and the caller gets the full list of problems.
func (s *ServiceCustomerImpl) Import(ctx context.Context, fileBytes []byte, fileType string) web.ImportResult {
	rows, err := spreadsheets.ParseSpreadsheet(fileBytes, fileType)
	if err == spreadsheets.ErrUnsupportedFileType {
		panic(exceptions.NewBadRequestError("Unsupported file type. Use xlsx or csv."))
	}
	helpers.PanicIfError(err)

	if len(rows) < 2 {
		panic(exceptions.NewBadRequestError("File must contain a header row and at least one data row."))
	}

	colMap := spreadsheets.MapHeaderColumns(rows[0], TemplateColumns)
	if missing := spreadsheets.MissingRequiredColumns(colMap, TemplateColumns); len(missing) > 0 {
		panic(exceptions.NewBadRequestError("Missing required column(s): " + strings.Join(missing, ", ")))
	}

	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	result := web.NewImportResult()
	type pending struct {
		customer models.Customer
		existing bool
		id       int
	}
	var queue []pending
	seenCodes := map[string]int{}

	for i, row := range rows[1:] {
		rowNumber := i + 2 // 1-based, and row 1 is the header

		code := spreadsheets.ColValue(row, colMap, "code")
		name := spreadsheets.ColValue(row, colMap, "name")
		industry := spreadsheets.ColValue(row, colMap, "industry")
		status := strings.ToLower(spreadsheets.ColValue(row, colMap, "status"))

		if code == "" && name == "" && industry == "" && status == "" {
			continue // blank row, e.g. trailing formatting in Excel
		}
		result.Rows++

		if code == "" {
			result.AddError(rowNumber, "Customer Code", "", "Customer Code is required")
			continue
		}
		if name == "" {
			result.AddError(rowNumber, "Customer Name", "", "Customer Name is required")
			continue
		}
		if status == "" {
			status = models.StatusActive
		}
		if !models.IsValidMasterDataStatus(status) {
			result.AddError(rowNumber, "Status", status, "Status must be active or inactive")
			continue
		}
		if firstRow, duplicate := seenCodes[code]; duplicate {
			result.AddError(rowNumber, "Customer Code", code,
				"Duplicate Customer Code in this file (also on row "+itoa(firstRow)+")")
			continue
		}
		seenCodes[code] = rowNumber

		item := pending{customer: models.Customer{
			Code: code, Name: name, Industry: industry, Status: status,
		}}

		existing, err := s.RepositoryCustomerInterface.FindByCode(ctx, tx, code)
		if err == nil {
			item.existing = true
			item.id = existing.Id
		} else if err != sql.ErrNoRows {
			helpers.PanicIfError(err)
		}

		queue = append(queue, item)
	}

	if result.HasErrors() {
		return *result
	}

	for _, item := range queue {
		if item.existing {
			item.customer.Id = item.id
			_, err := s.RepositoryCustomerInterface.Update(ctx, tx, item.customer)
			helpers.PanicIfError(err)
			result.Updated++
			continue
		}

		_, err := s.RepositoryCustomerInterface.Create(ctx, tx, item.customer)
		helpers.PanicIfError(err)
		result.Created++
	}

	result.Imported = true

	return *result
}

// assertCodeFree rejects a code already used by a different customer, so the caller
// gets a readable message rather than a unique-constraint violation.
func (s *ServiceCustomerImpl) assertCodeFree(ctx context.Context, tx *sql.Tx, code string, allowedId int) {
	existing, err := s.RepositoryCustomerInterface.FindByCode(ctx, tx, code)
	if err == sql.ErrNoRows {
		return
	}
	helpers.PanicIfError(err)

	if existing.Id != allowedId {
		panic(exceptions.NewBadRequestError("customer code is already in use"))
	}
}

func customerToResponse(c models.Customer) webCustomer.CustomerResponse {
	return webCustomer.CustomerResponse{
		Id: c.Id, Code: c.Code, Name: c.Name, Industry: c.Industry,
		Status: c.Status, CreatedAt: c.CreatedAt, UpdatedAt: c.UpdatedAt,
	}
}

func itoa(i int) string {
	return strconv.Itoa(i)
}
