package salespackage

import (
	"context"
	"database/sql"
	"strconv"
	"strings"
	"time"

	"github.com/malikabdulaziz/tmn-backend/models"
)

type RepositorySalesPackageImpl struct{}

func NewRepositorySalesPackageImpl() RepositorySalesPackageInterface {
	return &RepositorySalesPackageImpl{}
}

// allowed order columns and directions to avoid SQL injection
var allowedOrderBy = map[string]bool{"id": true, "name": true, "created_at": true, "updated_at": true}
var allowedOrderDir = map[string]bool{"ASC": true, "DESC": true}

func safeOrder(orderBy, orderDirection string) (string, string) {
	if !allowedOrderBy[orderBy] {
		orderBy = "created_at"
	}
	if !allowedOrderDir[orderDirection] {
		orderDirection = "DESC"
	}
	return orderBy, orderDirection
}

// Create inserts a new sales package and its building links
func (r *RepositorySalesPackageImpl) Create(ctx context.Context, tx *sql.Tx, pkg models.SalesPackage, buildingIds []int) (models.SalesPackage, error) {
	SQL := `INSERT INTO ` + models.SalesPackageTable + ` (name) VALUES ($1) RETURNING id, created_at, updated_at`
	err := tx.QueryRowContext(ctx, SQL, pkg.Name).Scan(&pkg.Id, &pkg.CreatedAt, &pkg.UpdatedAt)
	if err != nil {
		return models.SalesPackage{}, err
	}
	for _, bid := range buildingIds {
		if err := r.CreateBuildingLink(ctx, tx, pkg.Id, bid); err != nil {
			return models.SalesPackage{}, err
		}
	}
	// Load building refs for response
	buildings, err := r.findBuildingRefsBySalesPackageId(ctx, tx, pkg.Id)
	if err != nil {
		return models.SalesPackage{}, err
	}
	pkg.Buildings = buildings
	return pkg, nil
}

// FindAll retrieves all sales packages with pagination and ordering; loads building refs per package
func (r *RepositorySalesPackageImpl) FindAll(ctx context.Context, tx *sql.Tx, take int, skip int, orderBy string, orderDirection string) ([]models.SalesPackage, error) {
	orderBy, orderDirection = safeOrder(orderBy, orderDirection)
	SQL := `SELECT id, name, created_at, updated_at FROM ` + models.SalesPackageTable + `
		ORDER BY ` + orderBy + ` ` + orderDirection + ` LIMIT $1 OFFSET $2`
	rows, err := tx.QueryContext(ctx, SQL, take, skip)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var packages []models.SalesPackage
	var ids []int
	for rows.Next() {
		var n models.NullAbleSalesPackage
		if err := rows.Scan(&n.Id, &n.Name, &n.CreatedAt, &n.UpdatedAt); err != nil {
			return nil, err
		}
		pkg := models.NullAbleSalesPackageToSalesPackage(n)
		ids = append(ids, pkg.Id)
		packages = append(packages, pkg)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return packages, nil
	}
	refsMap, err := r.findBuildingRefsBySalesPackageIds(ctx, tx, ids)
	if err != nil {
		return nil, err
	}
	for i := range packages {
		packages[i].Buildings = refsMap[packages[i].Id]
	}
	return packages, nil
}

// CountAll returns total count of sales packages
func (r *RepositorySalesPackageImpl) CountAll(ctx context.Context, tx *sql.Tx) (int, error) {
	SQL := `SELECT COUNT(*) FROM ` + models.SalesPackageTable
	var total int
	err := tx.QueryRowContext(ctx, SQL).Scan(&total)
	return total, err
}

// FindById retrieves a sales package by ID with its building refs
func (r *RepositorySalesPackageImpl) FindById(ctx context.Context, tx *sql.Tx, id int) (models.SalesPackage, error) {
	SQL := `SELECT id, name, created_at, updated_at FROM ` + models.SalesPackageTable + ` WHERE id = $1`
	row := tx.QueryRowContext(ctx, SQL, id)
	var n models.NullAbleSalesPackage
	if err := row.Scan(&n.Id, &n.Name, &n.CreatedAt, &n.UpdatedAt); err != nil {
		return models.SalesPackage{}, err
	}
	pkg := models.NullAbleSalesPackageToSalesPackage(n)
	buildings, err := r.findBuildingRefsBySalesPackageId(ctx, tx, pkg.Id)
	if err != nil {
		return models.SalesPackage{}, err
	}
	pkg.Buildings = buildings
	return pkg, nil
}

// Update updates name and replaces building links
func (r *RepositorySalesPackageImpl) Update(ctx context.Context, tx *sql.Tx, pkg models.SalesPackage, buildingIds []int) (models.SalesPackage, error) {
	SQL := `UPDATE ` + models.SalesPackageTable + ` SET name = $1, updated_at = $2 WHERE id = $3 RETURNING updated_at`
	err := tx.QueryRowContext(ctx, SQL, pkg.Name, time.Now(), pkg.Id).Scan(&pkg.UpdatedAt)
	if err != nil {
		return models.SalesPackage{}, err
	}
	if err := r.DeleteBuildingLinksBySalesPackageId(ctx, tx, pkg.Id); err != nil {
		return models.SalesPackage{}, err
	}
	for _, bid := range buildingIds {
		if err := r.CreateBuildingLink(ctx, tx, pkg.Id, bid); err != nil {
			return models.SalesPackage{}, err
		}
	}
	buildings, err := r.findBuildingRefsBySalesPackageId(ctx, tx, pkg.Id)
	if err != nil {
		return models.SalesPackage{}, err
	}
	pkg.Buildings = buildings
	return pkg, nil
}

// DeleteBuildingLinksBySalesPackageId removes all building links for a package
func (r *RepositorySalesPackageImpl) DeleteBuildingLinksBySalesPackageId(ctx context.Context, tx *sql.Tx, salesPackageId int) error {
	SQL := `DELETE FROM ` + models.SalesPackageBuildingTable + ` WHERE sales_package_id = $1`
	_, err := tx.ExecContext(ctx, SQL, salesPackageId)
	return err
}

// CreateBuildingLink inserts one sales_package_id, building_id row
func (r *RepositorySalesPackageImpl) CreateBuildingLink(ctx context.Context, tx *sql.Tx, salesPackageId int, buildingId int) error {
	SQL := `INSERT INTO ` + models.SalesPackageBuildingTable + ` (sales_package_id, building_id) VALUES ($1, $2)`
	_, err := tx.ExecContext(ctx, SQL, salesPackageId, buildingId)
	return err
}

// Delete deletes a sales package (CASCADE removes junction rows)
func (r *RepositorySalesPackageImpl) Delete(ctx context.Context, tx *sql.Tx, id int) error {
	SQL := `DELETE FROM ` + models.SalesPackageTable + ` WHERE id = $1`
	_, err := tx.ExecContext(ctx, SQL, id)
	return err
}

// findBuildingRefsBySalesPackageId returns building id+name for a package
func (r *RepositorySalesPackageImpl) findBuildingRefsBySalesPackageId(ctx context.Context, tx *sql.Tx, salesPackageId int) ([]models.BuildingRef, error) {
	SQL := `SELECT b.id, b.name FROM ` + models.BuildingTable + ` b
		INNER JOIN ` + models.SalesPackageBuildingTable + ` spb ON spb.building_id = b.id
		WHERE spb.sales_package_id = $1 ORDER BY b.name`
	rows, err := tx.QueryContext(ctx, SQL, salesPackageId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var refs []models.BuildingRef
	for rows.Next() {
		var ref models.BuildingRef
		if err := rows.Scan(&ref.Id, &ref.Name); err != nil {
			return nil, err
		}
		refs = append(refs, ref)
	}
	return refs, rows.Err()
}

// findBuildingRefsBySalesPackageIds returns map[salesPackageId][]BuildingRef
func (r *RepositorySalesPackageImpl) findBuildingRefsBySalesPackageIds(ctx context.Context, tx *sql.Tx, salesPackageIds []int) (map[int][]models.BuildingRef, error) {
	if len(salesPackageIds) == 0 {
		return make(map[int][]models.BuildingRef), nil
	}
	placeholders := make([]string, len(salesPackageIds))
	args := make([]interface{}, len(salesPackageIds))
	for i, id := range salesPackageIds {
		placeholders[i] = "$" + strconv.Itoa(i+1)
		args[i] = id
	}
	SQL := `SELECT spb.sales_package_id, b.id, b.name FROM ` + models.BuildingTable + ` b
		INNER JOIN ` + models.SalesPackageBuildingTable + ` spb ON spb.building_id = b.id
		WHERE spb.sales_package_id IN (` + strings.Join(placeholders, ",") + `) ORDER BY spb.sales_package_id, b.name`
	rows, err := tx.QueryContext(ctx, SQL, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make(map[int][]models.BuildingRef)
	for rows.Next() {
		var pkgId int
		var ref models.BuildingRef
		if err := rows.Scan(&pkgId, &ref.Id, &ref.Name); err != nil {
			return nil, err
		}
		out[pkgId] = append(out[pkgId], ref)
	}
	return out, rows.Err()
}
