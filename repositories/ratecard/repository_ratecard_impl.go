package ratecard

import (
	"context"
	"database/sql"
	"time"

	"github.com/malikabdulaziz/tmn-backend/models"
)

type RepositoryRateCardImpl struct{}

func NewRepositoryRateCardImpl() RepositoryRateCardInterface {
	return &RepositoryRateCardImpl{}
}

// Versions carry their price counts so the list page can show how full a draft is
// without a second round trip per row.
var versionSelect = `SELECT v.id, v.version_code, v.description, v.currency, v.status,
	v.published_by_user_id, u.name, v.published_at,
	(SELECT COUNT(*) FROM ` + models.RateCardBuildingPriceTable + ` bp WHERE bp.rate_card_version_id = v.id),
	(SELECT COUNT(*) FROM ` + models.RateCardPackagePriceTable + ` pp WHERE pp.rate_card_version_id = v.id),
	v.created_at, v.updated_at
	FROM ` + models.RateCardVersionTable + ` v
	LEFT JOIN ` + models.UserTable + ` u ON u.id = v.published_by_user_id`

var versionAllowedOrderBy = map[string]string{
	"id": "v.id", "version_code": "v.version_code", "status": "v.status",
	"published_at": "v.published_at", "created_at": "v.created_at", "updated_at": "v.updated_at",
}

func versionSafeOrder(orderBy, orderDirection string) (string, string) {
	column, ok := versionAllowedOrderBy[orderBy]
	if !ok {
		column = "v.created_at"
	}
	if orderDirection != "ASC" && orderDirection != "DESC" {
		orderDirection = "DESC"
	}

	return column, orderDirection
}

func scanVersion(rows *sql.Rows) (models.RateCardVersion, error) {
	var n models.NullAbleRateCardVersion
	err := rows.Scan(&n.Id, &n.VersionCode, &n.Description, &n.Currency, &n.Status,
		&n.PublishedByUserId, &n.PublishedByName, &n.PublishedAt,
		&n.BuildingPriceCount, &n.PackagePriceCount, &n.CreatedAt, &n.UpdatedAt)
	if err != nil {
		return models.RateCardVersion{}, err
	}

	return models.NullAbleRateCardVersionToRateCardVersion(n), nil
}

func (r *RepositoryRateCardImpl) CreateVersion(ctx context.Context, tx *sql.Tx, v models.RateCardVersion) (models.RateCardVersion, error) {
	SQL := "INSERT INTO " + models.RateCardVersionTable +
		" (version_code, description, currency, status) VALUES ($1, $2, $3, $4)" +
		" RETURNING id, created_at, updated_at"
	err := tx.QueryRowContext(ctx, SQL, v.VersionCode, v.Description, v.Currency, v.Status).
		Scan(&v.Id, &v.CreatedAt, &v.UpdatedAt)

	return v, err
}

func (r *RepositoryRateCardImpl) FindAllVersions(ctx context.Context, tx *sql.Tx, take int, skip int, orderBy string, orderDirection string, search string) ([]models.RateCardVersion, error) {
	column, direction := versionSafeOrder(orderBy, orderDirection)

	var rows *sql.Rows
	var err error

	if search != "" {
		SQL := versionSelect + " WHERE v.version_code ILIKE $1 OR v.description ILIKE $1" +
			" ORDER BY " + column + " " + direction + " LIMIT $2 OFFSET $3"
		rows, err = tx.QueryContext(ctx, SQL, "%"+search+"%", take, skip)
	} else {
		SQL := versionSelect + " ORDER BY " + column + " " + direction + " LIMIT $1 OFFSET $2"
		rows, err = tx.QueryContext(ctx, SQL, take, skip)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.RateCardVersion
	for rows.Next() {
		item, err := scanVersion(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, item)
	}

	return list, rows.Err()
}

func (r *RepositoryRateCardImpl) CountAllVersions(ctx context.Context, tx *sql.Tx, search string) (int, error) {
	var total int
	var err error

	if search != "" {
		SQL := "SELECT COUNT(*) FROM " + models.RateCardVersionTable +
			" WHERE version_code ILIKE $1 OR description ILIKE $1"
		err = tx.QueryRowContext(ctx, SQL, "%"+search+"%").Scan(&total)
	} else {
		err = tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM "+models.RateCardVersionTable).Scan(&total)
	}

	return total, err
}

func (r *RepositoryRateCardImpl) FindVersionById(ctx context.Context, tx *sql.Tx, id int) (models.RateCardVersion, error) {
	return r.findVersion(ctx, tx, " WHERE v.id = $1", id)
}

func (r *RepositoryRateCardImpl) FindVersionByCode(ctx context.Context, tx *sql.Tx, code string) (models.RateCardVersion, error) {
	return r.findVersion(ctx, tx, " WHERE v.version_code = $1", code)
}

func (r *RepositoryRateCardImpl) FindCurrentVersion(ctx context.Context, tx *sql.Tx) (models.RateCardVersion, error) {
	return r.findVersion(ctx, tx, " WHERE v.status = $1", models.RateCardStatusCurrent)
}

func (r *RepositoryRateCardImpl) findVersion(ctx context.Context, tx *sql.Tx, where string, arg interface{}) (models.RateCardVersion, error) {
	rows, err := tx.QueryContext(ctx, versionSelect+where, arg)
	if err != nil {
		return models.RateCardVersion{}, err
	}
	defer rows.Close()

	if rows.Next() {
		return scanVersion(rows)
	}

	return models.RateCardVersion{}, sql.ErrNoRows
}

func (r *RepositoryRateCardImpl) UpdateVersion(ctx context.Context, tx *sql.Tx, v models.RateCardVersion) (models.RateCardVersion, error) {
	SQL := "UPDATE " + models.RateCardVersionTable +
		" SET version_code = $1, description = $2, currency = $3, updated_at = $4" +
		" WHERE id = $5 RETURNING updated_at"
	err := tx.QueryRowContext(ctx, SQL, v.VersionCode, v.Description, v.Currency, time.Now(), v.Id).
		Scan(&v.UpdatedAt)

	return v, err
}

func (r *RepositoryRateCardImpl) DeleteVersion(ctx context.Context, tx *sql.Tx, id int) error {
	_, err := tx.ExecContext(ctx, "DELETE FROM "+models.RateCardVersionTable+" WHERE id = $1", id)

	return err
}

// SetVersionStatus records who published and when. published_at is set on the way
// out of draft and never cleared, which the table's CHECK constraint also enforces.
func (r *RepositoryRateCardImpl) SetVersionStatus(ctx context.Context, tx *sql.Tx, id int, status string, publishedByUserId int) error {
	SQL := "UPDATE " + models.RateCardVersionTable +
		" SET status = $1, published_by_user_id = COALESCE($2, published_by_user_id)," +
		" published_at = COALESCE(published_at, CURRENT_TIMESTAMP), updated_at = $3 WHERE id = $4"

	var publisher interface{}
	if publishedByUserId > 0 {
		publisher = publishedByUserId
	}

	_, err := tx.ExecContext(ctx, SQL, status, publisher, time.Now(), id)

	return err
}

// DemoteCurrentVersion moves whatever is current to historical. Publishing calls this
// first, inside the same transaction, so the partial unique index never sees two.
func (r *RepositoryRateCardImpl) DemoteCurrentVersion(ctx context.Context, tx *sql.Tx) error {
	SQL := "UPDATE " + models.RateCardVersionTable +
		" SET status = $1, updated_at = $2 WHERE status = $3"
	_, err := tx.ExecContext(ctx, SQL, models.RateCardStatusHistorical, time.Now(), models.RateCardStatusCurrent)

	return err
}

// ---------------------------------------------------------------------------
// Building prices
// ---------------------------------------------------------------------------

var buildingPriceSelect = `SELECT p.id, p.rate_card_version_id, p.building_id, b.name,
	b.iris_code, b.building_type, b.citytown, p.price_idr_per_week, p.created_at, p.updated_at
	FROM ` + models.RateCardBuildingPriceTable + ` p
	JOIN ` + models.BuildingTable + ` b ON b.id = p.building_id`

func scanBuildingPrice(rows *sql.Rows) (models.RateCardBuildingPrice, error) {
	var n models.NullAbleRateCardBuildingPrice
	err := rows.Scan(&n.Id, &n.RateCardVersionId, &n.BuildingId, &n.BuildingName,
		&n.BuildingIrisCode, &n.BuildingType, &n.Citytown, &n.PriceIdrPerWeek,
		&n.CreatedAt, &n.UpdatedAt)
	if err != nil {
		return models.RateCardBuildingPrice{}, err
	}

	return models.NullAbleRateCardBuildingPriceToRateCardBuildingPrice(n), nil
}

func (r *RepositoryRateCardImpl) FindBuildingPrices(ctx context.Context, tx *sql.Tx, versionId int, take int, skip int, search string) ([]models.RateCardBuildingPrice, error) {
	var rows *sql.Rows
	var err error

	if search != "" {
		SQL := buildingPriceSelect + " WHERE p.rate_card_version_id = $1" +
			" AND (b.name ILIKE $2 OR b.iris_code ILIKE $2 OR b.citytown ILIKE $2)" +
			" ORDER BY b.name ASC LIMIT $3 OFFSET $4"
		rows, err = tx.QueryContext(ctx, SQL, versionId, "%"+search+"%", take, skip)
	} else {
		SQL := buildingPriceSelect + " WHERE p.rate_card_version_id = $1" +
			" ORDER BY b.name ASC LIMIT $2 OFFSET $3"
		rows, err = tx.QueryContext(ctx, SQL, versionId, take, skip)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.RateCardBuildingPrice
	for rows.Next() {
		item, err := scanBuildingPrice(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, item)
	}

	return list, rows.Err()
}

func (r *RepositoryRateCardImpl) CountBuildingPrices(ctx context.Context, tx *sql.Tx, versionId int, search string) (int, error) {
	base := "SELECT COUNT(*) FROM " + models.RateCardBuildingPriceTable + " p JOIN " +
		models.BuildingTable + " b ON b.id = p.building_id WHERE p.rate_card_version_id = $1"

	var total int
	var err error
	if search != "" {
		err = tx.QueryRowContext(ctx,
			base+" AND (b.name ILIKE $2 OR b.iris_code ILIKE $2 OR b.citytown ILIKE $2)",
			versionId, "%"+search+"%").Scan(&total)
	} else {
		err = tx.QueryRowContext(ctx, base, versionId).Scan(&total)
	}

	return total, err
}

// UpsertBuildingPrice lets an upload and a hand edit share one path: re-pricing a
// building already in the draft updates it rather than failing on the unique key.
func (r *RepositoryRateCardImpl) UpsertBuildingPrice(ctx context.Context, tx *sql.Tx, p models.RateCardBuildingPrice) error {
	SQL := "INSERT INTO " + models.RateCardBuildingPriceTable +
		" (rate_card_version_id, building_id, price_idr_per_week) VALUES ($1, $2, $3)" +
		" ON CONFLICT (rate_card_version_id, building_id)" +
		" DO UPDATE SET price_idr_per_week = EXCLUDED.price_idr_per_week, updated_at = CURRENT_TIMESTAMP"
	_, err := tx.ExecContext(ctx, SQL, p.RateCardVersionId, p.BuildingId, p.PriceIdrPerWeek)

	return err
}

func (r *RepositoryRateCardImpl) DeleteBuildingPrice(ctx context.Context, tx *sql.Tx, versionId int, buildingId int) error {
	_, err := tx.ExecContext(ctx, "DELETE FROM "+models.RateCardBuildingPriceTable+
		" WHERE rate_card_version_id = $1 AND building_id = $2", versionId, buildingId)

	return err
}

// ---------------------------------------------------------------------------
// Package prices
// ---------------------------------------------------------------------------

var packagePriceSelect = `SELECT p.id, p.rate_card_version_id, p.sales_package_id, sp.name,
	p.price_idr_per_week,
	(SELECT COUNT(*) FROM ` + models.RateCardPackageBuildingTable + ` pb
	  WHERE pb.rate_card_version_id = p.rate_card_version_id AND pb.sales_package_id = p.sales_package_id),
	p.created_at, p.updated_at
	FROM ` + models.RateCardPackagePriceTable + ` p
	JOIN ` + models.SalesPackageTable + ` sp ON sp.id = p.sales_package_id`

func scanPackagePrice(rows *sql.Rows) (models.RateCardPackagePrice, error) {
	var n models.NullAbleRateCardPackagePrice
	err := rows.Scan(&n.Id, &n.RateCardVersionId, &n.SalesPackageId, &n.SalesPackageName,
		&n.PriceIdrPerWeek, &n.BuildingCount, &n.CreatedAt, &n.UpdatedAt)
	if err != nil {
		return models.RateCardPackagePrice{}, err
	}

	return models.NullAbleRateCardPackagePriceToRateCardPackagePrice(n), nil
}

func (r *RepositoryRateCardImpl) FindPackagePrices(ctx context.Context, tx *sql.Tx, versionId int, take int, skip int, search string) ([]models.RateCardPackagePrice, error) {
	var rows *sql.Rows
	var err error

	if search != "" {
		SQL := packagePriceSelect + " WHERE p.rate_card_version_id = $1 AND sp.name ILIKE $2" +
			" ORDER BY sp.name ASC LIMIT $3 OFFSET $4"
		rows, err = tx.QueryContext(ctx, SQL, versionId, "%"+search+"%", take, skip)
	} else {
		SQL := packagePriceSelect + " WHERE p.rate_card_version_id = $1" +
			" ORDER BY sp.name ASC LIMIT $2 OFFSET $3"
		rows, err = tx.QueryContext(ctx, SQL, versionId, take, skip)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.RateCardPackagePrice
	for rows.Next() {
		item, err := scanPackagePrice(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, item)
	}

	return list, rows.Err()
}

func (r *RepositoryRateCardImpl) CountPackagePrices(ctx context.Context, tx *sql.Tx, versionId int, search string) (int, error) {
	base := "SELECT COUNT(*) FROM " + models.RateCardPackagePriceTable + " p JOIN " +
		models.SalesPackageTable + " sp ON sp.id = p.sales_package_id WHERE p.rate_card_version_id = $1"

	var total int
	var err error
	if search != "" {
		err = tx.QueryRowContext(ctx, base+" AND sp.name ILIKE $2", versionId, "%"+search+"%").Scan(&total)
	} else {
		err = tx.QueryRowContext(ctx, base, versionId).Scan(&total)
	}

	return total, err
}

func (r *RepositoryRateCardImpl) UpsertPackagePrice(ctx context.Context, tx *sql.Tx, p models.RateCardPackagePrice) error {
	SQL := "INSERT INTO " + models.RateCardPackagePriceTable +
		" (rate_card_version_id, sales_package_id, price_idr_per_week) VALUES ($1, $2, $3)" +
		" ON CONFLICT (rate_card_version_id, sales_package_id)" +
		" DO UPDATE SET price_idr_per_week = EXCLUDED.price_idr_per_week, updated_at = CURRENT_TIMESTAMP"
	_, err := tx.ExecContext(ctx, SQL, p.RateCardVersionId, p.SalesPackageId, p.PriceIdrPerWeek)

	return err
}

func (r *RepositoryRateCardImpl) DeletePackagePrice(ctx context.Context, tx *sql.Tx, versionId int, packageId int) error {
	_, err := tx.ExecContext(ctx, "DELETE FROM "+models.RateCardPackagePriceTable+
		" WHERE rate_card_version_id = $1 AND sales_package_id = $2", versionId, packageId)

	return err
}

// ReplacePackageBuildings rewrites a package's frozen composition wholesale. Diffing
// would leave stale rows behind when a building is dropped from the package.
func (r *RepositoryRateCardImpl) ReplacePackageBuildings(ctx context.Context, tx *sql.Tx, versionId int, packageId int, buildingIds []int) error {
	_, err := tx.ExecContext(ctx, "DELETE FROM "+models.RateCardPackageBuildingTable+
		" WHERE rate_card_version_id = $1 AND sales_package_id = $2", versionId, packageId)
	if err != nil {
		return err
	}

	for _, buildingId := range buildingIds {
		_, err = tx.ExecContext(ctx, "INSERT INTO "+models.RateCardPackageBuildingTable+
			" (rate_card_version_id, sales_package_id, building_id) VALUES ($1, $2, $3)",
			versionId, packageId, buildingId)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *RepositoryRateCardImpl) FindPackageBuildings(ctx context.Context, tx *sql.Tx, versionId int, packageId int) ([]models.RateCardPackageBuilding, error) {
	SQL := "SELECT pb.id, pb.rate_card_version_id, pb.sales_package_id, pb.building_id, b.name, pb.created_at" +
		" FROM " + models.RateCardPackageBuildingTable + " pb" +
		" JOIN " + models.BuildingTable + " b ON b.id = pb.building_id" +
		" WHERE pb.rate_card_version_id = $1 AND pb.sales_package_id = $2 ORDER BY b.name ASC"

	rows, err := tx.QueryContext(ctx, SQL, versionId, packageId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.RateCardPackageBuilding
	for rows.Next() {
		var item models.RateCardPackageBuilding
		var name sql.NullString
		if err := rows.Scan(&item.Id, &item.RateCardVersionId, &item.SalesPackageId,
			&item.BuildingId, &name, &item.CreatedAt); err != nil {
			return nil, err
		}
		item.BuildingName = name.String
		list = append(list, item)
	}

	return list, rows.Err()
}

// CopyPrices seeds a new draft from an existing version, so a yearly re-price starts
// from last year's numbers instead of a blank sheet.
func (r *RepositoryRateCardImpl) CopyPrices(ctx context.Context, tx *sql.Tx, fromVersionId int, toVersionId int) error {
	statements := []string{
		"INSERT INTO " + models.RateCardBuildingPriceTable +
			" (rate_card_version_id, building_id, price_idr_per_week)" +
			" SELECT $1, building_id, price_idr_per_week FROM " + models.RateCardBuildingPriceTable +
			" WHERE rate_card_version_id = $2",
		"INSERT INTO " + models.RateCardPackagePriceTable +
			" (rate_card_version_id, sales_package_id, price_idr_per_week)" +
			" SELECT $1, sales_package_id, price_idr_per_week FROM " + models.RateCardPackagePriceTable +
			" WHERE rate_card_version_id = $2",
		"INSERT INTO " + models.RateCardPackageBuildingTable +
			" (rate_card_version_id, sales_package_id, building_id)" +
			" SELECT $1, sales_package_id, building_id FROM " + models.RateCardPackageBuildingTable +
			" WHERE rate_card_version_id = $2",
	}

	for _, SQL := range statements {
		if _, err := tx.ExecContext(ctx, SQL, toVersionId, fromVersionId); err != nil {
			return err
		}
	}

	return nil
}

func (r *RepositoryRateCardImpl) FindLivePackageBuildingIds(ctx context.Context, tx *sql.Tx, packageId int) ([]int, error) {
	rows, err := tx.QueryContext(ctx,
		"SELECT building_id FROM sales_package_buildings WHERE sales_package_id = $1 ORDER BY building_id",
		packageId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}

	return ids, rows.Err()
}
