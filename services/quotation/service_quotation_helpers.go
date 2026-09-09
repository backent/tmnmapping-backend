package quotation

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/malikabdulaziz/tmn-backend/exceptions"
	"github.com/malikabdulaziz/tmn-backend/helpers"
	"github.com/malikabdulaziz/tmn-backend/models"
	webQuotation "github.com/malikabdulaziz/tmn-backend/web/quotation"
)

// ---------------------------------------------------------------------------
// Guards
// ---------------------------------------------------------------------------

func (s *ServiceQuotationImpl) mustFind(ctx context.Context, tx *sql.Tx, id int) models.Quotation {
	quotation, err := s.RepositoryQuotationInterface.FindById(ctx, tx, id)
	if err == sql.ErrNoRows {
		panic(exceptions.NewNotFoundError("quotation not found"))
	}
	helpers.PanicIfError(err)

	return quotation
}

// assertCanRead: the owner, the assigned approver, anyone who has already acted on
// it, and admin. Nobody else reads someone else's pipeline.
func (s *ServiceQuotationImpl) assertCanRead(
	quotation models.Quotation, approvals []models.QuotationApproval, actor Actor,
) {
	if actor.Role == models.RoleAdmin ||
		quotation.SalesUserId == actor.UserId ||
		quotation.CreatedByUserId == actor.UserId ||
		(quotation.RequiredApproverUserId != 0 && quotation.RequiredApproverUserId == actor.UserId) {
		return
	}

	// Approving and returning both clear required_approver_user_id, so the check
	// above stops matching the instant the decision is made -- which locked an
	// approver out of the quotation they had just signed off. Anyone the audit trail
	// names keeps read access: they need to see what they decided.
	for _, approval := range approvals {
		if approval.ActorUserId != 0 && approval.ActorUserId == actor.UserId {
			return
		}
	}

	panic(exceptions.NewForbidden("this quotation belongs to someone else"))
}

// mustFindEditable enforces the two rules that govern every write: only the owner
// (or the proxy who entered it, or an admin) may change a quotation, and only while
// it is a draft or has been returned.
func (s *ServiceQuotationImpl) mustFindEditable(ctx context.Context, tx *sql.Tx, id int, actor Actor) models.Quotation {
	quotation := s.mustFind(ctx, tx, id)

	if actor.Role != models.RoleAdmin &&
		quotation.SalesUserId != actor.UserId &&
		quotation.CreatedByUserId != actor.UserId {
		panic(exceptions.NewForbidden("this quotation belongs to someone else"))
	}

	if quotation.Status != StatusDraft && quotation.Status != StatusReturned {
		panic(exceptions.NewBadRequestError(
			"this quotation is " + quotation.Status + " and can only be changed while it is a draft or returned"))
	}

	return quotation
}

// mustFindPendingForApprover: only the person the routing named may act, and only
// while the quotation is actually waiting.
func (s *ServiceQuotationImpl) mustFindPendingForApprover(ctx context.Context, tx *sql.Tx, id int, actor Actor) models.Quotation {
	quotation := s.mustFind(ctx, tx, id)

	if !isPending(quotation.Status) {
		panic(exceptions.NewBadRequestError("this quotation is " + quotation.Status + " and is not awaiting approval"))
	}

	if quotation.RequiredApproverUserId != actor.UserId {
		panic(exceptions.NewForbidden("this quotation is waiting on a different approver"))
	}

	return quotation
}

// assertCanCreateFor covers proxy entry: entering a quotation owned by someone else
// requires an explicit allow-list entry, per spec 2026-07-23.
func (s *ServiceQuotationImpl) assertCanCreateFor(ctx context.Context, tx *sql.Tx, actor Actor, ownerId int) {
	actorUser, err := s.RepositoryUser.FindById(ctx, tx, actor.UserId)
	if err == sql.ErrNoRows {
		panic(exceptions.NewUnAuthorized("authorization invalid"))
	}
	helpers.PanicIfError(err)

	if ownerId == actor.UserId {
		if !actorUser.CanRaiseQuotations() {
			panic(exceptions.NewForbidden("your account is not permitted to create quotations"))
		}

		return
	}

	// Entering on behalf of someone else.
	if actor.Role == models.RoleAdmin {
		return
	}

	allowed, err := s.RepositoryUser.CanCreateOnBehalfOf(ctx, tx, actor.UserId, ownerId)
	helpers.PanicIfError(err)
	if !allowed {
		panic(exceptions.NewForbidden("you are not permitted to enter quotations for that sales owner"))
	}
}

// assertCustomerAndBrand enforces the visibility rule the whole customer model
// exists for: a seller may only quote a customer+brand they are the PIC for.
func (s *ServiceQuotationImpl) assertCustomerAndBrand(ctx context.Context, tx *sql.Tx, customerId, brandId, ownerId int, actor Actor) {
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

	assignment, err := s.RepositorySalesAssignment.FindByCustomerAndBrand(ctx, tx, customerId, brandId)
	if err == sql.ErrNoRows {
		panic(exceptions.NewBadRequestError("that brand has no sales PIC assigned, so it cannot be quoted yet"))
	}
	helpers.PanicIfError(err)

	// Admin can quote anything; everyone else must own the assignment.
	if actor.Role != models.RoleAdmin && assignment.SalesUserId != ownerId {
		panic(exceptions.NewForbidden("that customer and brand are assigned to a different sales owner"))
	}
}

// ---------------------------------------------------------------------------
// Selections
// ---------------------------------------------------------------------------

func (s *ServiceQuotationImpl) applySelections(ctx context.Context, tx *sql.Tx, quotationId int, placement, bonus *webQuotation.SelectionRequest) {
	version, err := s.RepositoryRateCard.FindCurrentVersion(ctx, tx)
	hasRateCard := err == nil
	if err != nil && err != sql.ErrNoRows {
		helpers.PanicIfError(err)
	}

	versionId := 0
	if hasRateCard {
		versionId = version.Id
	}

	var selections []models.QuotationSelection
	for _, pair := range []struct {
		kind string
		req  *webQuotation.SelectionRequest
	}{
		{models.SelectionKindPlacement, placement},
		{models.SelectionKindBonus, bonus},
	} {
		if pair.req == nil {
			continue
		}
		selections = append(selections, s.buildSelection(ctx, tx, versionId, pair.kind, *pair.req))
	}

	helpers.PanicIfError(s.RepositoryQuotationInterface.ReplaceSelections(ctx, tx, quotationId, selections))
}

// buildSelection resolves a client's choice into a priced selection.
//
// The client sends only WHAT it picked. Every price, traffic and impression figure
// is read here from the rate card and the master data, so the client is never
// trusted for money.
func (s *ServiceQuotationImpl) buildSelection(ctx context.Context, tx *sql.Tx, versionId int, kind string, request webQuotation.SelectionRequest) models.QuotationSelection {
	selection := models.QuotationSelection{
		Kind:               kind,
		Mode:               request.Mode,
		TvcDurationSeconds: request.TvcDurationSeconds,
		Weeks:              request.Weeks,
		Spots:              request.Spots,
		Items:              []models.QuotationSelectionItem{},
	}

	switch request.Mode {
	case models.SelectionModePackage:
		if request.SalesPackageId == 0 {
			panic(exceptions.NewBadRequestError("a package selection needs exactly one package"))
		}

		pkg, err := s.RepositorySalesPackage.FindById(ctx, tx, request.SalesPackageId)
		if err == sql.ErrNoRows {
			panic(exceptions.NewBadRequestError("sales package not found"))
		}
		helpers.PanicIfError(err)

		// A package carries its own screen count, traffic and impressions rather
		// than summing its buildings. See docs/QUOTATION_DOCUMENT_ANALYSIS.md §4.1.
		selection.SalesPackageId = pkg.Id
		selection.SalesPackageName = pkg.Name
		selection.ScreenCount = pkg.ScreenCount
		selection.Traffic = int64(pkg.Traffic)
		selection.Impressions = int64(pkg.Impressions)

		selection.GrossPrice = s.packageWeeklyRate(ctx, tx, versionId, pkg.Id) * int64(request.Weeks)

	case models.SelectionModeBuilding:
		if len(request.BuildingIds) == 0 {
			panic(exceptions.NewBadRequestError("a building selection needs at least one building"))
		}

		var weeklyRate int64
		for _, buildingId := range request.BuildingIds {
			building, err := s.RepositoryBuilding.FindById(ctx, tx, buildingId)
			if err == sql.ErrNoRows {
				panic(exceptions.NewBadRequestError("building not found"))
			}
			helpers.PanicIfError(err)

			price := s.buildingWeeklyRate(ctx, tx, versionId, buildingId)
			if price == 0 {
				panic(exceptions.NewBadRequestError(
					"\"" + building.Name + "\" has no price in the published rate card, so it cannot be quoted"))
			}

			weeklyRate += price
			selection.Traffic += int64(building.Audience)
			selection.Impressions += int64(building.Impression)
			selection.ScreenCount++
			selection.Items = append(selection.Items, models.QuotationSelectionItem{
				BuildingId:       building.Id,
				BuildingName:     building.Name,
				BuildingIrisCode: building.IrisCode,
				BuildingType:     building.BuildingType,
				Citytown:         building.Citytown,
				UnitPriceIdr:     price,
				Traffic:          int64(building.Audience),
				Impressions:      int64(building.Impression),
			})
		}

		selection.GrossPrice = weeklyRate * int64(request.Weeks)

	default:
		panic(exceptions.NewBadRequestError("unknown selection mode"))
	}

	return selection
}

// repriceSelection re-reads prices at submit time from the published rate card,
// discarding whatever was stored while the quotation was a draft.
func (s *ServiceQuotationImpl) repriceSelection(ctx context.Context, tx *sql.Tx, versionId int, selection models.QuotationSelection) models.QuotationSelection {
	request := webQuotation.SelectionRequest{
		Mode:               selection.Mode,
		SalesPackageId:     selection.SalesPackageId,
		TvcDurationSeconds: selection.TvcDurationSeconds,
		Weeks:              selection.Weeks,
		Spots:              selection.Spots,
	}
	for _, item := range selection.Items {
		request.BuildingIds = append(request.BuildingIds, item.BuildingId)
	}

	return s.buildSelection(ctx, tx, versionId, selection.Kind, request)
}

func (s *ServiceQuotationImpl) buildingWeeklyRate(ctx context.Context, tx *sql.Tx, versionId, buildingId int) int64 {
	if versionId == 0 {
		return 0
	}

	prices, err := s.RepositoryRateCard.FindBuildingPrices(ctx, tx, versionId, 100000, 0, "")
	helpers.PanicIfError(err)

	for _, price := range prices {
		if price.BuildingId == buildingId {
			return price.PriceIdrPerWeek
		}
	}

	return 0
}

func (s *ServiceQuotationImpl) packageWeeklyRate(ctx context.Context, tx *sql.Tx, versionId, packageId int) int64 {
	if versionId == 0 {
		panic(exceptions.NewBadRequestError("no rate card has been published yet, so nothing can be priced"))
	}

	prices, err := s.RepositoryRateCard.FindPackagePrices(ctx, tx, versionId, 100000, 0, "")
	helpers.PanicIfError(err)

	for _, price := range prices {
		if price.SalesPackageId == packageId {
			return price.PriceIdrPerWeek
		}
	}

	panic(exceptions.NewBadRequestError("that package has no price in the published rate card"))
}

// ---------------------------------------------------------------------------
// Mapping
// ---------------------------------------------------------------------------

func applyPricing(q *models.Quotation, summary PricingSummary) {
	q.PlacementGross = summary.PlacementGross
	q.PlacementDiscountAmount = summary.PlacementDiscountAmount
	q.PlacementNet = summary.PlacementNet
	q.BonusGross = summary.BonusGross
	q.BonusNet = summary.BonusNet
	q.TotalGross = summary.TotalGross
	q.TotalNet = summary.TotalNet
	q.EffectiveDiscountAmount = summary.EffectiveDiscountAmount
	q.EffectiveDiscountRate = summary.EffectiveDiscountRate
	q.Tax = summary.Tax
	q.TotalIncludingTax = summary.TotalIncludingTax
}

func toPricingResponse(s PricingSummary) webQuotation.PricingResponse {
	return webQuotation.PricingResponse{
		PlacementGross: s.PlacementGross, PlacementDiscountAmount: s.PlacementDiscountAmount,
		PlacementNet: s.PlacementNet, BonusGross: s.BonusGross, BonusNet: s.BonusNet,
		TotalGross: s.TotalGross, TotalNet: s.TotalNet,
		EffectiveDiscountAmount: s.EffectiveDiscountAmount,
		EffectiveDiscountRate:   s.EffectiveDiscountRate,
		Tax:                     s.Tax, TotalIncludingTax: s.TotalIncludingTax,
	}
}

func toSelectionResponse(s models.QuotationSelection) webQuotation.SelectionResponse {
	items := make([]webQuotation.SelectionItemResponse, len(s.Items))
	for i, item := range s.Items {
		items[i] = webQuotation.SelectionItemResponse{
			BuildingId: item.BuildingId, BuildingName: item.BuildingName,
			BuildingIrisCode: item.BuildingIrisCode, BuildingType: item.BuildingType,
			Citytown: item.Citytown, UnitPriceIdr: item.UnitPriceIdr,
			Traffic: item.Traffic, Impressions: item.Impressions,
		}
	}

	perWeek := int64(0)
	if s.Weeks > 0 {
		perWeek = s.GrossPrice / int64(s.Weeks)
	}

	return webQuotation.SelectionResponse{
		Kind: s.Kind, Mode: s.Mode, SalesPackageId: s.SalesPackageId,
		SalesPackageName: s.SalesPackageName, TvcDurationSeconds: s.TvcDurationSeconds,
		Weeks: s.Weeks, Spots: s.Spots, GrossPricePerWeek: perWeek, GrossPrice: s.GrossPrice,
		Traffic: s.Traffic, Impressions: s.Impressions, ScreenCount: s.ScreenCount,
		Items: items,
	}
}

func toResponse(q models.Quotation, approvals []models.QuotationApproval) webQuotation.QuotationResponse {
	selections := make([]webQuotation.SelectionResponse, len(q.Selections))
	for i, selection := range q.Selections {
		selections[i] = toSelectionResponse(selection)
	}

	events := make([]webQuotation.ApprovalResponse, len(approvals))
	for i, a := range approvals {
		events[i] = webQuotation.ApprovalResponse{
			Version: a.Version, ActorName: a.ActorName, ActorRole: a.ActorRole,
			Action: a.Action, Comment: a.Comment, CreatedAt: a.CreatedAt,
		}
	}

	return webQuotation.QuotationResponse{
		Id: q.Id, QuoteNumber: q.QuoteNumber,
		SalesUserId: q.SalesUserId, SalesName: q.SalesName, CreatedByName: q.CreatedByName,
		IsProxyEntry: q.CreatedByUserId != q.SalesUserId,
		CustomerId:   q.CustomerId, CustomerName: q.CustomerName,
		BrandId: q.BrandId, BrandName: q.BrandName,
		RateCardVersionCode: q.RateCardVersionCode,
		AttentionTo:         q.AttentionTo, JobTitle: q.JobTitle,
		ContactPhone: q.ContactPhone, ContactEmail: q.ContactEmail,
		CampaignYear: q.CampaignYear, ValidUntil: q.ValidUntil,
		Discount: q.Discount, TaxRate: q.TaxRate,
		Status:               q.Status,
		IsEditable:           q.Status == StatusDraft || q.Status == StatusReturned,
		RequiredApproverName: q.RequiredApproverName, Version: q.Version,
		Pricing: webQuotation.PricingResponse{
			PlacementGross: q.PlacementGross, PlacementDiscountAmount: q.PlacementDiscountAmount,
			PlacementNet: q.PlacementNet, BonusGross: q.BonusGross, BonusNet: q.BonusNet,
			TotalGross: q.TotalGross, TotalNet: q.TotalNet,
			EffectiveDiscountAmount: q.EffectiveDiscountAmount,
			EffectiveDiscountRate:   q.EffectiveDiscountRate,
			Tax:                     q.Tax, TotalIncludingTax: q.TotalIncludingTax,
		},
		Selections: selections, Approvals: events,
		CreatedAt: q.CreatedAt, UpdatedAt: q.UpdatedAt, ApprovedAt: q.ApprovedAt,
	}
}

// ---------------------------------------------------------------------------

func isPending(status string) bool {
	return status == StatusPendingManager ||
		status == StatusPendingBusinessControl ||
		status == StatusPendingCEO
}

func hasKind(selections []models.QuotationSelection, kind string) bool {
	for _, s := range selections {
		if s.Kind == kind {
			return true
		}
	}

	return false
}

func isBlank(value string) bool { return strings.TrimSpace(value) == "" }

func errorsIs(err, target error) bool { return errors.Is(err, target) }

func max(a, b int) int {
	if a > b {
		return a
	}

	return b
}
