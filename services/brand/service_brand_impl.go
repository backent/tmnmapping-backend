package brand

import (
	"context"
	"database/sql"
	"strconv"
	"strings"

	"github.com/malikabdulaziz/tmn-backend/exceptions"
	"github.com/malikabdulaziz/tmn-backend/helpers"
	"github.com/malikabdulaziz/tmn-backend/models"
	repositoriesBrand "github.com/malikabdulaziz/tmn-backend/repositories/brand"
	repositoriesCustomer "github.com/malikabdulaziz/tmn-backend/repositories/customer"
	"github.com/malikabdulaziz/tmn-backend/spreadsheets"
	"github.com/malikabdulaziz/tmn-backend/web"
	webBrand "github.com/malikabdulaziz/tmn-backend/web/brand"
)

type ServiceBrandImpl struct {
	DB *sql.DB
	repositoriesBrand.RepositoryBrandInterface
	RepositoryCustomer repositoriesCustomer.RepositoryCustomerInterface
}

func NewServiceBrandImpl(
	db *sql.DB,
	repoBrand repositoriesBrand.RepositoryBrandInterface,
	repoCustomer repositoriesCustomer.RepositoryCustomerInterface,
) ServiceBrandInterface {
	return &ServiceBrandImpl{DB: db, RepositoryBrandInterface: repoBrand, RepositoryCustomer: repoCustomer}
}

func (s *ServiceBrandImpl) Create(ctx context.Context, request webBrand.CreateBrandRequest) webBrand.BrandResponse {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	s.assertCustomerExists(ctx, tx, request.CustomerId)
	s.assertCodeFree(ctx, tx, request.Code, 0)

	created, err := s.RepositoryBrandInterface.Create(ctx, tx, models.Brand{
		Code: request.Code, CustomerId: request.CustomerId, Name: request.Name,
		Category: request.Category, Status: request.Status,
		AttentionTo: request.AttentionTo, JobTitle: request.JobTitle,
		ContactPhone: request.ContactPhone, ContactEmail: request.ContactEmail,
	})
	helpers.PanicIfError(err)

	// Re-read so the response carries the joined customer name.
	return s.responseById(ctx, tx, created.Id)
}

func (s *ServiceBrandImpl) FindAll(ctx context.Context, request webBrand.BrandRequestFindAll) ([]webBrand.BrandResponse, int) {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	list, err := s.RepositoryBrandInterface.FindAll(ctx, tx, request.GetTake(), request.GetSkip(),
		request.GetOrderBy(), request.GetOrderDirection(), request.GetSearch(), request.CustomerId)
	helpers.PanicIfError(err)

	total, err := s.RepositoryBrandInterface.CountAll(ctx, tx, request.GetSearch(), request.CustomerId)
	helpers.PanicIfError(err)

	responses := make([]webBrand.BrandResponse, len(list))
	for i, item := range list {
		responses[i] = brandToResponse(item)
	}

	return responses, total
}

func (s *ServiceBrandImpl) FindById(ctx context.Context, id int) webBrand.BrandResponse {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	return s.responseById(ctx, tx, id)
}

func (s *ServiceBrandImpl) Update(ctx context.Context, request webBrand.UpdateBrandRequest, id int) webBrand.BrandResponse {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	existing, err := s.RepositoryBrandInterface.FindById(ctx, tx, id)
	if err == sql.ErrNoRows {
		panic(exceptions.NewNotFoundError("brand not found"))
	}
	helpers.PanicIfError(err)

	s.assertCustomerExists(ctx, tx, request.CustomerId)
	s.assertCodeFree(ctx, tx, request.Code, id)

	existing.Code = request.Code
	existing.CustomerId = request.CustomerId
	existing.Name = request.Name
	existing.Category = request.Category
	existing.Status = request.Status
	existing.AttentionTo = request.AttentionTo
	existing.JobTitle = request.JobTitle
	existing.ContactPhone = request.ContactPhone
	existing.ContactEmail = request.ContactEmail

	_, err = s.RepositoryBrandInterface.Update(ctx, tx, existing)
	helpers.PanicIfError(err)

	return s.responseById(ctx, tx, id)
}

func (s *ServiceBrandImpl) Delete(ctx context.Context, id int) {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	_, err = s.RepositoryBrandInterface.FindById(ctx, tx, id)
	if err == sql.ErrNoRows {
		panic(exceptions.NewNotFoundError("brand not found"))
	}
	helpers.PanicIfError(err)

	helpers.PanicIfError(s.RepositoryBrandInterface.Delete(ctx, tx, id))
}

func (s *ServiceBrandImpl) Template() ([]byte, error) {
	return spreadsheets.BuildTemplate("Brands", TemplateColumns)
}

func (s *ServiceBrandImpl) Export(ctx context.Context, search string) ([]byte, error) {
	tx, err := s.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer helpers.CommitOrRollback(tx)

	list, err := s.RepositoryBrandInterface.FindAll(ctx, tx, 100000, 0, "code", "ASC", search, 0)
	if err != nil {
		return nil, err
	}

	headers := make([]string, len(TemplateColumns))
	for i, col := range TemplateColumns {
		headers[i] = col.Header
	}

	rows := make([][]interface{}, len(list))
	for i, item := range list {
		rows[i] = []interface{}{item.Code, item.CustomerCode, item.Name, item.Category, item.Status,
			item.AttentionTo, item.JobTitle, item.ContactPhone, item.ContactEmail}
	}

	return spreadsheets.BuildExport("Brands", headers, rows)
}

// Import upserts by Brand Code, resolving the owning customer by Customer Code.
func (s *ServiceBrandImpl) Import(ctx context.Context, fileBytes []byte, fileType string) web.ImportResult {
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
		brand    models.Brand
		existing bool
		id       int
	}
	var queue []pending
	seenCodes := map[string]int{}

	for i, row := range rows[1:] {
		rowNumber := i + 2

		code := spreadsheets.ColValue(row, colMap, "code")
		customerCode := spreadsheets.ColValue(row, colMap, "customer_code")
		name := spreadsheets.ColValue(row, colMap, "name")
		category := spreadsheets.ColValue(row, colMap, "category")
		status := strings.ToLower(spreadsheets.ColValue(row, colMap, "status"))
		attentionTo := spreadsheets.ColValue(row, colMap, "attention_to")
		jobTitle := spreadsheets.ColValue(row, colMap, "job_title")
		contactPhone := spreadsheets.ColValue(row, colMap, "contact_phone")
		contactEmail := spreadsheets.ColValue(row, colMap, "contact_email")

		if code == "" && customerCode == "" && name == "" && category == "" && status == "" &&
			attentionTo == "" && jobTitle == "" && contactPhone == "" && contactEmail == "" {
			continue
		}
		result.Rows++

		if code == "" {
			result.AddError(rowNumber, "Brand Code", "", "Brand Code is required")
			continue
		}
		if name == "" {
			result.AddError(rowNumber, "Brand Name", "", "Brand Name is required")
			continue
		}
		if customerCode == "" {
			result.AddError(rowNumber, "Customer Code", "", "Customer Code is required")
			continue
		}
		if status == "" {
			status = models.StatusActive
		}
		if !models.IsValidMasterDataStatus(status) {
			result.AddError(rowNumber, "Status", status, "Status must be active or inactive")
			continue
		}
		if contactError := brandContactError(attentionTo, jobTitle, contactPhone, contactEmail); contactError != "" {
			existingBrand, lookupErr := s.RepositoryBrandInterface.FindByCode(ctx, tx, code)

			// A brand that already carries a contact keeps it when the file leaves
			// these columns blank; only a new brand, or one still without a contact,
			// is refused.
			if lookupErr == sql.ErrNoRows || (lookupErr == nil && existingBrand.AttentionTo == "") {
				result.AddError(rowNumber, "Contact", "", contactError)
				continue
			}
		}

		if firstRow, duplicate := seenCodes[code]; duplicate {
			result.AddError(rowNumber, "Brand Code", code,
				"Duplicate Brand Code in this file (also on row "+strconv.Itoa(firstRow)+")")
			continue
		}
		seenCodes[code] = rowNumber

		customer, err := s.RepositoryCustomer.FindByCode(ctx, tx, customerCode)
		if err == sql.ErrNoRows {
			result.AddError(rowNumber, "Customer Code", customerCode,
				"No customer with this code. Upload the customer file first.")
			continue
		}
		helpers.PanicIfError(err)

		item := pending{brand: models.Brand{
			Code: code, CustomerId: customer.Id, Name: name, Category: category, Status: status,
			AttentionTo: attentionTo, JobTitle: jobTitle,
			ContactPhone: contactPhone, ContactEmail: contactEmail,
		}}

		existing, err := s.RepositoryBrandInterface.FindByCode(ctx, tx, code)
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
			item.brand.Id = item.id
			_, err := s.RepositoryBrandInterface.Update(ctx, tx, item.brand)
			helpers.PanicIfError(err)
			result.Updated++
			continue
		}

		_, err := s.RepositoryBrandInterface.Create(ctx, tx, item.brand)
		helpers.PanicIfError(err)
		result.Created++
	}

	result.Imported = true

	return *result
}

func (s *ServiceBrandImpl) responseById(ctx context.Context, tx *sql.Tx, id int) webBrand.BrandResponse {
	found, err := s.RepositoryBrandInterface.FindById(ctx, tx, id)
	if err == sql.ErrNoRows {
		panic(exceptions.NewNotFoundError("brand not found"))
	}
	helpers.PanicIfError(err)

	return brandToResponse(found)
}

func (s *ServiceBrandImpl) assertCustomerExists(ctx context.Context, tx *sql.Tx, customerId int) {
	_, err := s.RepositoryCustomer.FindById(ctx, tx, customerId)
	if err == sql.ErrNoRows {
		panic(exceptions.NewBadRequestError("customer not found"))
	}
	helpers.PanicIfError(err)
}

func (s *ServiceBrandImpl) assertCodeFree(ctx context.Context, tx *sql.Tx, code string, allowedId int) {
	existing, err := s.RepositoryBrandInterface.FindByCode(ctx, tx, code)
	if err == sql.ErrNoRows {
		return
	}
	helpers.PanicIfError(err)

	if existing.Id != allowedId {
		panic(exceptions.NewBadRequestError("brand code is already in use"))
	}
}

func brandToResponse(b models.Brand) webBrand.BrandResponse {
	return webBrand.BrandResponse{
		Id: b.Id, Code: b.Code, CustomerId: b.CustomerId, CustomerCode: b.CustomerCode,
		CustomerName: b.CustomerName, Name: b.Name, Category: b.Category, Status: b.Status,
		AttentionTo: b.AttentionTo, JobTitle: b.JobTitle,
		ContactPhone: b.ContactPhone, ContactEmail: b.ContactEmail,
		CreatedAt: b.CreatedAt, UpdatedAt: b.UpdatedAt,
	}
}
