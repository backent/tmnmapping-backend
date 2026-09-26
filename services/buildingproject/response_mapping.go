package buildingproject

import (
	"github.com/malikabdulaziz/tmn-backend/models"
	webBuildingProject "github.com/malikabdulaziz/tmn-backend/web/buildingproject"
)

// projectToResponse maps a project for the wire.
//
// The finance gate is applied in the REPOSITORY -- those columns are not selected for
// a caller without the permission -- and this is the second line, not the first. The
// flag is passed here so the pointers stay nil rather than pointing at a zero the
// caller could mistake for a real rental of nothing.
func projectToResponse(project models.BuildingProject, includeFinance bool) webBuildingProject.BuildingProjectResponse {
	response := webBuildingProject.BuildingProjectResponse{
		Id:               project.Id,
		ProjectIdIris:    project.ProjectIdIris,
		Name:             project.Name,
		BuildingType:     project.BuildingType,
		Grade:            project.Grade,
		Pic:              project.Pic,
		TmnProjectStatus: project.TmnProjectStatus,
		NoOfTower:        project.NoOfTower,
		NoOfScreen:       project.NoOfScreen,
		CreatedDate:      project.CreatedDate,
		Remark:           project.Remark,
		ContractType:     project.ContractType,
		ContractDate:     project.ContractDate,
		ContractStart:    project.ContractStart,
		ContractEnd:      project.ContractEnd,
		PeriodMonth:      project.PeriodMonth,
		PaymentTerm:      project.PaymentTerm,
		Exclusivity:      project.Exclusivity,
		DocType:          project.DocType,
		ContractStatus:   project.ContractStatus,
		CancelledAt:      project.CancelledAt,
		CancelLastStatus: project.CancelLastStatus,
		CancelReason:     project.CancelReason,
		BuildingCount:    project.BuildingCount,
		CreatedAt:        project.CreatedAt,
		UpdatedAt:        project.UpdatedAt,
	}

	if !includeFinance {
		return response
	}

	rental := project.AnnualRental
	company := project.CompanyName
	contractNo := project.ContractNo
	response.AnnualRental = &rental
	response.CompanyName = &company
	response.ContractNo = &contractNo

	// Derived on read, never stored: a hand-edited or stale figure in a spreadsheet
	// cannot enter the database, and correcting the rental corrects both at once.
	if pricePerScreen, ok := PricePerScreen(project); ok {
		response.PricePerScreen = &pricePerScreen
	}
	if contractValue, ok := ContractValue(project); ok {
		response.ContractValue = &contractValue
	}

	return response
}

// PricePerScreen is annual rental divided by screens -- rupiah per screen per YEAR.
// The source workbook multiplied this into monthly cells, which made twelve months
// sum to twelve times the year; that projection is out of scope, and the figure here
// is the unambiguous annual one.
func PricePerScreen(project models.BuildingProject) (int64, bool) {
	if project.AnnualRental == 0 || project.NoOfScreen == 0 {
		return 0, false
	}

	return project.AnnualRental / int64(project.NoOfScreen), true
}

// ContractValue is the whole contract: annual rental over its actual length.
func ContractValue(project models.BuildingProject) (int64, bool) {
	if project.AnnualRental == 0 || project.PeriodMonth == 0 {
		return 0, false
	}

	return project.AnnualRental * int64(project.PeriodMonth) / 12, true
}

// changeToResponse maps one history row.
//
// A change to a finance column carries the rental in old_value and new_value, so the
// row is exactly as sensitive as the column. Without the permission the field is still
// named -- knowing the rental was changed, by whom and when, is not itself secret --
// but the values are withheld.
func changeToResponse(change models.BuildingProjectChange, includeFinance bool) webBuildingProject.BuildingProjectChangeResponse {
	response := webBuildingProject.BuildingProjectChangeResponse{
		Id:            change.Id,
		ProjectId:     change.ProjectId,
		ProjectIdIris: change.ProjectIdIris,
		ActorUserId:   change.ActorUserId,
		ActorName:     change.ActorName,
		ActorRole:     change.ActorRole,
		Action:        change.Action,
		Source:        change.Source,
		BatchId:       change.BatchId,
		Field:         change.Field,
		OldValue:      change.OldValue,
		NewValue:      change.NewValue,
		CreatedAt:     change.CreatedAt,
	}

	if IsFinanceField(change.Field) && !includeFinance {
		response.OldValue = ""
		response.NewValue = ""
	}

	return response
}
