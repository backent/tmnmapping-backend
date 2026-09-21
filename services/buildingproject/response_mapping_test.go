package buildingproject

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/malikabdulaziz/tmn-backend/models"
)

func mappingSample() models.BuildingProject {
	return models.BuildingProject{
		Id: 7, ProjectIdIris: "PRJ-0001", Name: "Menara Arunika",
		BuildingType: "Office", Grade: "Grade A", Pic: "Dara",
		TmnProjectStatus: "Active", NoOfTower: 1, NoOfScreen: 6,
		CreatedDate: "2026-01-05", Remark: "note",
		ContractType: "Initial", ContractNo: "CON/2026/001",
		ContractDate: "2026-01-18", ContractStart: "2026-02-01", ContractEnd: "2027-01-31",
		PeriodMonth: 12, AnnualRental: 24000000, PaymentTerm: "Monthly",
		CompanyName: "PT Arunika", Exclusivity: "Non-Exclusive", DocType: "PKS",
		ContractStatus: "Signed", CancelledAt: "2026-08-01",
		CancelLastStatus: "Negotiation", CancelReason: "budget",
		BuildingCount: 3, CreatedAt: "2026-01-05T00:00:00Z", UpdatedAt: "2026-01-06T00:00:00Z",
	}
}

// A response mapper that forgets a field is the quiet failure this codebase has
// already shipped once: brand contacts saved correctly and always returned empty,
// because brandToResponse never copied the new columns. Reflection over the response
// struct means a field added later cannot be left behind silently.
func TestProjectToResponse_CarriesEveryField(t *testing.T) {
	response := projectToResponse(mappingSample(), true)

	value := reflect.ValueOf(response)
	fields := value.Type()

	for i := 0; i < value.NumField(); i++ {
		name := fields.Field(i).Name
		field := value.Field(i)

		if field.Kind() == reflect.Ptr {
			assert.False(t, field.IsNil(), "%s should be populated for a finance caller", name)

			continue
		}

		assert.False(t, field.IsZero(), "%s was not carried into the response", name)
	}
}

// Without the permission the three columns are never selected, so the model arrives
// already empty. The pointers must stay nil rather than pointing at a zero, so a
// caller cannot mistake "withheld" for "a rental of nothing".
func TestProjectToResponse_OmitsFinanceWithoutPermission(t *testing.T) {
	response := projectToResponse(mappingSample(), false)

	assert.Nil(t, response.AnnualRental)
	assert.Nil(t, response.CompanyName)
	assert.Nil(t, response.ContractNo)
	assert.Nil(t, response.PricePerScreen, "a derived figure leaks the rental just as well")
	assert.Nil(t, response.ContractValue)

	// Everything else still arrives: the project itself is not secret.
	assert.Equal(t, "Menara Arunika", response.Name)
	assert.Equal(t, "Signed", response.ContractStatus)
	assert.Equal(t, 6, response.NoOfScreen)
	assert.Equal(t, 3, response.BuildingCount)
}

func TestDerivedFigures(t *testing.T) {
	project := mappingSample()

	pricePerScreen, ok := PricePerScreen(project)
	assert.True(t, ok)
	assert.Equal(t, int64(4000000), pricePerScreen, "24,000,000 / 6 screens, per YEAR")

	contractValue, ok := ContractValue(project)
	assert.True(t, ok)
	assert.Equal(t, int64(24000000), contractValue, "24,000,000 x 12 months / 12")

	project.PeriodMonth = 24
	contractValue, ok = ContractValue(project)
	assert.True(t, ok)
	assert.Equal(t, int64(48000000), contractValue, "a two-year contract is worth two years")
}

// Dividing by a missing screen count or period must not panic or report zero as if
// it were a real figure.
func TestDerivedFigures_GuardAgainstMissingDivisor(t *testing.T) {
	project := mappingSample()
	project.NoOfScreen = 0
	_, ok := PricePerScreen(project)
	assert.False(t, ok)

	project = mappingSample()
	project.PeriodMonth = 0
	_, ok = ContractValue(project)
	assert.False(t, ok)

	project = mappingSample()
	project.AnnualRental = 0
	_, ok = PricePerScreen(project)
	assert.False(t, ok, "no rental means no price per screen, not a price of zero")
}

// The history names the field either way -- that the rental changed, and who changed
// it, is not itself secret -- but the values are withheld.
func TestChangeToResponse_WithholdsFinanceValues(t *testing.T) {
	change := models.BuildingProjectChange{
		Field: "annual_rental", OldValue: "24000000", NewValue: "31500000",
		Action: models.BuildingProjectActionUpdated, ActorName: "Dara",
	}

	withPermission := changeToResponse(change, true)
	assert.Equal(t, "24000000", withPermission.OldValue)
	assert.Equal(t, "31500000", withPermission.NewValue)

	without := changeToResponse(change, false)
	assert.Equal(t, "annual_rental", without.Field, "the field is still named")
	assert.Equal(t, "Dara", without.ActorName, "who changed it is still shown")
	assert.Empty(t, without.OldValue)
	assert.Empty(t, without.NewValue)
}

func TestChangeToResponse_NonFinanceIsUntouched(t *testing.T) {
	change := models.BuildingProjectChange{
		Field: "tmn_project_status", OldValue: "Active", NewValue: "Expired",
	}

	without := changeToResponse(change, false)
	assert.Equal(t, "Active", without.OldValue)
	assert.Equal(t, "Expired", without.NewValue)
}
