package buildingproject

import (
	"strconv"

	"github.com/malikabdulaziz/tmn-backend/models"
)

// Actor is the caller, resolved by RequireAuth. Both the id and the role are stored
// on every change row: the role is recorded AS IT WAS, because roles change and an
// audit trail that changes with them is not an audit trail.
type Actor struct {
	UserId int
	Role   string
}

// trackedFields lists every column a change is recorded for, in the order a reader
// wants them: identity, then classification, then the contract, then cancellation.
//
// Adding a column to building_projects without adding it here means edits to it go
// unrecorded. TestTrackedFieldsCoversModel exists to make that impossible to forget.
var trackedFields = []string{
	"project_id_iris",
	"name",
	"building_type",
	"grade",
	"pic",
	"tmn_project_status",
	"no_of_tower",
	"no_of_screen",
	"created_date",
	"remark",
	"contract_type",
	"contract_no",
	"contract_date",
	"contract_start",
	"contract_end",
	"period_month",
	"annual_rental",
	"payment_term",
	"company_name",
	"exclusivity",
	"doc_type",
	"contract_status",
	"cancelled_at",
	"cancel_last_status",
	"cancel_reason",
}

// financeFields are the columns gated by building-projects.finance. A change row for
// one of these carries the value in old_value / new_value, so the row is exactly as
// sensitive as the column.
var financeFields = map[string]bool{
	"annual_rental": true,
	"company_name":  true,
	"contract_no":   true,
}

// IsFinanceField reports whether field holds landlord contract money or counterparty.
func IsFinanceField(field string) bool { return financeFields[field] }

// fieldValues flattens a project into the string form stored in the change log.
// Everything is compared as a string because that is what old_value / new_value hold;
// comparing typed values first would mean two notions of equality to keep in step.
func fieldValues(p models.BuildingProject) map[string]string {
	return map[string]string{
		"project_id_iris":    p.ProjectIdIris,
		"name":               p.Name,
		"building_type":      p.BuildingType,
		"grade":              p.Grade,
		"pic":                p.Pic,
		"tmn_project_status": p.TmnProjectStatus,
		"no_of_tower":        intToLog(p.NoOfTower),
		"no_of_screen":       intToLog(p.NoOfScreen),
		"created_date":       p.CreatedDate,
		"remark":             p.Remark,
		"contract_type":      p.ContractType,
		"contract_no":        p.ContractNo,
		"contract_date":      p.ContractDate,
		"contract_start":     p.ContractStart,
		"contract_end":       p.ContractEnd,
		"period_month":       intToLog(p.PeriodMonth),
		"annual_rental":      int64ToLog(p.AnnualRental),
		"payment_term":       p.PaymentTerm,
		"company_name":       p.CompanyName,
		"exclusivity":        p.Exclusivity,
		"doc_type":           p.DocType,
		"contract_status":    p.ContractStatus,
		"cancelled_at":       p.CancelledAt,
		"cancel_last_status": p.CancelLastStatus,
		"cancel_reason":      p.CancelReason,
	}
}

// intToLog renders 0 as empty rather than "0". A missing count and a count of zero
// are the same thing here -- the column is nullable and the model has no null int --
// so logging "0" would report a change every time an empty field was saved untouched.
func intToLog(v int) string {
	if v == 0 {
		return ""
	}

	return strconv.Itoa(v)
}

func int64ToLog(v int64) string {
	if v == 0 {
		return ""
	}

	return strconv.FormatInt(v, 10)
}

// DiffProjects returns one change row per field that actually differs.
//
// Unchanged fields produce nothing. That is what keeps the log readable: an edit that
// corrects one typo is one row, not twenty-five, and a re-upload of an unmodified
// export produces none at all -- which is the round-trip test the import relies on.
func DiffProjects(before models.BuildingProject, after models.BuildingProject, actor Actor, source string, batchId string) []models.BuildingProjectChange {
	oldValues := fieldValues(before)
	newValues := fieldValues(after)

	// The key can itself be edited, so the row is filed under where it ended up.
	iris := after.ProjectIdIris
	if iris == "" {
		iris = before.ProjectIdIris
	}

	changes := make([]models.BuildingProjectChange, 0, len(trackedFields))
	for _, field := range trackedFields {
		if oldValues[field] == newValues[field] {
			continue
		}

		changes = append(changes, models.BuildingProjectChange{
			ProjectId:     after.Id,
			ProjectIdIris: iris,
			ActorUserId:   actor.UserId,
			ActorRole:     actor.Role,
			Action:        models.BuildingProjectActionUpdated,
			Source:        source,
			BatchId:       batchId,
			Field:         field,
			OldValue:      oldValues[field],
			NewValue:      newValues[field],
		})
	}

	return changes
}

// CreationChange is the single row recording that a project came into existence.
//
// One row, not one per populated field: at creation every field is "new", so a
// per-field log would bury the one fact that matters. The values are recoverable from
// the row itself for as long as it exists, and from the deletion row after that.
func CreationChange(created models.BuildingProject, actor Actor, source string, batchId string) models.BuildingProjectChange {
	return models.BuildingProjectChange{
		ProjectId:     created.Id,
		ProjectIdIris: created.ProjectIdIris,
		ActorUserId:   actor.UserId,
		ActorRole:     actor.Role,
		Action:        models.BuildingProjectActionCreated,
		Source:        source,
		BatchId:       batchId,
		NewValue:      created.Name,
	}
}

// DeletionChange records a project being removed. The name is kept in old_value
// because the row it names is about to stop existing, and project_id will be nulled
// by the foreign key -- leaving project_id_iris and this value as the only evidence.
func DeletionChange(deleted models.BuildingProject, actor Actor, source string, batchId string) models.BuildingProjectChange {
	return models.BuildingProjectChange{
		ProjectId:     deleted.Id,
		ProjectIdIris: deleted.ProjectIdIris,
		ActorUserId:   actor.UserId,
		ActorRole:     actor.Role,
		Action:        models.BuildingProjectActionDeleted,
		Source:        source,
		BatchId:       batchId,
		OldValue:      deleted.Name,
	}
}
