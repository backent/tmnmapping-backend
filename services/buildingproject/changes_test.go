package buildingproject_test

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/malikabdulaziz/tmn-backend/models"
	"github.com/malikabdulaziz/tmn-backend/services/buildingproject"
)

func sampleProject() models.BuildingProject {
	return models.BuildingProject{
		Id:               7,
		ProjectIdIris:    "PRJ-0001",
		Name:             "Menara Arunika",
		BuildingType:     "Office",
		Grade:            "Grade A",
		Pic:              "Dara",
		TmnProjectStatus: "Active",
		NoOfTower:        1,
		NoOfScreen:       6,
		CreatedDate:      "2026-01-05",
		ContractType:     "Initial",
		ContractNo:       "CON/2026/001",
		ContractStart:    "2026-02-01",
		ContractEnd:      "2027-01-31",
		PeriodMonth:      12,
		AnnualRental:     24000000,
		PaymentTerm:      "Monthly",
		CompanyName:      "PT Arunika",
		Exclusivity:      "Non-Exclusive",
		DocType:          "PKS",
		ContractStatus:   "Signed",
	}
}

var testActor = buildingproject.Actor{UserId: 3, Role: models.RoleAdmin}

// An unchanged save must leave no trace. This is what keeps the history readable, and
// it is also the round-trip guarantee the import depends on: re-uploading an unmodified
// export must log nothing.
func TestDiffProjects_NoChangesLogsNothing(t *testing.T) {
	project := sampleProject()

	changes := buildingproject.DiffProjects(project, project, testActor, models.BuildingProjectSourceForm, "")

	assert.Empty(t, changes)
}

func TestDiffProjects_OneFieldOneRow(t *testing.T) {
	before := sampleProject()
	after := sampleProject()
	after.AnnualRental = 31500000

	changes := buildingproject.DiffProjects(before, after, testActor, models.BuildingProjectSourceForm, "")

	assert.Len(t, changes, 1)
	assert.Equal(t, "annual_rental", changes[0].Field)
	assert.Equal(t, "24000000", changes[0].OldValue)
	assert.Equal(t, "31500000", changes[0].NewValue)
	assert.Equal(t, models.BuildingProjectActionUpdated, changes[0].Action)
	assert.Equal(t, models.BuildingProjectSourceForm, changes[0].Source)
	assert.Equal(t, 3, changes[0].ActorUserId)
	assert.Equal(t, models.RoleAdmin, changes[0].ActorRole)
	assert.Empty(t, changes[0].BatchId, "a form edit carries no batch id")
}

// Clearing is a change like any other, and must be recorded with what was lost --
// the change log is the only undo trail after a destructive upload.
func TestDiffProjects_ClearingIsRecordedWithTheOldValue(t *testing.T) {
	before := sampleProject()
	after := sampleProject()
	after.CompanyName = ""

	changes := buildingproject.DiffProjects(before, after, testActor, models.BuildingProjectSourceImport, "batch-1")

	assert.Len(t, changes, 1)
	assert.Equal(t, "company_name", changes[0].Field)
	assert.Equal(t, "PT Arunika", changes[0].OldValue)
	assert.Equal(t, "", changes[0].NewValue)
	assert.Equal(t, "batch-1", changes[0].BatchId)
}

// A nullable count read back as 0 must not look like a change from empty. Saving an
// untouched project would otherwise log no_of_tower every single time.
func TestDiffProjects_ZeroCountIsNotAChangeFromEmpty(t *testing.T) {
	before := sampleProject()
	before.NoOfTower = 0
	after := sampleProject()
	after.NoOfTower = 0

	changes := buildingproject.DiffProjects(before, after, testActor, models.BuildingProjectSourceForm, "")

	assert.Empty(t, changes)
}

func TestDiffProjects_RenamedKeyFilesUnderTheNewValue(t *testing.T) {
	before := sampleProject()
	after := sampleProject()
	after.ProjectIdIris = "PRJ-0002"

	changes := buildingproject.DiffProjects(before, after, testActor, models.BuildingProjectSourceForm, "")

	assert.Len(t, changes, 1)
	assert.Equal(t, "PRJ-0002", changes[0].ProjectIdIris)
	assert.Equal(t, "PRJ-0001", changes[0].OldValue)
}

// Every editable column on the model must be tracked. Adding a column to
// BuildingProject without adding it to trackedFields would mean edits to it go
// silently unrecorded -- exactly the failure the change log exists to prevent.
func TestDiffProjects_EveryEditableFieldIsTracked(t *testing.T) {
	// Columns that are not user-editable state, so not tracked.
	untracked := map[string]bool{
		"Id": true, "BuildingCount": true, "CreatedAt": true, "UpdatedAt": true,
	}

	before := models.BuildingProject{}
	value := reflect.ValueOf(&before).Elem()
	fieldType := value.Type()

	for i := 0; i < value.NumField(); i++ {
		name := fieldType.Field(i).Name
		if untracked[name] {
			continue
		}

		after := models.BuildingProject{}
		target := reflect.ValueOf(&after).Elem().Field(i)

		switch target.Kind() {
		case reflect.String:
			target.SetString("changed")
		case reflect.Int:
			target.SetInt(99)
		case reflect.Int64:
			target.SetInt(99)
		default:
			t.Fatalf("field %s has unhandled kind %s", name, target.Kind())
		}

		changes := buildingproject.DiffProjects(before, after, testActor, models.BuildingProjectSourceForm, "")
		assert.Len(t, changes, 1,
			"changing %s produced no change row: add it to trackedFields in changes.go", name)
	}
}

func TestIsFinanceField(t *testing.T) {
	assert.True(t, buildingproject.IsFinanceField("annual_rental"))
	assert.True(t, buildingproject.IsFinanceField("company_name"))
	assert.True(t, buildingproject.IsFinanceField("contract_no"))
	assert.False(t, buildingproject.IsFinanceField("name"))
	assert.False(t, buildingproject.IsFinanceField("contract_status"))
}

func TestCreationAndDeletionChanges(t *testing.T) {
	project := sampleProject()

	created := buildingproject.CreationChange(project, testActor, models.BuildingProjectSourceForm, "")
	assert.Equal(t, models.BuildingProjectActionCreated, created.Action)
	assert.Empty(t, created.Field, "creation describes the whole row, not one field")
	assert.Equal(t, "Menara Arunika", created.NewValue)

	deleted := buildingproject.DeletionChange(project, testActor, models.BuildingProjectSourceForm, "")
	assert.Equal(t, models.BuildingProjectActionDeleted, deleted.Action)
	assert.Equal(t, "Menara Arunika", deleted.OldValue,
		"the name is the only evidence left once project_id is nulled")
	assert.Equal(t, "PRJ-0001", deleted.ProjectIdIris)
}
