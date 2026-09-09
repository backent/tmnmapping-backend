package quotation

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/malikabdulaziz/tmn-backend/exceptions"
	"github.com/malikabdulaziz/tmn-backend/helpers"
	"github.com/malikabdulaziz/tmn-backend/models"
	repositoriesBrand "github.com/malikabdulaziz/tmn-backend/repositories/brand"
	repositoriesBuilding "github.com/malikabdulaziz/tmn-backend/repositories/building"
	repositoriesCustomer "github.com/malikabdulaziz/tmn-backend/repositories/customer"
	repositoriesQuotation "github.com/malikabdulaziz/tmn-backend/repositories/quotation"
	repositoriesRateCard "github.com/malikabdulaziz/tmn-backend/repositories/ratecard"
	repositoriesSalesAssignment "github.com/malikabdulaziz/tmn-backend/repositories/salesassignment"
	repositoriesSalesPackage "github.com/malikabdulaziz/tmn-backend/repositories/salespackage"
	repositoriesUser "github.com/malikabdulaziz/tmn-backend/repositories/user"
	webQuotation "github.com/malikabdulaziz/tmn-backend/web/quotation"
)

// DefaultValidityDays matches the real template: prepared 22 June, valid until
// 22 July.
const DefaultValidityDays = 30

type ServiceQuotationImpl struct {
	DB *sql.DB
	repositoriesQuotation.RepositoryQuotationInterface
	RepositoryRateCard        repositoriesRateCard.RepositoryRateCardInterface
	RepositoryBuilding        repositoriesBuilding.RepositoryBuildingInterface
	RepositoryCustomer        repositoriesCustomer.RepositoryCustomerInterface
	RepositoryBrand           repositoriesBrand.RepositoryBrandInterface
	RepositoryUser            repositoriesUser.RepositoryUserInterface
	RepositorySalesAssignment repositoriesSalesAssignment.RepositorySalesAssignmentInterface
	RepositorySalesPackage    repositoriesSalesPackage.RepositorySalesPackageInterface
	Thresholds                Thresholds
}

func NewServiceQuotationImpl(
	db *sql.DB,
	repoQuotation repositoriesQuotation.RepositoryQuotationInterface,
	repoRateCard repositoriesRateCard.RepositoryRateCardInterface,
	repoBuilding repositoriesBuilding.RepositoryBuildingInterface,
	repoCustomer repositoriesCustomer.RepositoryCustomerInterface,
	repoBrand repositoriesBrand.RepositoryBrandInterface,
	repoUser repositoriesUser.RepositoryUserInterface,
	repoSalesAssignment repositoriesSalesAssignment.RepositorySalesAssignmentInterface,
	repoSalesPackage repositoriesSalesPackage.RepositorySalesPackageInterface,
) ServiceQuotationInterface {
	return &ServiceQuotationImpl{
		DB:                           db,
		RepositoryQuotationInterface: repoQuotation,
		RepositoryRateCard:           repoRateCard,
		RepositoryBuilding:           repoBuilding,
		RepositoryCustomer:           repoCustomer,
		RepositoryBrand:              repoBrand,
		RepositoryUser:               repoUser,
		RepositorySalesAssignment:    repoSalesAssignment,
		Thresholds:                   DefaultThresholds(),
	}
}

// txDirectory adapts the user repository to the ApproverDirectory the pure routing
// function needs, binding it to one transaction.
type txDirectory struct {
	ctx  context.Context
	tx   *sql.Tx
	repo repositoriesUser.RepositoryUserInterface
}

func (d txDirectory) FindUsersByRole(role string) ([]models.User, error) {
	return d.repo.FindByRole(d.ctx, d.tx, role)
}

// ---------------------------------------------------------------------------
// Reads
// ---------------------------------------------------------------------------

func (s *ServiceQuotationImpl) FindAll(ctx context.Context, request webQuotation.QuotationRequestFindAll, actor Actor) ([]webQuotation.QuotationResponse, int) {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	scope := s.scopeFor(request, actor)

	list, err := s.RepositoryQuotationInterface.FindAll(ctx, tx, request.GetTake(), request.GetSkip(),
		request.GetOrderBy(), request.GetOrderDirection(), scope)
	helpers.PanicIfError(err)

	total, err := s.RepositoryQuotationInterface.CountAll(ctx, tx, scope)
	helpers.PanicIfError(err)

	responses := make([]webQuotation.QuotationResponse, len(list))
	for i, item := range list {
		responses[i] = toResponse(item, nil)
	}

	return responses, total
}

// scopeFor is the visibility rule.
//
// Admin sees everything. Everyone else sees only what they own, unless they are
// explicitly asking for their approval queue -- an approver must be able to read a
// quotation they did not write in order to decide on it.
func (s *ServiceQuotationImpl) scopeFor(request webQuotation.QuotationRequestFindAll, actor Actor) repositoriesQuotation.ListScope {
	scope := repositoriesQuotation.ListScope{Status: request.Status, Search: request.GetSearch()}

	switch {
	case request.AwaitingMe:
		scope.AwaitingApprovalByUserId = actor.UserId
	case request.Mine:
		scope.OwnedByUserId = actor.UserId
	case actor.Role == models.RoleAdmin:
		// unrestricted
	default:
		scope.OwnedByUserId = actor.UserId
	}

	return scope
}

func (s *ServiceQuotationImpl) FindById(ctx context.Context, id int, actor Actor) webQuotation.QuotationResponse {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	quotation := s.mustFind(ctx, tx, id)

	// Loaded before the check, not after: the audit trail is part of who may read.
	approvals, err := s.RepositoryQuotationInterface.FindApprovals(ctx, tx, id)
	helpers.PanicIfError(err)

	s.assertCanRead(quotation, approvals, actor)

	return toResponse(quotation, approvals)
}

func (s *ServiceQuotationImpl) DashboardCounts(ctx context.Context, actor Actor) webQuotation.DashboardCounts {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	byStatus, err := s.RepositoryQuotationInterface.CountByStatusForOwner(ctx, tx, actor.UserId)
	helpers.PanicIfError(err)

	counts := webQuotation.DashboardCounts{
		Draft:    byStatus[StatusDraft],
		Returned: byStatus[StatusReturned],
		Approved: byStatus[StatusApproved],
	}
	counts.Pending = byStatus[StatusPendingManager] +
		byStatus[StatusPendingBusinessControl] + byStatus[StatusPendingCEO]
	for _, n := range byStatus {
		counts.All += n
	}

	return counts
}

// ---------------------------------------------------------------------------
// Writes
// ---------------------------------------------------------------------------

func (s *ServiceQuotationImpl) Create(ctx context.Context, request webQuotation.CreateQuotationRequest, actor Actor) webQuotation.QuotationResponse {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	ownerId := request.SalesOwnerUserId
	if ownerId == 0 {
		ownerId = actor.UserId
	}
	s.assertCanCreateFor(ctx, tx, actor, ownerId)
	s.assertCustomerAndBrand(ctx, tx, request.CustomerId, request.BrandId, ownerId, actor)

	year := request.CampaignYear
	if year == 0 {
		year = time.Now().Year()
	}

	number, err := s.RepositoryQuotationInterface.NextQuoteNumber(ctx, tx, year)
	helpers.PanicIfError(err)

	validUntil := request.ValidUntil
	if validUntil == "" {
		validUntil = time.Now().AddDate(0, 0, DefaultValidityDays).Format("2006-01-02")
	}

	taxRate := request.TaxRate
	if taxRate == 0 {
		taxRate = DefaultTaxRate
	}

	created, err := s.RepositoryQuotationInterface.Create(ctx, tx, models.Quotation{
		QuoteNumber:     number,
		SalesUserId:     ownerId,
		CreatedByUserId: actor.UserId,
		CustomerId:      request.CustomerId,
		BrandId:         request.BrandId,
		AttentionTo:     request.AttentionTo,
		JobTitle:        request.JobTitle,
		ContactPhone:    request.ContactPhone,
		ContactEmail:    request.ContactEmail,
		CampaignYear:    year,
		ValidUntil:      validUntil,
		Discount:        request.Discount,
		TaxRate:         taxRate,
		Status:          StatusDraft,
	})
	helpers.PanicIfError(err)

	s.applySelections(ctx, tx, created.Id, request.Placement, request.Bonus)

	return toResponse(s.mustFind(ctx, tx, created.Id), nil)
}

func (s *ServiceQuotationImpl) Update(ctx context.Context, request webQuotation.UpdateQuotationRequest, id int, actor Actor) webQuotation.QuotationResponse {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	existing := s.mustFindEditable(ctx, tx, id, actor)
	s.assertCustomerAndBrand(ctx, tx, request.CustomerId, request.BrandId, existing.SalesUserId, actor)

	existing.CustomerId = request.CustomerId
	existing.BrandId = request.BrandId
	existing.AttentionTo = request.AttentionTo
	existing.JobTitle = request.JobTitle
	existing.ContactPhone = request.ContactPhone
	existing.ContactEmail = request.ContactEmail
	existing.Discount = request.Discount
	if request.CampaignYear > 0 {
		existing.CampaignYear = request.CampaignYear
	}
	if request.ValidUntil != "" {
		existing.ValidUntil = request.ValidUntil
	}
	if request.TaxRate > 0 {
		existing.TaxRate = request.TaxRate
	}

	helpers.PanicIfError(s.RepositoryQuotationInterface.UpdateDraft(ctx, tx, existing))
	s.applySelections(ctx, tx, id, request.Placement, request.Bonus)

	return toResponse(s.mustFind(ctx, tx, id), nil)
}

func (s *ServiceQuotationImpl) Delete(ctx context.Context, id int, actor Actor) {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	s.mustFindEditable(ctx, tx, id, actor)
	helpers.PanicIfError(s.RepositoryQuotationInterface.Delete(ctx, tx, id))
}

// Submit prices the quotation server-side, resolves the approver, and records an
// immutable version snapshot.
//
// The client never supplies money. Everything is recomputed here from the current
// rate card, so a stale or tampered client cannot change what is charged.
func (s *ServiceQuotationImpl) Submit(ctx context.Context, id int, actor Actor) webQuotation.QuotationResponse {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	quotation := s.mustFindEditable(ctx, tx, id, actor)

	if len(quotation.Selections) == 0 {
		panic(exceptions.NewBadRequestError("this quotation has nothing selected"))
	}
	if !hasKind(quotation.Selections, models.SelectionKindPlacement) {
		panic(exceptions.NewBadRequestError("a quotation needs a placement; bonus alone cannot be sold"))
	}

	version, err := s.RepositoryRateCard.FindCurrentVersion(ctx, tx)
	if err == sql.ErrNoRows {
		panic(exceptions.NewBadRequestError("no rate card has been published yet, so nothing can be priced"))
	}
	helpers.PanicIfError(err)

	// Re-price every selection from the published rate card, then recompute totals.
	priced := make([]models.QuotationSelection, len(quotation.Selections))
	input := PricingInput{Discount: quotation.Discount, TaxRate: quotation.TaxRate}

	for i, selection := range quotation.Selections {
		repriced := s.repriceSelection(ctx, tx, version.Id, selection)
		priced[i] = repriced

		pricing := SelectionPricing{
			GrossPricePerWeek: repriced.GrossPrice / int64(max(repriced.Weeks, 1)),
			Weeks:             repriced.Weeks,
		}
		if repriced.Kind == models.SelectionKindPlacement {
			input.Placement = pricing
		} else {
			input.Bonus = pricing
		}
	}

	summary, err := CalculatePricing(input)
	if err != nil {
		panic(exceptions.NewBadRequestError(err.Error()))
	}

	route, err := ResolveApprovalRoute(quotation.Discount, quotation.SalesUserId, s.Thresholds,
		txDirectory{ctx: ctx, tx: tx, repo: s.RepositoryUser})
	if err != nil {
		panic(exceptions.NewBadRequestError(approvalErrorMessage(err)))
	}

	helpers.PanicIfError(s.RepositoryQuotationInterface.ReplaceSelections(ctx, tx, id, priced))

	isResubmission := quotation.Status == StatusReturned
	if isResubmission {
		quotation.Version++
	}

	quotation.RateCardVersionId = version.Id
	quotation.Status = route.Status
	quotation.RequiredApproverUserId = route.ApproverId
	applyPricing(&quotation, summary)
	quotation.ApprovedAt = ""

	helpers.PanicIfError(s.RepositoryQuotationInterface.UpdatePricingAndStatus(ctx, tx, quotation))

	snapshot, err := json.Marshal(quotation)
	helpers.PanicIfError(err)
	helpers.PanicIfError(s.RepositoryQuotationInterface.CreateVersionSnapshot(ctx, tx, id, quotation.Version, snapshot, route.ApproverId))

	action := models.ApprovalActionSubmitted
	if isResubmission {
		action = models.ApprovalActionResubmitted
	}
	helpers.PanicIfError(s.RepositoryQuotationInterface.CreateApproval(ctx, tx, models.QuotationApproval{
		QuotationId: id, Version: quotation.Version, ActorUserId: actor.UserId,
		ActorRole: actor.Role, Action: action,
	}))

	return toResponse(s.mustFind(ctx, tx, id), nil)
}

func (s *ServiceQuotationImpl) Approve(ctx context.Context, id int, actor Actor) webQuotation.QuotationResponse {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	quotation := s.mustFindPendingForApprover(ctx, tx, id, actor)

	quotation.Status = StatusApproved
	quotation.ApprovedAt = time.Now().Format(time.RFC3339)
	quotation.RequiredApproverUserId = 0
	helpers.PanicIfError(s.RepositoryQuotationInterface.UpdatePricingAndStatus(ctx, tx, quotation))

	helpers.PanicIfError(s.RepositoryQuotationInterface.CreateApproval(ctx, tx, models.QuotationApproval{
		QuotationId: id, Version: quotation.Version, ActorUserId: actor.UserId,
		ActorRole: actor.Role, Action: models.ApprovalActionApproved,
	}))

	return toResponse(s.mustFind(ctx, tx, id), nil)
}

// Return sends a quotation back to its owner. The comment is mandatory -- the
// database enforces it too, but refusing here gives a readable message.
func (s *ServiceQuotationImpl) Return(ctx context.Context, id int, comment string, actor Actor) webQuotation.QuotationResponse {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	if isBlank(comment) {
		panic(exceptions.NewBadRequestError("a reason is required when returning a quotation"))
	}

	quotation := s.mustFindPendingForApprover(ctx, tx, id, actor)

	quotation.Status = StatusReturned
	quotation.RequiredApproverUserId = 0
	quotation.ApprovedAt = ""
	helpers.PanicIfError(s.RepositoryQuotationInterface.UpdatePricingAndStatus(ctx, tx, quotation))

	helpers.PanicIfError(s.RepositoryQuotationInterface.CreateApproval(ctx, tx, models.QuotationApproval{
		QuotationId: id, Version: quotation.Version, ActorUserId: actor.UserId,
		ActorRole: actor.Role, Action: models.ApprovalActionReturned, Comment: comment,
	}))

	return toResponse(s.mustFind(ctx, tx, id), nil)
}

// PreviewPricing runs the same functions Submit will, so the figure the wizard shows
// is the figure the seller gets.
func (s *ServiceQuotationImpl) PreviewPricing(ctx context.Context, request webQuotation.PricingPreviewRequest, actor Actor) webQuotation.PricingPreviewResponse {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	version, err := s.RepositoryRateCard.FindCurrentVersion(ctx, tx)
	if err == sql.ErrNoRows {
		panic(exceptions.NewBadRequestError("no rate card has been published yet, so nothing can be priced"))
	}
	helpers.PanicIfError(err)

	taxRate := request.TaxRate
	if taxRate == 0 {
		taxRate = DefaultTaxRate
	}

	input := PricingInput{Discount: request.Discount, TaxRate: taxRate}
	var sections []webQuotation.SelectionResponse

	for _, pair := range []struct {
		kind string
		req  *webQuotation.SelectionRequest
	}{
		{models.SelectionKindPlacement, request.Placement},
		{models.SelectionKindBonus, request.Bonus},
	} {
		if pair.req == nil {
			continue
		}

		selection := s.buildSelection(ctx, tx, version.Id, pair.kind, *pair.req)
		pricing := SelectionPricing{
			GrossPricePerWeek: selection.GrossPrice / int64(max(selection.Weeks, 1)),
			Weeks:             selection.Weeks,
		}
		if pair.kind == models.SelectionKindPlacement {
			input.Placement = pricing
		} else {
			input.Bonus = pricing
		}
		sections = append(sections, toSelectionResponse(selection))
	}

	summary, err := CalculatePricing(input)
	if err != nil {
		panic(exceptions.NewBadRequestError(err.Error()))
	}

	return webQuotation.PricingPreviewResponse{
		Pricing:  toPricingResponse(summary),
		Approval: s.approvalHint(ctx, tx, request.Discount, actor.UserId),
		Sections: sections,
	}
}

// approvalHint answers "who will approve this" before submit, without failing the
// request when the directory is incomplete -- the wizard should still price.
func (s *ServiceQuotationImpl) approvalHint(ctx context.Context, tx *sql.Tx, discount float64, ownerId int) webQuotation.ApprovalHint {
	band := s.Thresholds.GetDiscountBand(discount)
	hint := webQuotation.ApprovalHint{Band: band, ApproverRole: RoleForBand(band)}

	route, err := ResolveApprovalRoute(discount, ownerId, s.Thresholds,
		txDirectory{ctx: ctx, tx: tx, repo: s.RepositoryUser})
	if err != nil {
		hint.Reason = approvalErrorMessage(err)

		return hint
	}

	approver, err := s.RepositoryUser.FindById(ctx, tx, route.ApproverId)
	if err == nil {
		hint.ApproverName = approver.Name
	}
	hint.ApproverRole = route.ApproverRole
	hint.Resolvable = true

	return hint
}

func approvalErrorMessage(err error) string {
	switch {
	case errorsIs(err, ErrNoApproverForRole):
		return "no user holds the approver role this discount requires; assign one under Administration > Users"
	case errorsIs(err, ErrAmbiguousApproverForRole):
		return "more than one user holds the approver role this discount requires; exactly one is needed"
	default:
		return err.Error()
	}
}
