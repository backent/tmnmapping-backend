package models_test

import (
	"database/sql"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/malikabdulaziz/tmn-backend/models"
)

// A converter that forgets a field is a bug this codebase has now shipped twice:
// brandToResponse dropped the new contact columns in September, and
// NullAbleBuildingToBuilding dropped project_id and project_id_iris the moment they
// were added -- scanned from the query, never copied out, so a linked building read
// back as unlinked.
//
// Reflection over the struct means the next field added cannot be left behind.
func TestNullAbleBuildingToBuilding_CarriesEveryField(t *testing.T) {
	// Fields that are genuinely not part of the nullable row.
	derivedElsewhere := map[string]bool{
		// Joined from building_projects by the caller, not a buildings column.
		"ProjectDisplayName": true,
		// Parsed from JSON rather than copied, and covered below.
		"Images": true,
	}

	nullable := models.NullAbleBuilding{}
	value := reflect.ValueOf(&nullable).Elem()
	fields := value.Type()

	for i := 0; i < value.NumField(); i++ {
		name := fields.Field(i).Name
		if derivedElsewhere[name] {
			continue
		}

		field := value.Field(i)
		switch field.Interface().(type) {
		case sql.NullString:
			field.Set(reflect.ValueOf(sql.NullString{String: "set", Valid: true}))
		case sql.NullInt64:
			field.Set(reflect.ValueOf(sql.NullInt64{Int64: 42, Valid: true}))
		case sql.NullFloat64:
			field.Set(reflect.ValueOf(sql.NullFloat64{Float64: 1.5, Valid: true}))
		case sql.NullBool:
			field.Set(reflect.ValueOf(sql.NullBool{Bool: true, Valid: true}))
		}
	}

	building := models.NullAbleBuildingToBuilding(nullable)

	result := reflect.ValueOf(building)
	resultFields := result.Type()

	for i := 0; i < result.NumField(); i++ {
		name := resultFields.Field(i).Name
		if derivedElsewhere[name] {
			continue
		}

		assert.False(t, result.Field(i).IsZero(),
			"%s was scanned from the row but not copied into the model", name)
	}
}

func TestNullAbleBuildingToBuilding_ImagesAlwaysDecodeToASlice(t *testing.T) {
	// A null images column must not produce a nil slice: the JSON response would
	// then carry null where every consumer expects [].
	building := models.NullAbleBuildingToBuilding(models.NullAbleBuilding{})

	assert.NotNil(t, building.Images)
	assert.Empty(t, building.Images)
}
