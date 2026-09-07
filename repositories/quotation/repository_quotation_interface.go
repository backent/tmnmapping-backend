package quotation

import (
	"context"
	"database/sql"

	"github.com/malikabdulaziz/tmn-backend/models"
)

// ListScope decides which quotations a caller sees.
type ListScope struct {
	// OwnedByUserId limits results to quotations this user commercially owns.
	OwnedByUserId int
	// AwaitingApprovalByUserId limits results to this user's approval queue.
	AwaitingApprovalByUserId int
	// Status filters by a single status when set.
	Status string
	Search string
}

type RepositoryQuotationInterface interface {
	Create(ctx context.Context, tx *sql.Tx, quotation models.Quotation) (models.Quotation, error)
	FindAll(ctx context.Context, tx *sql.Tx, take int, skip int, orderBy string, orderDirection string, scope ListScope) ([]models.Quotation, error)
	CountAll(ctx context.Context, tx *sql.Tx, scope ListScope) (int, error)
	FindById(ctx context.Context, tx *sql.Tx, id int) (models.Quotation, error)
	UpdateDraft(ctx context.Context, tx *sql.Tx, quotation models.Quotation) error
	UpdatePricingAndStatus(ctx context.Context, tx *sql.Tx, quotation models.Quotation) error
	Delete(ctx context.Context, tx *sql.Tx, id int) error

	ReplaceSelections(ctx context.Context, tx *sql.Tx, quotationId int, selections []models.QuotationSelection) error
	FindSelections(ctx context.Context, tx *sql.Tx, quotationId int) ([]models.QuotationSelection, error)

	CreateVersionSnapshot(ctx context.Context, tx *sql.Tx, quotationId int, version int, snapshot []byte, approverUserId int) error
	CreateApproval(ctx context.Context, tx *sql.Tx, approval models.QuotationApproval) error
	FindApprovals(ctx context.Context, tx *sql.Tx, quotationId int) ([]models.QuotationApproval, error)

	// NextQuoteNumber reserves the next sequential number for a year.
	NextQuoteNumber(ctx context.Context, tx *sql.Tx, year int) (string, error)

	// CountByStatusForOwner backs the sales dashboard counts.
	CountByStatusForOwner(ctx context.Context, tx *sql.Tx, ownerUserId int) (map[string]int, error)
}
