package buildingproject

import (
	"context"
	"strconv"
	"strings"

	"github.com/malikabdulaziz/tmn-backend/exceptions"
	"github.com/malikabdulaziz/tmn-backend/helpers"
	"github.com/malikabdulaziz/tmn-backend/models"
	repositoriesBuildingProject "github.com/malikabdulaziz/tmn-backend/repositories/buildingproject"
	"github.com/malikabdulaziz/tmn-backend/spreadsheets"
	"github.com/malikabdulaziz/tmn-backend/web"
)

// Import reads a project workbook. With dryRun it validates and counts, writing
// nothing, so the operator sees what an upload will do -- including what it will
// CLEAR -- before deciding.
//
// A bad row is reported and left out; the valid rows still apply, as in the price
// import. Refusing a whole file for one bad cell would mean a 2,000-row sheet could
// never be uploaded until it was perfect.
func (s *ServiceBuildingProjectImpl) Import(ctx context.Context, fileBytes []byte, fileType string, dryRun bool, actor Actor) web.ImportResult {
	rows, err := spreadsheets.ParseSpreadsheet(fileBytes, fileType)
	if err != nil {
		panic(exceptions.NewBadRequest(err.Error()))
	}
	if len(rows) < 2 {
		panic(exceptions.NewBadRequest("the file has no data rows"))
	}

	colMap := spreadsheets.MapHeaderColumns(rows[0], BuildingProjectColumns)
	if missing := spreadsheets.MissingRequiredColumns(colMap, BuildingProjectColumns); len(missing) > 0 {
		// A file in the old shape lands here, which is the point: rejecting it whole
		// beats importing its totals row as a project.
		panic(exceptions.NewBadRequest("the file is missing required columns: " + strings.Join(missing, ", ")))
	}

	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	result := web.NewImportResult()
	batchId := newBatchId()

	// A project the file names twice is a mistake the operator must resolve: applying
	// both would mean the second silently wins, and which one that is depends on row
	// order nobody thinks about.
	seen := map[string]int{}

	type pending struct {
		rowNumber int
		before    models.BuildingProject
		after     models.BuildingProject
		isNew     bool
	}
	var queue []pending

	for i, row := range rows[1:] {
		rowNumber := i + 2

		if isBlankRow(row) {
			continue
		}

		result.Rows++

		project, rowErrs := parseRow(row, colMap)
		if len(rowErrs) > 0 {
			addRowErrors(result, rowNumber, rowErrs)

			continue
		}

		if firstRow, duplicate := seen[project.ProjectIdIris]; duplicate {
			result.AddError(rowNumber, headerFor("project_id_iris"), project.ProjectIdIris,
				"This project appears twice in this file (also on row "+strconv.Itoa(firstRow)+")")

			continue
		}
		seen[project.ProjectIdIris] = rowNumber

		// Read WITH finance: the comparison must see every column, or an import run
		// by someone without the permission would read the rental as empty and
		// report it as a clear. The values never leave this function.
		existing, err := s.RepositoryBuildingProjectInterface.FindByIris(ctx, tx, project.ProjectIdIris, true)
		isNew := err != nil

		if isNew {
			result.Created++
			queue = append(queue, pending{rowNumber: rowNumber, after: project, isNew: true})

			continue
		}

		project.Id = existing.Id

		// A caller without the finance permission cannot supply those columns -- the
		// export they edited did not contain them -- so their blankness is an
		// artefact of the permission, not an instruction to clear.
		if !canSeeFinance(actor) {
			project.AnnualRental = existing.AnnualRental
			project.CompanyName = existing.CompanyName
			project.ContractNo = existing.ContractNo
		}

		changes := DiffProjects(existing, project, actor, models.BuildingProjectSourceImport, batchId)
		if len(changes) == 0 {
			result.Unchanged++

			continue
		}

		for _, cleared := range clearedFields(existing, project) {
			result.Cleared++

			value := cleared.Old
			if IsFinanceField(cleared.Field) && !canSeeFinance(actor) {
				value = ""
			}

			result.AddNotice(rowNumber, headerFor(cleared.Field), value,
				"This upload clears "+headerFor(cleared.Field))
		}

		result.Updated++
		queue = append(queue, pending{rowNumber: rowNumber, before: existing, after: project})
	}

	if dryRun {
		result.DryRun = true

		return *result
	}

	for _, item := range queue {
		if item.isNew {
			created, err := s.RepositoryBuildingProjectInterface.Create(ctx, tx, item.after)
			helpers.PanicIfError(err)

			helpers.PanicIfError(s.RepositoryBuildingProjectInterface.RecordChanges(ctx, tx,
				[]models.BuildingProjectChange{
					CreationChange(created, actor, models.BuildingProjectSourceImport, batchId),
				}))

			continue
		}

		_, err := s.RepositoryBuildingProjectInterface.Update(ctx, tx, item.after)
		helpers.PanicIfError(err)

		helpers.PanicIfError(s.RepositoryBuildingProjectInterface.RecordChanges(ctx, tx,
			DiffProjects(item.before, item.after, actor, models.BuildingProjectSourceImport, batchId)))
	}

	result.Imported = true

	return *result
}

// Export writes the current table in the template's shape, so an export can be
// edited and uploaded back. The key column is always present: an export without it
// could only ever create duplicates on re-upload.
//
// A caller without the finance permission gets the file with those three columns
// blank rather than absent. Absent would break the round trip -- the importer would
// read a missing column as nothing to say -- while blank plus the permission check on
// the way back in leaves the stored values untouched.
func (s *ServiceBuildingProjectImpl) Export(ctx context.Context, actor Actor) ([]byte, error) {
	tx, err := s.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer helpers.CommitOrRollback(tx)

	finance := canSeeFinance(actor)
	list, err := s.RepositoryBuildingProjectInterface.FindAll(ctx, tx, 100000, 0, "project_id_iris", "ASC",
		repositoriesBuildingProject.ListFilter{}, finance)
	if err != nil {
		return nil, err
	}

	rows := make([][]interface{}, len(list))
	for i, project := range list {
		rows[i] = []interface{}{
			project.ProjectIdIris, project.Name, project.BuildingType, project.Grade,
			project.Pic, project.TmnProjectStatus,
			blankIfZero(project.NoOfTower), blankIfZero(project.NoOfScreen),
			project.CreatedDate, project.Remark,
			project.ContractType, project.ContractNo, project.ContractDate,
			project.ContractStart, project.ContractEnd, blankIfZero(project.PeriodMonth),
			blankIfZero64(project.AnnualRental), project.PaymentTerm, project.CompanyName,
			project.Exclusivity, project.DocType, project.ContractStatus,
			project.CancelledAt, project.CancelLastStatus, project.CancelReason,
		}
	}

	return spreadsheets.BuildExport(BuildingProjectSheetName, TemplateHeaders(), rows)
}

// Template is the empty workbook the operator downloads before filling one in.
func (s *ServiceBuildingProjectImpl) Template() ([]byte, error) {
	return BuildTemplate()
}

// blankIfZero writes an unset count as an empty cell rather than 0. Exporting 0 and
// re-importing it would turn "never filled in" into a real zero on the way back.
func blankIfZero(v int) interface{} {
	if v == 0 {
		return ""
	}

	return v
}

func blankIfZero64(v int64) interface{} {
	if v == 0 {
		return ""
	}

	return v
}
