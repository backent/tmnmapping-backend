package buildingprice

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
	repositoriesBuildingPrice "github.com/malikabdulaziz/tmn-backend/repositories/buildingprice"
	"github.com/malikabdulaziz/tmn-backend/spreadsheets"
	"github.com/malikabdulaziz/tmn-backend/web"
	webBuildingPrice "github.com/malikabdulaziz/tmn-backend/web/buildingprice"
)

type ServiceBuildingPriceImpl struct {
	DB                      *sql.DB
	RepositoryBuildingPrice repositoriesBuildingPrice.RepositoryBuildingPriceInterface
	RepositoryBuilding      repositoriesBuilding.RepositoryBuildingInterface
}

func NewServiceBuildingPriceImpl(
	db *sql.DB,
	repoPrice repositoriesBuildingPrice.RepositoryBuildingPriceInterface,
	repoBuilding repositoriesBuilding.RepositoryBuildingInterface,
) ServiceBuildingPriceInterface {
	return &ServiceBuildingPriceImpl{
		DB:                      db,
		RepositoryBuildingPrice: repoPrice,
		RepositoryBuilding:      repoBuilding,
	}
}

func (s *ServiceBuildingPriceImpl) FindAll(ctx context.Context, request webBuildingPrice.BuildingPriceRequestFindAll) ([]webBuildingPrice.BuildingPriceResponse, int) {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	list, err := s.RepositoryBuildingPrice.FindAll(ctx, tx, request.GetTake(), request.GetSkip(), request.GetSearch())
	helpers.PanicIfError(err)

	total, err := s.RepositoryBuildingPrice.CountAll(ctx, tx, request.GetSearch())
	helpers.PanicIfError(err)

	responses := make([]webBuildingPrice.BuildingPriceResponse, len(list))
	for i, item := range list {
		responses[i] = toResponse(item)
	}

	return responses, total
}

func (s *ServiceBuildingPriceImpl) Upsert(ctx context.Context, request webBuildingPrice.UpsertBuildingPriceRequest) webBuildingPrice.BuildingPriceResponse {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	if _, err := s.RepositoryBuilding.FindById(ctx, tx, request.BuildingId); err == sql.ErrNoRows {
		panic(exceptions.NewBadRequestError("building not found"))
	} else {
		helpers.PanicIfError(err)
	}

	helpers.PanicIfError(s.RepositoryBuildingPrice.Upsert(ctx, tx, request.BuildingId, request.PriceIdrPerWeek))

	saved, err := s.RepositoryBuildingPrice.FindByBuildingId(ctx, tx, request.BuildingId)
	helpers.PanicIfError(err)

	return toResponse(saved)
}

func (s *ServiceBuildingPriceImpl) Delete(ctx context.Context, buildingId int) {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	if _, err := s.RepositoryBuildingPrice.FindByBuildingId(ctx, tx, buildingId); err == sql.ErrNoRows {
		panic(exceptions.NewNotFoundError("that building has no price"))
	} else {
		helpers.PanicIfError(err)
	}

	helpers.PanicIfError(s.RepositoryBuildingPrice.Delete(ctx, tx, buildingId))
}

func (s *ServiceBuildingPriceImpl) Import(ctx context.Context, fileBytes []byte, fileType string, dryRun bool) web.ImportResult {
	rows := parseUpload(fileBytes, fileType)

	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	colMap := spreadsheets.MapHeaderColumns(rows[0], BuildingPriceColumns)
	result := web.NewImportResult()

	existing, err := s.RepositoryBuildingPrice.FindAllPrices(ctx, tx)
	helpers.PanicIfError(err)

	type pending struct {
		buildingId int
		price      int64
	}
	var queue []pending
	seen := map[string]int{}

	for i, row := range rows[1:] {
		rowNumber := i + 2

		irisCode := spreadsheets.ColValue(row, colMap, "iris_building_id")
		rawPrice := spreadsheets.ColValue(row, colMap, "price")

		if irisCode == "" && rawPrice == "" {
			continue
		}

		// The source workbook repeats its header row as a section break, so a row
		// whose key literally reads like the header is layout, not data.
		if strings.EqualFold(irisCode, "IRIS Building ID") {
			continue
		}

		result.Rows++

		if irisCode == "" {
			result.AddError(rowNumber, "IRIS Building ID", "", "IRIS Building ID is required")
			continue
		}
		if firstRow, duplicate := seen[irisCode]; duplicate {
			result.AddError(rowNumber, "IRIS Building ID", irisCode,
				"This building is priced twice in this file (also on row "+strconv.Itoa(firstRow)+")")
			continue
		}
		seen[irisCode] = rowNumber

		price, err := parsePrice(rawPrice)
		if err != nil {
			result.AddError(rowNumber, "Price per Week (IDR)", rawPrice, err.Error())
			continue
		}

		// A zero price means the building has no screens installed yet. Importing it
		// would offer that building for nothing.
		if price == 0 {
			result.Skipped++
			continue
		}

		building, err := s.RepositoryBuilding.FindByIrisCode(ctx, tx, irisCode)
		if err == sql.ErrNoRows {
			result.AddError(rowNumber, "IRIS Building ID", irisCode,
				"No building with this IRIS code. Export the current list to get valid values.")
			continue
		}
		helpers.PanicIfError(err)

		queue = append(queue, pending{buildingId: building.Id, price: price})
	}

	// A bad row is reported and left out; the valid rows still apply. The preview
	// lists every rejected row and why before anything is written, so applying a
	// partly bad file is a decision the operator makes knowingly. Refusing it
	// outright meant the business's own rate card workbook -- 16 unusable rows out of
	// 1,624 -- could never be uploaded as it is.
	var changes []pending
	for _, item := range queue {
		current, had := existing[item.buildingId]
		switch {
		case !had:
			result.Created++
			changes = append(changes, item)
		case current != item.price:
			result.Updated++
			changes = append(changes, item)
		default:
			result.Unchanged++
		}
	}

	if dryRun {
		result.DryRun = true
		return *result
	}

	for _, item := range changes {
		helpers.PanicIfError(s.RepositoryBuildingPrice.Upsert(ctx, tx, item.buildingId, item.price))
	}
	result.Imported = true

	return *result
}

func (s *ServiceBuildingPriceImpl) Export(ctx context.Context) ([]byte, error) {
	tx, err := s.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer helpers.CommitOrRollback(tx)

	list, err := s.RepositoryBuildingPrice.FindAll(ctx, tx, 100000, 0, "")
	if err != nil {
		return nil, err
	}

	headers := make([]string, len(BuildingPriceColumns))
	for i, col := range BuildingPriceColumns {
		headers[i] = col.Header
	}

	rows := make([][]interface{}, len(list))
	for i, item := range list {
		// Emits the key the importer reads, so an exported file can be edited and
		// uploaded straight back.
		rows[i] = []interface{}{item.BuildingIrisCode, item.BuildingName, item.PriceIdrPerWeek}
	}

	return spreadsheets.BuildExport("Building Prices", headers, rows)
}

func (s *ServiceBuildingPriceImpl) Template() ([]byte, error) {
	return spreadsheets.BuildTemplate("Building Prices", BuildingPriceColumns)
}

func parseUpload(fileBytes []byte, fileType string) [][]string {
	rows, err := spreadsheets.ParseSpreadsheet(fileBytes, fileType)
	if err == spreadsheets.ErrUnsupportedFileType {
		panic(exceptions.NewBadRequestError("Unsupported file type. Use xlsx or csv."))
	}
	helpers.PanicIfError(err)

	if len(rows) < 2 {
		panic(exceptions.NewBadRequestError("File must contain a header row and at least one data row."))
	}

	rows[0] = applyHeaderAliases(rows[0])

	colMap := spreadsheets.MapHeaderColumns(rows[0], BuildingPriceColumns)
	if missing := spreadsheets.MissingRequiredColumns(colMap, BuildingPriceColumns); len(missing) > 0 {
		panic(exceptions.NewBadRequestError("Missing required column(s): " + strings.Join(missing, ", ")))
	}

	return rows
}

// applyHeaderAliases renames a recognised alias to the canonical price header, but
// only when the canonical one is absent: a sheet that has both keeps its own.
func applyHeaderAliases(header []string) []string {
	canonical := priceHeader()
	for _, h := range header {
		if strings.EqualFold(strings.TrimSpace(h), canonical) {
			return header
		}
	}

	out := make([]string, len(header))
	copy(out, header)
	for i, h := range out {
		for _, alias := range priceHeaderAliases {
			if strings.EqualFold(strings.TrimSpace(h), alias) {
				out[i] = canonical
				return out
			}
		}
	}

	return out
}

func priceHeader() string {
	for _, col := range BuildingPriceColumns {
		if col.Key == "price" {
			return col.Header
		}
	}

	return "Price per Week (IDR)"
}

// parsePrice reads a whole-rupiah amount the ways spreadsheets present one:
// "1100000", "1,100,000", "1.100.000", "Rp 1.100.000" or "1100000.00".
//
// A dot followed by one or two digits is a decimal rendering of a whole number and
// must be zeros; any other dot is an Indonesian thousands separator.
func parsePrice(raw string) (int64, error) {
	s := strings.TrimSpace(raw)
	s = strings.TrimPrefix(strings.TrimPrefix(s, "Rp"), "IDR")
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, ",", "")

	if i := strings.LastIndex(s, "."); i >= 0 && len(s)-i-1 <= 2 && !strings.Contains(s[:i], ".") {
		if strings.Trim(s[i+1:], "0") != "" {
			return 0, errors.New("price must be whole rupiah")
		}
		s = s[:i]
	} else {
		s = strings.ReplaceAll(s, ".", "")
	}

	if s == "" {
		return 0, errors.New("price is required")
	}

	value, err := strconv.ParseInt(s, 10, 64)
	if err != nil || value < 0 {
		return 0, errors.New("price must be a whole, non-negative number of rupiah")
	}

	return value, nil
}

func toResponse(item models.BuildingPrice) webBuildingPrice.BuildingPriceResponse {
	return webBuildingPrice.BuildingPriceResponse{
		Id:               item.Id,
		BuildingId:       item.BuildingId,
		BuildingName:     item.BuildingName,
		BuildingIrisCode: item.BuildingIrisCode,
		BuildingType:     item.BuildingType,
		Citytown:         item.Citytown,
		PriceIdrPerWeek:  item.PriceIdrPerWeek,
		CreatedAt:        item.CreatedAt,
		UpdatedAt:        item.UpdatedAt,
	}
}
