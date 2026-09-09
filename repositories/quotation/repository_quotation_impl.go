package quotation

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/malikabdulaziz/tmn-backend/models"
)

type RepositoryQuotationImpl struct{}

func NewRepositoryQuotationImpl() RepositoryQuotationInterface {
	return &RepositoryQuotationImpl{}
}

// Every read resolves the people and the customer, because no quotation screen is
// useful without them.
var quotationSelect = `SELECT q.id, q.quote_number,
	q.sales_user_id, su.username, su.name,
	q.created_by_user_id, cu.name,
	q.customer_id, c.code, c.name,
	q.brand_id, b.code, b.name,
	q.rate_card_version_id, rc.version_code,
	q.attention_to, q.job_title, q.contact_phone, q.contact_email,
	q.campaign_year, q.valid_until,
	q.discount, q.tax_rate, q.status, q.required_approver_user_id, au.name, q.version,
	q.placement_gross, q.placement_discount_amount, q.placement_net,
	q.bonus_gross, q.bonus_net, q.total_gross, q.total_net,
	q.effective_discount_amount, q.effective_discount_rate,
	q.tax, q.total_including_tax,
	q.created_at, q.updated_at, q.approved_at
	FROM ` + models.QuotationTable + ` q
	JOIN ` + models.UserTable + ` su ON su.id = q.sales_user_id
	JOIN ` + models.UserTable + ` cu ON cu.id = q.created_by_user_id
	JOIN ` + models.CustomerTable + ` c ON c.id = q.customer_id
	JOIN ` + models.BrandTable + ` b ON b.id = q.brand_id
	LEFT JOIN ` + models.RateCardVersionTable + ` rc ON rc.id = q.rate_card_version_id
	LEFT JOIN ` + models.UserTable + ` au ON au.id = q.required_approver_user_id`

var quotationAllowedOrderBy = map[string]string{
	"id": "q.id", "quote_number": "q.quote_number", "status": "q.status",
	"created_at": "q.created_at", "updated_at": "q.updated_at",
	"total_net": "q.total_net", "customer_name": "c.name",
}

func safeOrder(orderBy, orderDirection string) (string, string) {
	column, ok := quotationAllowedOrderBy[orderBy]
	if !ok {
		column = "q.created_at"
	}
	if orderDirection != "ASC" && orderDirection != "DESC" {
		orderDirection = "DESC"
	}

	return column, orderDirection
}

func scanQuotation(rows *sql.Rows) (models.Quotation, error) {
	var n models.NullAbleQuotation
	err := rows.Scan(&n.Id, &n.QuoteNumber,
		&n.SalesUserId, &n.SalesUsername, &n.SalesName,
		&n.CreatedByUserId, &n.CreatedByName,
		&n.CustomerId, &n.CustomerCode, &n.CustomerName,
		&n.BrandId, &n.BrandCode, &n.BrandName,
		&n.RateCardVersionId, &n.RateCardVersionCode,
		&n.AttentionTo, &n.JobTitle, &n.ContactPhone, &n.ContactEmail,
		&n.CampaignYear, &n.ValidUntil,
		&n.Discount, &n.TaxRate, &n.Status, &n.RequiredApproverUserId, &n.RequiredApproverName, &n.Version,
		&n.PlacementGross, &n.PlacementDiscountAmount, &n.PlacementNet,
		&n.BonusGross, &n.BonusNet, &n.TotalGross, &n.TotalNet,
		&n.EffectiveDiscountAmount, &n.EffectiveDiscountRate,
		&n.Tax, &n.TotalIncludingTax,
		&n.CreatedAt, &n.UpdatedAt, &n.ApprovedAt)
	if err != nil {
		return models.Quotation{}, err
	}

	return models.NullAbleQuotationToQuotation(n), nil
}

// buildScope turns a ListScope into a WHERE clause and its arguments.
//
// The scope is what keeps a salesperson from reading someone else's pipeline, so it
// is built here rather than left to each caller to remember.
func buildScope(scope ListScope, startAt int) (string, []interface{}) {
	var clauses []string
	var args []interface{}

	// bind appends the argument AND returns its placeholder, so the two can never
	// disagree. They previously did: the placeholder was numbered after the append,
	// which made the first argument $2 and left $1 unreferenced. Postgres cannot
	// infer a type for a parameter that appears nowhere, so every scoped query --
	// meaning every list a non-admin ever loads -- failed with 42P18.
	bind := func(value interface{}) string {
		args = append(args, value)

		return "$" + strconv.Itoa(startAt+len(args)-1)
	}

	if scope.OwnedByUserId > 0 {
		clauses = append(clauses, "q.sales_user_id = "+bind(scope.OwnedByUserId))
	}
	if scope.AwaitingApprovalByUserId > 0 {
		clauses = append(clauses, "q.required_approver_user_id = "+bind(scope.AwaitingApprovalByUserId))
		clauses = append(clauses, "q.status IN ('pending_manager','pending_business_control','pending_ceo')")
	}
	if scope.Status != "" {
		clauses = append(clauses, "q.status = "+bind(scope.Status))
	}
	if scope.Search != "" {
		p := bind("%" + scope.Search + "%")
		clauses = append(clauses, "(q.quote_number ILIKE "+p+" OR c.name ILIKE "+p+" OR b.name ILIKE "+p+")")
	}

	if len(clauses) == 0 {
		return "", args
	}

	return " WHERE " + strings.Join(clauses, " AND "), args
}

func (r *RepositoryQuotationImpl) Create(ctx context.Context, tx *sql.Tx, q models.Quotation) (models.Quotation, error) {
	SQL := `INSERT INTO ` + models.QuotationTable + ` (quote_number, sales_user_id, created_by_user_id,
		customer_id, brand_id, rate_card_version_id, attention_to, job_title, contact_phone, contact_email,
		campaign_year, valid_until, discount, tax_rate, status)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
		RETURNING id, created_at, updated_at`
	err := tx.QueryRowContext(ctx, SQL,
		q.QuoteNumber, q.SalesUserId, q.CreatedByUserId, q.CustomerId, q.BrandId,
		nullIfZero(q.RateCardVersionId), q.AttentionTo, q.JobTitle, q.ContactPhone, q.ContactEmail,
		nullIfZero(q.CampaignYear), nullIfEmpty(q.ValidUntil), q.Discount, q.TaxRate, q.Status,
	).Scan(&q.Id, &q.CreatedAt, &q.UpdatedAt)

	return q, err
}

func (r *RepositoryQuotationImpl) FindAll(ctx context.Context, tx *sql.Tx, take int, skip int, orderBy string, orderDirection string, scope ListScope) ([]models.Quotation, error) {
	column, direction := safeOrder(orderBy, orderDirection)
	where, args := buildScope(scope, 1)

	SQL := quotationSelect + where + " ORDER BY " + column + " " + direction +
		" LIMIT $" + strconv.Itoa(len(args)+1) + " OFFSET $" + strconv.Itoa(len(args)+2)
	args = append(args, take, skip)

	rows, err := tx.QueryContext(ctx, SQL, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.Quotation
	for rows.Next() {
		item, err := scanQuotation(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, item)
	}

	return list, rows.Err()
}

func (r *RepositoryQuotationImpl) CountAll(ctx context.Context, tx *sql.Tx, scope ListScope) (int, error) {
	where, args := buildScope(scope, 1)
	SQL := `SELECT COUNT(*) FROM ` + models.QuotationTable + ` q
		JOIN ` + models.CustomerTable + ` c ON c.id = q.customer_id
		JOIN ` + models.BrandTable + ` b ON b.id = q.brand_id` + where

	var total int
	err := tx.QueryRowContext(ctx, SQL, args...).Scan(&total)

	return total, err
}

func (r *RepositoryQuotationImpl) FindById(ctx context.Context, tx *sql.Tx, id int) (models.Quotation, error) {
	rows, err := tx.QueryContext(ctx, quotationSelect+" WHERE q.id = $1", id)
	if err != nil {
		return models.Quotation{}, err
	}
	defer rows.Close()

	if !rows.Next() {
		return models.Quotation{}, sql.ErrNoRows
	}

	quotation, err := scanQuotation(rows)
	if err != nil {
		return models.Quotation{}, err
	}
	rows.Close()

	selections, err := r.FindSelections(ctx, tx, id)
	if err != nil {
		return models.Quotation{}, err
	}
	quotation.Selections = selections

	return quotation, nil
}

// UpdateDraft writes only the fields a draft owner may change. Pricing and status
// are deliberately absent: those are the server's to set on submit.
func (r *RepositoryQuotationImpl) UpdateDraft(ctx context.Context, tx *sql.Tx, q models.Quotation) error {
	SQL := `UPDATE ` + models.QuotationTable + ` SET customer_id = $1, brand_id = $2,
		attention_to = $3, job_title = $4, contact_phone = $5, contact_email = $6,
		campaign_year = $7, valid_until = $8, discount = $9, tax_rate = $10, updated_at = $11
		WHERE id = $12`
	_, err := tx.ExecContext(ctx, SQL, q.CustomerId, q.BrandId, q.AttentionTo, q.JobTitle,
		q.ContactPhone, q.ContactEmail, nullIfZero(q.CampaignYear), nullIfEmpty(q.ValidUntil),
		q.Discount, q.TaxRate, time.Now(), q.Id)

	return err
}

func (r *RepositoryQuotationImpl) UpdatePricingAndStatus(ctx context.Context, tx *sql.Tx, q models.Quotation) error {
	SQL := `UPDATE ` + models.QuotationTable + ` SET
		rate_card_version_id = $1, status = $2, required_approver_user_id = $3, version = $4,
		placement_gross = $5, placement_discount_amount = $6, placement_net = $7,
		bonus_gross = $8, bonus_net = $9, total_gross = $10, total_net = $11,
		effective_discount_amount = $12, effective_discount_rate = $13,
		tax = $14, total_including_tax = $15, approved_at = $16, updated_at = $17
		WHERE id = $18`
	_, err := tx.ExecContext(ctx, SQL,
		nullIfZero(q.RateCardVersionId), q.Status, nullIfZero(q.RequiredApproverUserId), q.Version,
		q.PlacementGross, q.PlacementDiscountAmount, q.PlacementNet,
		q.BonusGross, q.BonusNet, q.TotalGross, q.TotalNet,
		q.EffectiveDiscountAmount, q.EffectiveDiscountRate,
		q.Tax, q.TotalIncludingTax, nullIfEmpty(q.ApprovedAt), time.Now(), q.Id)

	return err
}

func (r *RepositoryQuotationImpl) Delete(ctx context.Context, tx *sql.Tx, id int) error {
	_, err := tx.ExecContext(ctx, "DELETE FROM "+models.QuotationTable+" WHERE id = $1", id)

	return err
}

// ReplaceSelections rewrites both sides wholesale. Diffing would leave orphaned items
// behind when a selection changes mode.
func (r *RepositoryQuotationImpl) ReplaceSelections(ctx context.Context, tx *sql.Tx, quotationId int, selections []models.QuotationSelection) error {
	if _, err := tx.ExecContext(ctx,
		"DELETE FROM "+models.QuotationSelectionTable+" WHERE quotation_id = $1", quotationId); err != nil {
		return err
	}

	for _, selection := range selections {
		var selectionId int
		SQL := `INSERT INTO ` + models.QuotationSelectionTable + ` (quotation_id, kind, mode,
			sales_package_id, sales_package_name, tvc_duration_seconds, weeks, spots,
			gross_price, traffic, impressions, screen_count)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12) RETURNING id`
		err := tx.QueryRowContext(ctx, SQL, quotationId, selection.Kind, selection.Mode,
			nullIfZero(selection.SalesPackageId), nullIfEmpty(selection.SalesPackageName),
			selection.TvcDurationSeconds, selection.Weeks, selection.Spots,
			selection.GrossPrice, selection.Traffic, selection.Impressions, selection.ScreenCount,
		).Scan(&selectionId)
		if err != nil {
			return err
		}

		for _, item := range selection.Items {
			itemSQL := `INSERT INTO ` + models.QuotationSelectionItemTable + ` (quotation_selection_id,
				building_id, building_name, building_iris_code, building_type, citytown,
				unit_price_idr, traffic, impressions) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`
			if _, err := tx.ExecContext(ctx, itemSQL, selectionId, nullIfZero(item.BuildingId),
				item.BuildingName, item.BuildingIrisCode, item.BuildingType, item.Citytown,
				item.UnitPriceIdr, item.Traffic, item.Impressions); err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *RepositoryQuotationImpl) FindSelections(ctx context.Context, tx *sql.Tx, quotationId int) ([]models.QuotationSelection, error) {
	SQL := `SELECT id, quotation_id, kind, mode, COALESCE(sales_package_id,0), COALESCE(sales_package_name,''),
		tvc_duration_seconds, weeks, spots, gross_price, traffic, impressions, screen_count
		FROM ` + models.QuotationSelectionTable + ` WHERE quotation_id = $1 ORDER BY kind DESC`

	rows, err := tx.QueryContext(ctx, SQL, quotationId)
	if err != nil {
		return nil, err
	}

	var selections []models.QuotationSelection
	for rows.Next() {
		var s models.QuotationSelection
		if err := rows.Scan(&s.Id, &s.QuotationId, &s.Kind, &s.Mode, &s.SalesPackageId,
			&s.SalesPackageName, &s.TvcDurationSeconds, &s.Weeks, &s.Spots,
			&s.GrossPrice, &s.Traffic, &s.Impressions, &s.ScreenCount); err != nil {
			rows.Close()
			return nil, err
		}
		s.Items = []models.QuotationSelectionItem{}
		selections = append(selections, s)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for i := range selections {
		items, err := r.findSelectionItems(ctx, tx, selections[i].Id)
		if err != nil {
			return nil, err
		}
		selections[i].Items = items
	}

	return selections, nil
}

func (r *RepositoryQuotationImpl) findSelectionItems(ctx context.Context, tx *sql.Tx, selectionId int) ([]models.QuotationSelectionItem, error) {
	SQL := `SELECT id, quotation_selection_id, COALESCE(building_id,0), building_name,
		COALESCE(building_iris_code,''), COALESCE(building_type,''), COALESCE(citytown,''),
		unit_price_idr, traffic, impressions
		FROM ` + models.QuotationSelectionItemTable + ` WHERE quotation_selection_id = $1
		ORDER BY building_name`

	rows, err := tx.QueryContext(ctx, SQL, selectionId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []models.QuotationSelectionItem{}
	for rows.Next() {
		var item models.QuotationSelectionItem
		if err := rows.Scan(&item.Id, &item.QuotationSelectionId, &item.BuildingId,
			&item.BuildingName, &item.BuildingIrisCode, &item.BuildingType, &item.Citytown,
			&item.UnitPriceIdr, &item.Traffic, &item.Impressions); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *RepositoryQuotationImpl) CreateVersionSnapshot(ctx context.Context, tx *sql.Tx, quotationId int, version int, snapshot []byte, approverUserId int) error {
	SQL := `INSERT INTO ` + models.QuotationVersionTable + ` (quotation_id, version, snapshot, required_approver_user_id)
		VALUES ($1,$2,$3,$4)`
	_, err := tx.ExecContext(ctx, SQL, quotationId, version, snapshot, nullIfZero(approverUserId))

	return err
}

func (r *RepositoryQuotationImpl) CreateApproval(ctx context.Context, tx *sql.Tx, a models.QuotationApproval) error {
	SQL := `INSERT INTO ` + models.QuotationApprovalTable + ` (quotation_id, version, actor_user_id,
		actor_role, action, comment) VALUES ($1,$2,$3,$4,$5,$6)`
	_, err := tx.ExecContext(ctx, SQL, a.QuotationId, a.Version, nullIfZero(a.ActorUserId),
		a.ActorRole, a.Action, nullIfEmpty(a.Comment))

	return err
}

func (r *RepositoryQuotationImpl) FindApprovals(ctx context.Context, tx *sql.Tx, quotationId int) ([]models.QuotationApproval, error) {
	SQL := `SELECT a.id, a.quotation_id, a.version, COALESCE(a.actor_user_id,0), COALESCE(u.name,''),
		COALESCE(a.actor_role,''), a.action, COALESCE(a.comment,''), a.created_at
		FROM ` + models.QuotationApprovalTable + ` a
		LEFT JOIN ` + models.UserTable + ` u ON u.id = a.actor_user_id
		WHERE a.quotation_id = $1 ORDER BY a.created_at ASC, a.id ASC`

	rows, err := tx.QueryContext(ctx, SQL, quotationId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []models.QuotationApproval{}
	for rows.Next() {
		var a models.QuotationApproval
		if err := rows.Scan(&a.Id, &a.QuotationId, &a.Version, &a.ActorUserId, &a.ActorName,
			&a.ActorRole, &a.Action, &a.Comment, &a.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, a)
	}

	return list, rows.Err()
}

// NextQuoteNumber produces Q-<year>-<sequence>, sequence counted within the year.
//
// It reads the current maximum inside the caller's transaction. Two concurrent
// submits could still collide, and the unique constraint on quote_number turns that
// into a visible error rather than a duplicate.
func (r *RepositoryQuotationImpl) NextQuoteNumber(ctx context.Context, tx *sql.Tx, year int) (string, error) {
	prefix := fmt.Sprintf("Q-%d-", year)

	var count int
	err := tx.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM "+models.QuotationTable+" WHERE quote_number LIKE $1", prefix+"%").Scan(&count)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s%04d", prefix, count+1), nil
}

func (r *RepositoryQuotationImpl) CountByStatusForOwner(ctx context.Context, tx *sql.Tx, ownerUserId int) (map[string]int, error) {
	rows, err := tx.QueryContext(ctx,
		"SELECT status, COUNT(*) FROM "+models.QuotationTable+" WHERE sales_user_id = $1 GROUP BY status",
		ownerUserId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counts := map[string]int{}
	for rows.Next() {
		var status string
		var total int
		if err := rows.Scan(&status, &total); err != nil {
			return nil, err
		}
		counts[status] = total
	}

	return counts, rows.Err()
}

func nullIfZero(value int) interface{} {
	if value == 0 {
		return nil
	}

	return value
}

func nullIfEmpty(value string) interface{} {
	if value == "" {
		return nil
	}

	return value
}
