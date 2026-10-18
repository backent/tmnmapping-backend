package salesassignment

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/malikabdulaziz/tmn-backend/exceptions"
	"github.com/malikabdulaziz/tmn-backend/helpers"
	"github.com/malikabdulaziz/tmn-backend/models"
	repositoriesBrand "github.com/malikabdulaziz/tmn-backend/repositories/brand"
	repositoriesCustomer "github.com/malikabdulaziz/tmn-backend/repositories/customer"
	repositoriesSalesAssignment "github.com/malikabdulaziz/tmn-backend/repositories/salesassignment"
	repositoriesUser "github.com/malikabdulaziz/tmn-backend/repositories/user"
	"github.com/malikabdulaziz/tmn-backend/spreadsheets"
	"github.com/malikabdulaziz/tmn-backend/web"
	webSalesAssignment "github.com/malikabdulaziz/tmn-backend/web/salesassignment"
)

const dateLayout = "2006-01-02"

// errDateFormat marks a cell whose date could not be parsed. It is compared by the
// caller, never surfaced directly -- the operator sees a per-row import error instead.
var errDateFormat = errors.New("unparseable date")

type ServiceSalesAssignmentImpl struct {
	DB *sql.DB
	repositoriesSalesAssignment.RepositorySalesAssignmentInterface
	RepositoryCustomer repositoriesCustomer.RepositoryCustomerInterface
	RepositoryBrand    repositoriesBrand.RepositoryBrandInterface
	RepositoryUser     repositoriesUser.RepositoryUserInterface
}

func NewServiceSalesAssignmentImpl(
	db *sql.DB,
	repoAssignment repositoriesSalesAssignment.RepositorySalesAssignmentInterface,
	repoCustomer repositoriesCustomer.RepositoryCustomerInterface,
	repoBrand repositoriesBrand.RepositoryBrandInterface,
	repoUser repositoriesUser.RepositoryUserInterface,
) ServiceSalesAssignmentInterface {
	return &ServiceSalesAssignmentImpl{
		DB:                                 db,
		RepositorySalesAssignmentInterface: repoAssignment,
		RepositoryCustomer:                 repoCustomer,
		RepositoryBrand:                    repoBrand,
		RepositoryUser:                     repoUser,
	}
}

func (s *ServiceSalesAssignmentImpl) Create(ctx context.Context, request webSalesAssignment.CreateSalesAssignmentRequest) webSalesAssignment.SalesAssignmentResponse {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	s.assertReferencesValid(ctx, tx, request.CustomerId, request.BrandId, request.SalesUserId)
	s.assertPairFree(ctx, tx, request.CustomerId, request.BrandId, 0)

	created, err := s.RepositorySalesAssignmentInterface.Create(ctx, tx, models.SalesAssignment{
		CustomerId: request.CustomerId, BrandId: request.BrandId, SalesUserId: request.SalesUserId,
		Status: request.Status, RegistrationDate: request.RegistrationDate, ExpiryDate: request.ExpiryDate,
	})
	helpers.PanicIfError(err)

	return s.responseById(ctx, tx, created.Id)
}

func (s *ServiceSalesAssignmentImpl) FindAll(ctx context.Context, request webSalesAssignment.SalesAssignmentRequestFindAll) ([]webSalesAssignment.SalesAssignmentResponse, int) {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	list, err := s.RepositorySalesAssignmentInterface.FindAll(ctx, tx, request.GetTake(), request.GetSkip(),
		request.GetOrderBy(), request.GetOrderDirection(), request.GetSearch())
	helpers.PanicIfError(err)

	total, err := s.RepositorySalesAssignmentInterface.CountAll(ctx, tx, request.GetSearch())
	helpers.PanicIfError(err)

	responses := make([]webSalesAssignment.SalesAssignmentResponse, len(list))
	for i, item := range list {
		responses[i] = assignmentToResponse(item)
	}

	return responses, total
}

func (s *ServiceSalesAssignmentImpl) FindById(ctx context.Context, id int) webSalesAssignment.SalesAssignmentResponse {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	return s.responseById(ctx, tx, id)
}

func (s *ServiceSalesAssignmentImpl) Update(ctx context.Context, request webSalesAssignment.UpdateSalesAssignmentRequest, id int) webSalesAssignment.SalesAssignmentResponse {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	existing, err := s.RepositorySalesAssignmentInterface.FindById(ctx, tx, id)
	if err == sql.ErrNoRows {
		panic(exceptions.NewNotFoundError("sales assignment not found"))
	}
	helpers.PanicIfError(err)

	s.assertReferencesValid(ctx, tx, request.CustomerId, request.BrandId, request.SalesUserId)
	s.assertPairFree(ctx, tx, request.CustomerId, request.BrandId, id)

	existing.CustomerId = request.CustomerId
	existing.BrandId = request.BrandId
	existing.SalesUserId = request.SalesUserId
	existing.Status = request.Status
	existing.RegistrationDate = request.RegistrationDate
	existing.ExpiryDate = request.ExpiryDate

	_, err = s.RepositorySalesAssignmentInterface.Update(ctx, tx, existing)
	helpers.PanicIfError(err)

	return s.responseById(ctx, tx, id)
}

func (s *ServiceSalesAssignmentImpl) Delete(ctx context.Context, id int) {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	_, err = s.RepositorySalesAssignmentInterface.FindById(ctx, tx, id)
	if err == sql.ErrNoRows {
		panic(exceptions.NewNotFoundError("sales assignment not found"))
	}
	helpers.PanicIfError(err)

	helpers.PanicIfError(s.RepositorySalesAssignmentInterface.Delete(ctx, tx, id))
}

func (s *ServiceSalesAssignmentImpl) Template() ([]byte, error) {
	return spreadsheets.BuildTemplate("Sales Assignments", TemplateColumns)
}

func (s *ServiceSalesAssignmentImpl) Export(ctx context.Context, search string) ([]byte, error) {
	tx, err := s.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer helpers.CommitOrRollback(tx)

	list, err := s.RepositorySalesAssignmentInterface.FindAll(ctx, tx, 100000, 0, "customer_name", "ASC", search)
	if err != nil {
		return nil, err
	}

	headers := make([]string, len(TemplateColumns))
	for i, col := range TemplateColumns {
		headers[i] = col.Header
	}

	rows := make([][]interface{}, len(list))
	for i, item := range list {
		rows[i] = []interface{}{
			item.CustomerCode, item.BrandCode, item.SalesUsername, item.Status,
			formatDate(item.RegistrationDate), formatDate(item.ExpiryDate),
		}
	}

	return spreadsheets.BuildExport("Sales Assignments", headers, rows)
}

// Import upserts by the customer+brand pair, which is the table's unique key: one
// PIC per brand. Re-uploading a pair reassigns it rather than failing.
func (s *ServiceSalesAssignmentImpl) Import(ctx context.Context, fileBytes []byte, fileType string) web.ImportResult {
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
		assignment models.SalesAssignment
		existing   bool
		id         int
	}
	var queue []pending
	seenPairs := map[string]int{}

	for i, row := range rows[1:] {
		rowNumber := i + 2

		customerCode := spreadsheets.ColValue(row, colMap, "customer_code")
		brandCode := spreadsheets.ColValue(row, colMap, "brand_code")
		username := spreadsheets.ColValue(row, colMap, "sales_username")
		status := strings.ToLower(spreadsheets.ColValue(row, colMap, "status"))
		registration := spreadsheets.ColValue(row, colMap, "registration_date")
		expiry := spreadsheets.ColValue(row, colMap, "expiry_date")

		if customerCode == "" && brandCode == "" && username == "" && status == "" &&
			registration == "" && expiry == "" {
			continue
		}
		result.Rows++

		if customerCode == "" || brandCode == "" || username == "" {
			result.AddError(rowNumber, "Customer Code / Brand Code / Sales Username", "",
				"Customer Code, Brand Code and Sales Username are all required")
			continue
		}
		if status == "" {
			status = models.StatusActive
		}
		if !models.IsValidMasterDataStatus(status) {
			result.AddError(rowNumber, "Status", status, "Status must be active or inactive")
			continue
		}

		registration, err = normalizeDate(registration)
		if err != nil {
			result.AddError(rowNumber, "Registration Date", registration, "Use the format YYYY-MM-DD")
			continue
		}
		expiry, err = normalizeDate(expiry)
		if err != nil {
			result.AddError(rowNumber, "Expiry Date", expiry, "Use the format YYYY-MM-DD")
			continue
		}
		if registration != "" && expiry != "" && expiry < registration {
			result.AddError(rowNumber, "Expiry Date", expiry,
				"Expiry Date cannot be earlier than Registration Date")
			continue
		}

		pairKey := customerCode + "\x00" + brandCode
		if firstRow, duplicate := seenPairs[pairKey]; duplicate {
			result.AddError(rowNumber, "Brand Code", brandCode,
				"This customer and brand already appear on row "+strconv.Itoa(firstRow)+
					". A brand can have only one sales PIC.")
			continue
		}
		seenPairs[pairKey] = rowNumber

		customer, err := s.RepositoryCustomer.FindByCode(ctx, tx, customerCode)
		if err == sql.ErrNoRows {
			result.AddError(rowNumber, "Customer Code", customerCode, "No customer with this code")
			continue
		}
		helpers.PanicIfError(err)

		brand, err := s.RepositoryBrand.FindByCode(ctx, tx, brandCode)
		if err == sql.ErrNoRows {
			result.AddError(rowNumber, "Brand Code", brandCode, "No brand with this code")
			continue
		}
		helpers.PanicIfError(err)

		if brand.CustomerId != customer.Id {
			result.AddError(rowNumber, "Brand Code", brandCode,
				"This brand belongs to "+brand.CustomerCode+", not "+customerCode)
			continue
		}

		user, err := s.RepositoryUser.FindByUsername(ctx, tx, username)
		if err == sql.ErrNoRows {
			result.AddError(rowNumber, "Sales Username", username,
				"No user with this username. Create the account first.")
			continue
		}
		helpers.PanicIfError(err)

		item := pending{assignment: models.SalesAssignment{
			CustomerId: customer.Id, BrandId: brand.Id, SalesUserId: user.Id,
			Status: status, RegistrationDate: registration, ExpiryDate: expiry,
		}}

		existing, err := s.RepositorySalesAssignmentInterface.FindByCustomerAndBrand(ctx, tx, customer.Id, brand.Id)
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
			item.assignment.Id = item.id
			_, err := s.RepositorySalesAssignmentInterface.Update(ctx, tx, item.assignment)
			helpers.PanicIfError(err)
			result.Updated++
			continue
		}

		_, err := s.RepositorySalesAssignmentInterface.Create(ctx, tx, item.assignment)
		helpers.PanicIfError(err)
		result.Created++
	}

	result.Imported = true

	return *result
}

func (s *ServiceSalesAssignmentImpl) responseById(ctx context.Context, tx *sql.Tx, id int) webSalesAssignment.SalesAssignmentResponse {
	found, err := s.RepositorySalesAssignmentInterface.FindById(ctx, tx, id)
	if err == sql.ErrNoRows {
		panic(exceptions.NewNotFoundError("sales assignment not found"))
	}
	helpers.PanicIfError(err)

	return assignmentToResponse(found)
}

func (s *ServiceSalesAssignmentImpl) assertReferencesValid(ctx context.Context, tx *sql.Tx, customerId, brandId, salesUserId int) {
	if _, err := s.RepositoryCustomer.FindById(ctx, tx, customerId); err == sql.ErrNoRows {
		panic(exceptions.NewBadRequestError("customer not found"))
	} else {
		helpers.PanicIfError(err)
	}

	brand, err := s.RepositoryBrand.FindById(ctx, tx, brandId)
	if err == sql.ErrNoRows {
		panic(exceptions.NewBadRequestError("brand not found"))
	}
	helpers.PanicIfError(err)

	if brand.CustomerId != customerId {
		panic(exceptions.NewBadRequestError("that brand does not belong to the selected customer"))
	}

	if _, err := s.RepositoryUser.FindById(ctx, tx, salesUserId); err == sql.ErrNoRows {
		panic(exceptions.NewBadRequestError("sales user not found"))
	} else {
		helpers.PanicIfError(err)
	}
}

// assertPairFree enforces one PIC per customer+brand ahead of the unique constraint.
func (s *ServiceSalesAssignmentImpl) assertPairFree(ctx context.Context, tx *sql.Tx, customerId, brandId, allowedId int) {
	existing, err := s.RepositorySalesAssignmentInterface.FindByCustomerAndBrand(ctx, tx, customerId, brandId)
	if err == sql.ErrNoRows {
		return
	}
	helpers.PanicIfError(err)

	if existing.Id != allowedId {
		panic(exceptions.NewBadRequestError("that brand already has a sales PIC assigned"))
	}
}

// normalizeDate accepts YYYY-MM-DD, and also the RFC3339 timestamps Excel produces
// when a cell is formatted as a date rather than text.
func normalizeDate(value string) (string, error) {
	if value == "" {
		return "", nil
	}

	for _, layout := range []string{dateLayout, time.RFC3339, "2006-01-02 15:04:05", "01/02/2006", "02/01/2006"} {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed.Format(dateLayout), nil
		}
	}

	return value, errDateFormat
}

func formatDate(value string) string {
	normalized, err := normalizeDate(value)
	if err != nil {
		return value
	}

	return normalized
}

func assignmentToResponse(a models.SalesAssignment) webSalesAssignment.SalesAssignmentResponse {
	return webSalesAssignment.SalesAssignmentResponse{
		Id: a.Id, CustomerId: a.CustomerId, CustomerCode: a.CustomerCode, CustomerName: a.CustomerName,
		BrandId: a.BrandId, BrandCode: a.BrandCode, BrandName: a.BrandName,
		SalesUserId: a.SalesUserId, SalesUsername: a.SalesUsername, SalesName: a.SalesName,
		Status: a.Status, RegistrationDate: formatDate(a.RegistrationDate),
		ExpiryDate: formatDate(a.ExpiryDate), CreatedAt: a.CreatedAt, UpdatedAt: a.UpdatedAt,
	}
}
