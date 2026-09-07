package ratecard

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"

	"github.com/malikabdulaziz/tmn-backend/exceptions"
	"github.com/malikabdulaziz/tmn-backend/helpers"
	"github.com/malikabdulaziz/tmn-backend/models"
	repositoriesBuilding "github.com/malikabdulaziz/tmn-backend/repositories/building"
	repositoriesRateCard "github.com/malikabdulaziz/tmn-backend/repositories/ratecard"
	repositoriesSalesPackage "github.com/malikabdulaziz/tmn-backend/repositories/salespackage"
	"github.com/malikabdulaziz/tmn-backend/spreadsheets"
	"github.com/malikabdulaziz/tmn-backend/web"
	webRateCard "github.com/malikabdulaziz/tmn-backend/web/ratecard"
)

type ServiceRateCardImpl struct {
	DB *sql.DB
	repositoriesRateCard.RepositoryRateCardInterface
	RepositoryBuilding     repositoriesBuilding.RepositoryBuildingInterface
	RepositorySalesPackage repositoriesSalesPackage.RepositorySalesPackageInterface
}

func NewServiceRateCardImpl(
	db *sql.DB,
	repoRateCard repositoriesRateCard.RepositoryRateCardInterface,
	repoBuilding repositoriesBuilding.RepositoryBuildingInterface,
	repoSalesPackage repositoriesSalesPackage.RepositorySalesPackageInterface,
) ServiceRateCardInterface {
	return &ServiceRateCardImpl{
		DB:                          db,
		RepositoryRateCardInterface: repoRateCard,
		RepositoryBuilding:          repoBuilding,
		RepositorySalesPackage:      repoSalesPackage,
	}
}

// ---------------------------------------------------------------------------
// Versions
// ---------------------------------------------------------------------------

func (s *ServiceRateCardImpl) CreateVersion(ctx context.Context, request webRateCard.CreateVersionRequest) webRateCard.VersionResponse {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	s.assertVersionCodeFree(ctx, tx, request.VersionCode, 0)

	currency := request.Currency
	if currency == "" {
		currency = "IDR"
	}

	created, err := s.RepositoryRateCardInterface.CreateVersion(ctx, tx, models.RateCardVersion{
		VersionCode: request.VersionCode,
		Description: request.Description,
		Currency:    currency,
		Status:      models.RateCardStatusDraft,
	})
	helpers.PanicIfError(err)

	if request.CopyFromVersionId > 0 {
		if _, err := s.RepositoryRateCardInterface.FindVersionById(ctx, tx, request.CopyFromVersionId); err == sql.ErrNoRows {
			panic(exceptions.NewBadRequestError("the version to copy from does not exist"))
		} else {
			helpers.PanicIfError(err)
		}

		helpers.PanicIfError(s.RepositoryRateCardInterface.CopyPrices(ctx, tx, request.CopyFromVersionId, created.Id))
	}

	return s.versionResponse(ctx, tx, created.Id)
}

func (s *ServiceRateCardImpl) FindAllVersions(ctx context.Context, request webRateCard.RateCardRequestFindAll) ([]webRateCard.VersionResponse, int) {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	list, err := s.RepositoryRateCardInterface.FindAllVersions(ctx, tx, request.GetTake(), request.GetSkip(),
		request.GetOrderBy(), request.GetOrderDirection(), request.GetSearch())
	helpers.PanicIfError(err)

	total, err := s.RepositoryRateCardInterface.CountAllVersions(ctx, tx, request.GetSearch())
	helpers.PanicIfError(err)

	responses := make([]webRateCard.VersionResponse, len(list))
	for i, item := range list {
		responses[i] = versionToResponse(item)
	}

	return responses, total
}

func (s *ServiceRateCardImpl) FindVersionById(ctx context.Context, id int) webRateCard.VersionResponse {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	return s.versionResponse(ctx, tx, id)
}

// FindCurrentVersion is what the quotation wizard will price against.
func (s *ServiceRateCardImpl) FindCurrentVersion(ctx context.Context) webRateCard.VersionResponse {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	found, err := s.RepositoryRateCardInterface.FindCurrentVersion(ctx, tx)
	if err == sql.ErrNoRows {
		panic(exceptions.NewNotFoundError("no rate card has been published yet"))
	}
	helpers.PanicIfError(err)

	return versionToResponse(found)
}

func (s *ServiceRateCardImpl) UpdateVersion(ctx context.Context, request webRateCard.UpdateVersionRequest, id int) webRateCard.VersionResponse {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	existing := s.mustFindEditableVersion(ctx, tx, id)
	s.assertVersionCodeFree(ctx, tx, request.VersionCode, id)

	existing.VersionCode = request.VersionCode
	existing.Description = request.Description
	if request.Currency != "" {
		existing.Currency = request.Currency
	}

	_, err = s.RepositoryRateCardInterface.UpdateVersion(ctx, tx, existing)
	helpers.PanicIfError(err)

	return s.versionResponse(ctx, tx, id)
}

func (s *ServiceRateCardImpl) DeleteVersion(ctx context.Context, id int) {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	// Only a draft can be deleted. A published version is what approved quotations
	// were priced against, so removing it would rewrite their history.
	s.mustFindEditableVersion(ctx, tx, id)

	helpers.PanicIfError(s.RepositoryRateCardInterface.DeleteVersion(ctx, tx, id))
}

// PublishVersion promotes a draft to current, demoting whatever was current to
// historical, and freezes each priced package's building membership.
//
// All of it happens in one transaction so the partial unique index never sees two
// current versions, and a failure part-way leaves nothing half-published.
func (s *ServiceRateCardImpl) PublishVersion(ctx context.Context, id int, actorUserId int) webRateCard.VersionResponse {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	version := s.mustFindEditableVersion(ctx, tx, id)

	if version.BuildingPriceCount == 0 && version.PackagePriceCount == 0 {
		panic(exceptions.NewBadRequestError("this rate card has no prices; add some before publishing"))
	}

	// Freeze the composition of every priced package as it stands right now.
	packagePrices, err := s.RepositoryRateCardInterface.FindPackagePrices(ctx, tx, id, 100000, 0, "")
	helpers.PanicIfError(err)

	for _, packagePrice := range packagePrices {
		buildingIds, err := s.RepositoryRateCardInterface.FindLivePackageBuildingIds(ctx, tx, packagePrice.SalesPackageId)
		helpers.PanicIfError(err)

		if len(buildingIds) == 0 {
			panic(exceptions.NewBadRequestError(
				"sales package \"" + packagePrice.SalesPackageName + "\" is priced but contains no buildings"))
		}

		helpers.PanicIfError(s.RepositoryRateCardInterface.ReplacePackageBuildings(ctx, tx, id, packagePrice.SalesPackageId, buildingIds))
	}

	helpers.PanicIfError(s.RepositoryRateCardInterface.DemoteCurrentVersion(ctx, tx))
	helpers.PanicIfError(s.RepositoryRateCardInterface.SetVersionStatus(ctx, tx, id, models.RateCardStatusCurrent, actorUserId))

	return s.versionResponse(ctx, tx, id)
}

// ---------------------------------------------------------------------------
// Building prices
// ---------------------------------------------------------------------------

func (s *ServiceRateCardImpl) FindBuildingPrices(ctx context.Context, versionId int, request webRateCard.RateCardRequestFindAll) ([]webRateCard.BuildingPriceResponse, int) {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	s.mustFindVersion(ctx, tx, versionId)

	list, err := s.RepositoryRateCardInterface.FindBuildingPrices(ctx, tx, versionId,
		request.GetTake(), request.GetSkip(), request.GetSearch())
	helpers.PanicIfError(err)

	total, err := s.RepositoryRateCardInterface.CountBuildingPrices(ctx, tx, versionId, request.GetSearch())
	helpers.PanicIfError(err)

	responses := make([]webRateCard.BuildingPriceResponse, len(list))
	for i, item := range list {
		responses[i] = buildingPriceToResponse(item)
	}

	return responses, total
}

func (s *ServiceRateCardImpl) UpsertBuildingPrice(ctx context.Context, versionId int, request webRateCard.UpsertBuildingPriceRequest) webRateCard.BuildingPriceResponse {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	s.mustFindEditableVersion(ctx, tx, versionId)

	if _, err := s.RepositoryBuilding.FindById(ctx, tx, request.BuildingId); err == sql.ErrNoRows {
		panic(exceptions.NewBadRequestError("building not found"))
	} else {
		helpers.PanicIfError(err)
	}

	helpers.PanicIfError(s.RepositoryRateCardInterface.UpsertBuildingPrice(ctx, tx, models.RateCardBuildingPrice{
		RateCardVersionId: versionId,
		BuildingId:        request.BuildingId,
		PriceIdrPer4Weeks: request.PriceIdrPer4Weeks,
	}))

	prices, err := s.RepositoryRateCardInterface.FindBuildingPrices(ctx, tx, versionId, 1, 0, "")
	helpers.PanicIfError(err)
	for _, price := range prices {
		if price.BuildingId == request.BuildingId {
			return buildingPriceToResponse(price)
		}
	}

	return webRateCard.BuildingPriceResponse{
		RateCardVersionId: versionId,
		BuildingId:        request.BuildingId,
		PriceIdrPer4Weeks: request.PriceIdrPer4Weeks,
	}
}

func (s *ServiceRateCardImpl) DeleteBuildingPrice(ctx context.Context, versionId int, buildingId int) {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	s.mustFindEditableVersion(ctx, tx, versionId)
	helpers.PanicIfError(s.RepositoryRateCardInterface.DeleteBuildingPrice(ctx, tx, versionId, buildingId))
}

func (s *ServiceRateCardImpl) BuildingPriceTemplate() ([]byte, error) {
	return spreadsheets.BuildTemplate("Building Prices", BuildingPriceColumns)
}

func (s *ServiceRateCardImpl) ExportBuildingPrices(ctx context.Context, versionId int) ([]byte, error) {
	tx, err := s.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer helpers.CommitOrRollback(tx)

	list, err := s.RepositoryRateCardInterface.FindBuildingPrices(ctx, tx, versionId, 100000, 0, "")
	if err != nil {
		return nil, err
	}

	headers := headersOf(BuildingPriceColumns)
	rows := make([][]interface{}, len(list))
	for i, item := range list {
		// The export carries External Building ID so the file can be edited and
		// uploaded straight back; iris_code is shown for readability only.
		building, err := s.RepositoryBuilding.FindById(ctx, tx, item.BuildingId)
		externalId := ""
		if err == nil {
			externalId = building.ExternalBuildingId
		}
		rows[i] = []interface{}{externalId, item.BuildingName, item.PriceIdrPer4Weeks}
	}

	return spreadsheets.BuildExport("Building Prices", headers, rows)
}

func (s *ServiceRateCardImpl) ImportBuildingPrices(ctx context.Context, versionId int, fileBytes []byte, fileType string) web.ImportResult {
	rows := s.parseUpload(fileBytes, fileType, BuildingPriceColumns)

	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	s.mustFindEditableVersion(ctx, tx, versionId)

	colMap := spreadsheets.MapHeaderColumns(rows[0], BuildingPriceColumns)
	result := web.NewImportResult()

	type pending struct {
		buildingId int
		price      int64
	}
	var queue []pending
	seen := map[string]int{}

	for i, row := range rows[1:] {
		rowNumber := i + 2

		externalId := spreadsheets.ColValue(row, colMap, "external_building_id")
		rawPrice := spreadsheets.ColValue(row, colMap, "price")

		if externalId == "" && rawPrice == "" {
			continue
		}
		result.Rows++

		if externalId == "" {
			result.AddError(rowNumber, "External Building ID", "", "External Building ID is required")
			continue
		}
		if firstRow, duplicate := seen[externalId]; duplicate {
			result.AddError(rowNumber, "External Building ID", externalId,
				"This building is priced twice in this file (also on row "+strconv.Itoa(firstRow)+")")
			continue
		}
		seen[externalId] = rowNumber

		price, err := parsePrice(rawPrice)
		if err != nil {
			result.AddError(rowNumber, "Price per 4 Weeks (IDR)", rawPrice, err.Error())
			continue
		}

		building, err := s.RepositoryBuilding.FindByExternalId(ctx, tx, externalId)
		if err == sql.ErrNoRows {
			result.AddError(rowNumber, "External Building ID", externalId,
				"No building with this ID. Export the current list to get valid values.")
			continue
		}
		helpers.PanicIfError(err)

		queue = append(queue, pending{buildingId: building.Id, price: price})
	}

	if result.HasErrors() {
		return *result
	}

	existing, err := s.RepositoryRateCardInterface.CountBuildingPrices(ctx, tx, versionId, "")
	helpers.PanicIfError(err)

	for _, item := range queue {
		helpers.PanicIfError(s.RepositoryRateCardInterface.UpsertBuildingPrice(ctx, tx, models.RateCardBuildingPrice{
			RateCardVersionId: versionId,
			BuildingId:        item.buildingId,
			PriceIdrPer4Weeks: item.price,
		}))
	}

	after, err := s.RepositoryRateCardInterface.CountBuildingPrices(ctx, tx, versionId, "")
	helpers.PanicIfError(err)

	result.Created = after - existing
	result.Updated = len(queue) - result.Created
	result.Imported = true

	return *result
}

// ---------------------------------------------------------------------------
// Package prices
// ---------------------------------------------------------------------------

func (s *ServiceRateCardImpl) FindPackagePrices(ctx context.Context, versionId int, request webRateCard.RateCardRequestFindAll) ([]webRateCard.PackagePriceResponse, int) {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	s.mustFindVersion(ctx, tx, versionId)

	list, err := s.RepositoryRateCardInterface.FindPackagePrices(ctx, tx, versionId,
		request.GetTake(), request.GetSkip(), request.GetSearch())
	helpers.PanicIfError(err)

	total, err := s.RepositoryRateCardInterface.CountPackagePrices(ctx, tx, versionId, request.GetSearch())
	helpers.PanicIfError(err)

	responses := make([]webRateCard.PackagePriceResponse, len(list))
	for i, item := range list {
		responses[i] = packagePriceToResponse(item)
	}

	return responses, total
}

func (s *ServiceRateCardImpl) UpsertPackagePrice(ctx context.Context, versionId int, request webRateCard.UpsertPackagePriceRequest) webRateCard.PackagePriceResponse {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	s.mustFindEditableVersion(ctx, tx, versionId)

	if _, err := s.RepositorySalesPackage.FindById(ctx, tx, request.SalesPackageId); err == sql.ErrNoRows {
		panic(exceptions.NewBadRequestError("sales package not found"))
	} else {
		helpers.PanicIfError(err)
	}

	helpers.PanicIfError(s.RepositoryRateCardInterface.UpsertPackagePrice(ctx, tx, models.RateCardPackagePrice{
		RateCardVersionId: versionId,
		SalesPackageId:    request.SalesPackageId,
		PriceIdrPer4Weeks: request.PriceIdrPer4Weeks,
	}))

	prices, err := s.RepositoryRateCardInterface.FindPackagePrices(ctx, tx, versionId, 100000, 0, "")
	helpers.PanicIfError(err)
	for _, price := range prices {
		if price.SalesPackageId == request.SalesPackageId {
			return packagePriceToResponse(price)
		}
	}

	return webRateCard.PackagePriceResponse{
		RateCardVersionId: versionId,
		SalesPackageId:    request.SalesPackageId,
		PriceIdrPer4Weeks: request.PriceIdrPer4Weeks,
	}
}

func (s *ServiceRateCardImpl) DeletePackagePrice(ctx context.Context, versionId int, packageId int) {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	s.mustFindEditableVersion(ctx, tx, versionId)
	helpers.PanicIfError(s.RepositoryRateCardInterface.DeletePackagePrice(ctx, tx, versionId, packageId))
}

func (s *ServiceRateCardImpl) PackagePriceTemplate() ([]byte, error) {
	return spreadsheets.BuildTemplate("Package Prices", PackagePriceColumns)
}

func (s *ServiceRateCardImpl) ExportPackagePrices(ctx context.Context, versionId int) ([]byte, error) {
	tx, err := s.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer helpers.CommitOrRollback(tx)

	list, err := s.RepositoryRateCardInterface.FindPackagePrices(ctx, tx, versionId, 100000, 0, "")
	if err != nil {
		return nil, err
	}

	rows := make([][]interface{}, len(list))
	for i, item := range list {
		rows[i] = []interface{}{item.SalesPackageName, item.PriceIdrPer4Weeks}
	}

	return spreadsheets.BuildExport("Package Prices", headersOf(PackagePriceColumns), rows)
}

func (s *ServiceRateCardImpl) ImportPackagePrices(ctx context.Context, versionId int, fileBytes []byte, fileType string) web.ImportResult {
	rows := s.parseUpload(fileBytes, fileType, PackagePriceColumns)

	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	s.mustFindEditableVersion(ctx, tx, versionId)

	colMap := spreadsheets.MapHeaderColumns(rows[0], PackagePriceColumns)
	result := web.NewImportResult()

	type pending struct {
		packageId int
		price     int64
	}
	var queue []pending
	seen := map[string]int{}

	for i, row := range rows[1:] {
		rowNumber := i + 2

		name := spreadsheets.ColValue(row, colMap, "package_name")
		rawPrice := spreadsheets.ColValue(row, colMap, "price")

		if name == "" && rawPrice == "" {
			continue
		}
		result.Rows++

		if name == "" {
			result.AddError(rowNumber, "Sales Package", "", "Sales Package is required")
			continue
		}
		if firstRow, duplicate := seen[name]; duplicate {
			result.AddError(rowNumber, "Sales Package", name,
				"This package is priced twice in this file (also on row "+strconv.Itoa(firstRow)+")")
			continue
		}
		seen[name] = rowNumber

		price, err := parsePrice(rawPrice)
		if err != nil {
			result.AddError(rowNumber, "Price per 4 Weeks (IDR)", rawPrice, err.Error())
			continue
		}

		matches, err := s.RepositorySalesPackage.FindByNames(ctx, tx, []string{name})
		helpers.PanicIfError(err)
		if len(matches) == 0 {
			result.AddError(rowNumber, "Sales Package", name, "No sales package with this name")
			continue
		}

		queue = append(queue, pending{packageId: matches[0].Id, price: price})
	}

	if result.HasErrors() {
		return *result
	}

	before, err := s.RepositoryRateCardInterface.CountPackagePrices(ctx, tx, versionId, "")
	helpers.PanicIfError(err)

	for _, item := range queue {
		helpers.PanicIfError(s.RepositoryRateCardInterface.UpsertPackagePrice(ctx, tx, models.RateCardPackagePrice{
			RateCardVersionId: versionId,
			SalesPackageId:    item.packageId,
			PriceIdrPer4Weeks: item.price,
		}))
	}

	after, err := s.RepositoryRateCardInterface.CountPackagePrices(ctx, tx, versionId, "")
	helpers.PanicIfError(err)

	result.Created = after - before
	result.Updated = len(queue) - result.Created
	result.Imported = true

	return *result
}

// ---------------------------------------------------------------------------
// Shared helpers
// ---------------------------------------------------------------------------

// parseUpload does the file-shape checks that happen before any transaction opens,
// so a malformed file never touches the database.
func (s *ServiceRateCardImpl) parseUpload(fileBytes []byte, fileType string, columns []spreadsheets.SheetColumn) [][]string {
	rows, err := spreadsheets.ParseSpreadsheet(fileBytes, fileType)
	if err == spreadsheets.ErrUnsupportedFileType {
		panic(exceptions.NewBadRequestError("Unsupported file type. Use xlsx or csv."))
	}
	helpers.PanicIfError(err)

	if len(rows) < 2 {
		panic(exceptions.NewBadRequestError("File must contain a header row and at least one data row."))
	}

	colMap := spreadsheets.MapHeaderColumns(rows[0], columns)
	if missing := spreadsheets.MissingRequiredColumns(colMap, columns); len(missing) > 0 {
		panic(exceptions.NewBadRequestError("Missing required column(s): " + strings.Join(missing, ", ")))
	}

	return rows
}

func (s *ServiceRateCardImpl) mustFindVersion(ctx context.Context, tx *sql.Tx, id int) models.RateCardVersion {
	found, err := s.RepositoryRateCardInterface.FindVersionById(ctx, tx, id)
	if err == sql.ErrNoRows {
		panic(exceptions.NewNotFoundError("rate card version not found"))
	}
	helpers.PanicIfError(err)

	return found
}

// mustFindEditableVersion is the single gate on every write. Publishing freezes a
// version for good, because approved quotations reference it and must never re-price.
func (s *ServiceRateCardImpl) mustFindEditableVersion(ctx context.Context, tx *sql.Tx, id int) models.RateCardVersion {
	found := s.mustFindVersion(ctx, tx, id)

	if !models.IsEditableRateCardStatus(found.Status) {
		panic(exceptions.NewBadRequestError(
			"this rate card is " + found.Status + " and can no longer be changed; create a new draft instead"))
	}

	return found
}

func (s *ServiceRateCardImpl) assertVersionCodeFree(ctx context.Context, tx *sql.Tx, code string, allowedId int) {
	existing, err := s.RepositoryRateCardInterface.FindVersionByCode(ctx, tx, code)
	if err == sql.ErrNoRows {
		return
	}
	helpers.PanicIfError(err)

	if existing.Id != allowedId {
		panic(exceptions.NewBadRequestError("that version code is already in use"))
	}
}

func (s *ServiceRateCardImpl) versionResponse(ctx context.Context, tx *sql.Tx, id int) webRateCard.VersionResponse {
	return versionToResponse(s.mustFindVersion(ctx, tx, id))
}

func headersOf(columns []spreadsheets.SheetColumn) []string {
	headers := make([]string, len(columns))
	for i, col := range columns {
		headers[i] = col.Header
	}

	return headers
}

// parsePrice accepts what people actually type: thousands separators, a trailing
// ".00" from Excel, and surrounding spaces. Rupiah has no minor unit, so a real
// fractional value is rejected rather than silently rounded.
func parsePrice(raw string) (int64, error) {
	cleaned := strings.NewReplacer(",", "", " ", "", "_", "").Replace(strings.TrimSpace(raw))
	if cleaned == "" {
		return 0, errors.New("Price is required")
	}

	if dot := strings.Index(cleaned, "."); dot >= 0 {
		fraction := strings.TrimRight(cleaned[dot+1:], "0")
		if fraction != "" {
			return 0, errors.New("Price must be a whole number of rupiah")
		}
		cleaned = cleaned[:dot]
	}

	price, err := strconv.ParseInt(cleaned, 10, 64)
	if err != nil {
		return 0, errors.New("Price must be a number")
	}
	if price < 0 {
		return 0, errors.New("Price cannot be negative")
	}

	return price, nil
}

func versionToResponse(v models.RateCardVersion) webRateCard.VersionResponse {
	return webRateCard.VersionResponse{
		Id: v.Id, VersionCode: v.VersionCode, Description: v.Description, Currency: v.Currency,
		Status: v.Status, IsEditable: models.IsEditableRateCardStatus(v.Status),
		PublishedByUserId: v.PublishedByUserId, PublishedByName: v.PublishedByName,
		PublishedAt: v.PublishedAt, BuildingPriceCount: v.BuildingPriceCount,
		PackagePriceCount: v.PackagePriceCount, CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt,
	}
}

func buildingPriceToResponse(p models.RateCardBuildingPrice) webRateCard.BuildingPriceResponse {
	return webRateCard.BuildingPriceResponse{
		Id: p.Id, RateCardVersionId: p.RateCardVersionId, BuildingId: p.BuildingId,
		BuildingName: p.BuildingName, BuildingIrisCode: p.BuildingIrisCode,
		BuildingType: p.BuildingType, Citytown: p.Citytown,
		PriceIdrPer4Weeks: p.PriceIdrPer4Weeks, CreatedAt: p.CreatedAt, UpdatedAt: p.UpdatedAt,
	}
}

func packagePriceToResponse(p models.RateCardPackagePrice) webRateCard.PackagePriceResponse {
	return webRateCard.PackagePriceResponse{
		Id: p.Id, RateCardVersionId: p.RateCardVersionId, SalesPackageId: p.SalesPackageId,
		SalesPackageName: p.SalesPackageName, PriceIdrPer4Weeks: p.PriceIdrPer4Weeks,
		BuildingCount: p.BuildingCount, CreatedAt: p.CreatedAt, UpdatedAt: p.UpdatedAt,
	}
}
