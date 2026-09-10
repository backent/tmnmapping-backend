package quotation_test

import (
	"context"
	"testing"

	"github.com/malikabdulaziz/tmn-backend/exceptions"
	"github.com/malikabdulaziz/tmn-backend/models"
	service "github.com/malikabdulaziz/tmn-backend/services/quotation"
	"github.com/malikabdulaziz/tmn-backend/testutil"
	"github.com/malikabdulaziz/tmn-backend/testutil/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type svcDeps struct {
	buildingPrice *mocks.MockRepositoryBuildingPrice
	quotation     *mocks.MockRepositoryQuotation
	rateCard      *mocks.MockRepositoryRateCard
	building      *mocks.MockRepositoryBuilding
	customer      *mocks.MockRepositoryCustomer
	brand         *mocks.MockRepositoryBrand
	user          *mocks.MockRepositoryUser
	assignment    *mocks.MockRepositorySalesAssignment
	pkg           *mocks.MockRepositorySalesPackage
}

func newQuotationService(t *testing.T, commits bool) (service.ServiceQuotationInterface, svcDeps, func()) {
	t.Helper()

	db, sqlMock := testutil.NewMockDB(t)
	d := svcDeps{
		buildingPrice: &mocks.MockRepositoryBuildingPrice{},
		quotation:     &mocks.MockRepositoryQuotation{},
		rateCard:      &mocks.MockRepositoryRateCard{},
		building:      &mocks.MockRepositoryBuilding{},
		customer:      &mocks.MockRepositoryCustomer{},
		brand:         &mocks.MockRepositoryBrand{},
		user:          &mocks.MockRepositoryUser{},
		assignment:    &mocks.MockRepositorySalesAssignment{},
		pkg:           &mocks.MockRepositorySalesPackage{},
	}

	sqlMock.ExpectBegin()
	if commits {
		sqlMock.ExpectCommit()
	} else {
		sqlMock.ExpectRollback()
	}

	svc := service.NewServiceQuotationImpl(db, d.quotation, d.buildingPrice, d.building,
		d.customer, d.brand, d.user, d.assignment, d.pkg)

	return svc, d, func() { assert.NoError(t, sqlMock.ExpectationsWereMet()) }
}

func salesActor(id int) service.Actor {
	return service.Actor{UserId: id, Role: models.RoleSales}
}

func draftQuotation(id, ownerId int) models.Quotation {
	return models.Quotation{
		Id: id, QuoteNumber: "Q-2026-0001", SalesUserId: ownerId, CreatedByUserId: ownerId,
		Status: service.StatusDraft, Discount: 50, TaxRate: service.DefaultTaxRate, Version: 1,
		Selections: []models.QuotationSelection{{
			Kind: models.SelectionKindPlacement, Mode: models.SelectionModeBuilding, Weeks: 4,
			Items: []models.QuotationSelectionItem{{BuildingId: 1}},
		}},
	}
}

// ---------------------------------------------------------------------------
// Visibility
// ---------------------------------------------------------------------------

// A salesperson must not be able to read someone else's pipeline by guessing an id.
func TestFindById_RefusesAnotherSalespersonsQuotation(t *testing.T) {
	svc, d, assertMock := newQuotationService(t, false)

	d.quotation.On("FindById", mock.Anything, mock.Anything, 5).
		Return(draftQuotation(5, 111), nil)
	d.quotation.On("FindApprovals", mock.Anything, mock.Anything, 5).
		Return([]models.QuotationApproval{}, nil)

	assert.PanicsWithValue(t,
		exceptions.NewForbidden("this quotation belongs to someone else"),
		func() { svc.FindById(context.Background(), 5, salesActor(222)) })

	assertMock()
}

// Approving clears required_approver_user_id, so the approver stops being the
// ASSIGNED approver the moment they decide. They must still be able to open the
// quotation they just signed off -- otherwise the approve call succeeds and the
// screen that reloads behind it fails with "this quotation belongs to someone else".
func TestFindById_AllowsAnApproverAfterTheyHaveApproved(t *testing.T) {
	svc, d, assertMock := newQuotationService(t, true)

	q := draftQuotation(5, 111)
	q.Status = service.StatusApproved
	q.RequiredApproverUserId = 0

	d.quotation.On("FindById", mock.Anything, mock.Anything, 5).Return(q, nil)
	d.quotation.On("FindApprovals", mock.Anything, mock.Anything, 5).
		Return([]models.QuotationApproval{
			{QuotationId: 5, Version: 1, ActorUserId: 111, Action: models.ApprovalActionSubmitted},
			{QuotationId: 5, Version: 1, ActorUserId: 222, Action: models.ApprovalActionApproved},
		}, nil)

	resp := svc.FindById(context.Background(), 5, service.Actor{UserId: 222, Role: models.RoleHeadOfSales})

	assert.Equal(t, 5, resp.Id)
	assertMock()
}

// The same applies after returning one: the reason they wrote is theirs to re-read.
func TestFindById_AllowsAnApproverAfterTheyHaveReturned(t *testing.T) {
	svc, d, assertMock := newQuotationService(t, true)

	q := draftQuotation(5, 111)
	q.Status = service.StatusReturned
	q.RequiredApproverUserId = 0

	d.quotation.On("FindById", mock.Anything, mock.Anything, 5).Return(q, nil)
	d.quotation.On("FindApprovals", mock.Anything, mock.Anything, 5).
		Return([]models.QuotationApproval{
			{QuotationId: 5, Version: 1, ActorUserId: 222, Action: models.ApprovalActionReturned, Comment: "too deep"},
		}, nil)

	resp := svc.FindById(context.Background(), 5, service.Actor{UserId: 222, Role: models.RoleHeadOfSales})

	assert.Equal(t, 5, resp.Id)
	assertMock()
}

// An unrelated approver is still refused: acting on it is what grants access, not
// merely holding an approver role.
func TestFindById_RefusesAnApproverWhoNeverActed(t *testing.T) {
	svc, d, assertMock := newQuotationService(t, false)

	q := draftQuotation(5, 111)
	q.Status = service.StatusApproved
	q.RequiredApproverUserId = 0

	d.quotation.On("FindById", mock.Anything, mock.Anything, 5).Return(q, nil)
	d.quotation.On("FindApprovals", mock.Anything, mock.Anything, 5).
		Return([]models.QuotationApproval{
			{QuotationId: 5, Version: 1, ActorUserId: 222, Action: models.ApprovalActionApproved},
		}, nil)

	assert.PanicsWithValue(t,
		exceptions.NewForbidden("this quotation belongs to someone else"),
		func() {
			svc.FindById(context.Background(), 5, service.Actor{UserId: 333, Role: models.RoleCEO})
		})

	assertMock()
}

// The named approver must be able to read a quotation they did not write, or they
// cannot decide on it.
func TestFindById_AllowsTheAssignedApprover(t *testing.T) {
	svc, d, assertMock := newQuotationService(t, true)

	q := draftQuotation(5, 111)
	q.Status = service.StatusPendingManager
	q.RequiredApproverUserId = 222

	d.quotation.On("FindById", mock.Anything, mock.Anything, 5).Return(q, nil)
	d.quotation.On("FindApprovals", mock.Anything, mock.Anything, 5).
		Return([]models.QuotationApproval{}, nil)

	resp := svc.FindById(context.Background(), 5, service.Actor{UserId: 222, Role: models.RoleHeadOfSales})

	assert.Equal(t, 5, resp.Id)
	assertMock()
}

// ---------------------------------------------------------------------------
// Editing
// ---------------------------------------------------------------------------

// Once submitted, a quotation is out of the seller's hands until it comes back.
func TestUpdate_RefusedOncePending(t *testing.T) {
	svc, d, assertMock := newQuotationService(t, false)

	q := draftQuotation(5, 111)
	q.Status = service.StatusPendingManager
	d.quotation.On("FindById", mock.Anything, mock.Anything, 5).Return(q, nil)

	assert.Panics(t, func() {
		svc.Delete(context.Background(), 5, salesActor(111))
	})
	d.quotation.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything, mock.Anything)
	assertMock()
}

func TestDelete_RefusesAnotherOwnersDraft(t *testing.T) {
	svc, d, assertMock := newQuotationService(t, false)

	d.quotation.On("FindById", mock.Anything, mock.Anything, 5).Return(draftQuotation(5, 111), nil)

	assert.PanicsWithValue(t,
		exceptions.NewForbidden("this quotation belongs to someone else"),
		func() { svc.Delete(context.Background(), 5, salesActor(222)) })

	assertMock()
}

// ---------------------------------------------------------------------------
// Submit
// ---------------------------------------------------------------------------

// Bonus is free, so a quotation of bonus alone would be worth nothing.
func TestSubmit_RefusesBonusWithoutPlacement(t *testing.T) {
	svc, d, assertMock := newQuotationService(t, false)

	q := draftQuotation(5, 111)
	q.Selections = []models.QuotationSelection{{
		Kind: models.SelectionKindBonus, Mode: models.SelectionModeBuilding, Weeks: 4,
	}}
	d.quotation.On("FindById", mock.Anything, mock.Anything, 5).Return(q, nil)

	assert.PanicsWithValue(t,
		exceptions.NewBadRequestError("a quotation needs a placement; bonus alone cannot be sold"),
		func() { svc.Submit(context.Background(), 5, salesActor(111)) })

	assertMock()
}

func TestSubmit_RefusesEmptyQuotation(t *testing.T) {
	svc, d, assertMock := newQuotationService(t, false)

	q := draftQuotation(5, 111)
	q.Selections = nil
	d.quotation.On("FindById", mock.Anything, mock.Anything, 5).Return(q, nil)

	assert.PanicsWithValue(t,
		exceptions.NewBadRequestError("this quotation has nothing selected"),
		func() { svc.Submit(context.Background(), 5, salesActor(111)) })

	assertMock()
}

// ---------------------------------------------------------------------------
// Approve and return
// ---------------------------------------------------------------------------

// Holding an approver role is not enough: routing named one person.
func TestApprove_RefusesAnyoneButTheNamedApprover(t *testing.T) {
	svc, d, assertMock := newQuotationService(t, false)

	q := draftQuotation(5, 111)
	q.Status = service.StatusPendingManager
	q.RequiredApproverUserId = 222
	d.quotation.On("FindById", mock.Anything, mock.Anything, 5).Return(q, nil)

	assert.PanicsWithValue(t,
		exceptions.NewForbidden("this quotation is waiting on a different approver"),
		func() {
			svc.Approve(context.Background(), 5, service.Actor{UserId: 333, Role: models.RoleCEO})
		})

	d.quotation.AssertNotCalled(t, "UpdatePricingAndStatus", mock.Anything, mock.Anything, mock.Anything)
	assertMock()
}

func TestApprove_RefusesAQuotationThatIsNotPending(t *testing.T) {
	svc, d, assertMock := newQuotationService(t, false)

	d.quotation.On("FindById", mock.Anything, mock.Anything, 5).Return(draftQuotation(5, 111), nil)

	assert.PanicsWithValue(t,
		exceptions.NewBadRequestError("this quotation is draft and is not awaiting approval"),
		func() {
			svc.Approve(context.Background(), 5, service.Actor{UserId: 222, Role: models.RoleHeadOfSales})
		})

	assertMock()
}

// The spec forbids returning without a reason outright. The service refuses before
// the database CHECK would, so the message is readable.
func TestReturn_RequiresAComment(t *testing.T) {
	svc, _, assertMock := newQuotationService(t, false)

	assert.PanicsWithValue(t,
		exceptions.NewBadRequestError("a reason is required when returning a quotation"),
		func() {
			svc.Return(context.Background(), 5, "   ",
				service.Actor{UserId: 222, Role: models.RoleHeadOfSales})
		})

	assertMock()
}

func TestReturn_SendsItBackToTheOwner(t *testing.T) {
	svc, d, assertMock := newQuotationService(t, true)

	q := draftQuotation(5, 111)
	q.Status = service.StatusPendingManager
	q.RequiredApproverUserId = 222

	returned := q
	returned.Status = service.StatusReturned

	d.quotation.On("FindById", mock.Anything, mock.Anything, 5).Return(q, nil).Once()

	var saved models.Quotation
	d.quotation.On("UpdatePricingAndStatus", mock.Anything, mock.Anything, mock.AnythingOfType("models.Quotation")).
		Run(func(args mock.Arguments) { saved = args.Get(2).(models.Quotation) }).
		Return(nil)

	var recorded models.QuotationApproval
	d.quotation.On("CreateApproval", mock.Anything, mock.Anything, mock.AnythingOfType("models.QuotationApproval")).
		Run(func(args mock.Arguments) { recorded = args.Get(2).(models.QuotationApproval) }).
		Return(nil)

	d.quotation.On("FindById", mock.Anything, mock.Anything, 5).Return(returned, nil)

	svc.Return(context.Background(), 5, "Discount too aggressive for this account",
		service.Actor{UserId: 222, Role: models.RoleHeadOfSales})

	assert.Equal(t, service.StatusReturned, saved.Status)
	assert.Equal(t, 0, saved.RequiredApproverUserId, "no longer waiting on anyone")
	assert.Equal(t, models.ApprovalActionReturned, recorded.Action)
	assert.Equal(t, "Discount too aggressive for this account", recorded.Comment)
	assertMock()
}

// ---------------------------------------------------------------------------
// Dashboard
// ---------------------------------------------------------------------------

func TestDashboardCounts_GroupsThePendingStatuses(t *testing.T) {
	svc, d, assertMock := newQuotationService(t, true)

	d.quotation.On("CountByStatusForOwner", mock.Anything, mock.Anything, 111).
		Return(map[string]int{
			service.StatusDraft:                  2,
			service.StatusPendingManager:         3,
			service.StatusPendingBusinessControl: 1,
			service.StatusPendingCEO:             1,
			service.StatusReturned:               4,
			service.StatusApproved:               5,
		}, nil)

	counts := svc.DashboardCounts(context.Background(), salesActor(111))

	assert.Equal(t, 2, counts.Draft)
	assert.Equal(t, 5, counts.Pending, "the three pending statuses are one number to a seller")
	assert.Equal(t, 4, counts.Returned)
	assert.Equal(t, 5, counts.Approved)
	assert.Equal(t, 16, counts.All)
	assertMock()
}
