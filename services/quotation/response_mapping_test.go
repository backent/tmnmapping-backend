package quotation_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/malikabdulaziz/tmn-backend/models"
	service "github.com/malikabdulaziz/tmn-backend/services/quotation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// The brand response once declared the contact fields and never populated them, so
// every screen showed a saved contact as empty. These guard the quotation response
// against the same class of mistake: a field added to the DTO but not to the mapper.
//
// The quotation carries its own copy of the contact -- prefilled from the brand, then
// frozen -- and the printed document renders from it, so a dropped field here would
// silently blank the person the quotation is addressed to.

func TestQuotationResponseCarriesEveryFieldItDeclares(t *testing.T) {
	svc, d, assertMock := newQuotationService(t, true)

	full := models.Quotation{
		Id: 5, QuoteNumber: "Q-2026-0001",
		SalesUserId: 111, SalesName: "Ayu", CreatedByUserId: 112, CreatedByName: "Rina",
		CustomerId: 3, CustomerName: "PT Example", BrandId: 4, BrandName: "Kopi Kenangan",
		RateCardVersionCode: "RC-2026",
		AttentionTo:         "Budi Santoso", JobTitle: "Marketing Director",
		ContactPhone: "+62 812 3456 7890", ContactEmail: "budi@example.com",
		CampaignYear: 2026, ValidUntil: "2026-10-12", Discount: 65, TaxRate: 0.11,
		Status: service.StatusDraft, RequiredApproverName: "Head", Version: 2,
		PlacementGross: 1_520_000_000, PlacementDiscountAmount: 988_000_000,
		PlacementNet: 532_000_000, BonusGross: 280_000_000, BonusNet: 1,
		TotalGross: 1_800_000_000, TotalNet: 532_000_000,
		EffectiveDiscountAmount: 1_268_000_000, EffectiveDiscountRate: 70.44,
		Tax: 58_520_000, TotalIncludingTax: 590_520_000,
		CreatedAt: "2026-06-22", UpdatedAt: "2026-06-23", ApprovedAt: "2026-06-24",
		Selections: []models.QuotationSelection{{
			Kind: models.SelectionKindPlacement, Mode: models.SelectionModeBuilding, Weeks: 4,
		}},
	}

	d.quotation.On("FindById", mock.Anything, mock.Anything, 5).Return(full, nil)
	d.quotation.On("FindApprovals", mock.Anything, mock.Anything, 5).
		Return([]models.QuotationApproval{{
			QuotationId: 5, Version: 1, ActorUserId: 222, ActorName: "Head",
			ActorRole: "head_of_sales", Action: models.ApprovalActionApproved,
			Comment: "fine", CreatedAt: "2026-06-24",
		}}, nil)

	response := svc.FindById(context.Background(), 5, service.Actor{UserId: 111, Role: models.RoleSales})

	value := reflect.ValueOf(response)
	// Every field carries a non-zero value in the fixture, including the two derived
	// booleans: the quotation is a draft (IsEditable) entered by a proxy (IsProxyEntry).
	for i := 0; i < value.NumField(); i++ {
		field := value.Type().Field(i)

		assert.False(t, value.Field(i).IsZero(),
			"%s is declared on QuotationResponse but never populated by toResponse", field.Name)
	}

	assert.Equal(t, "Budi Santoso", response.AttentionTo)
	assert.Equal(t, "budi@example.com", response.ContactEmail)
	assertMock()
}
