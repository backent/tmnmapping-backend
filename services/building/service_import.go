package building

import (
	"context"
	"database/sql"
	"strconv"
	"strings"

	"github.com/malikabdulaziz/tmn-backend/exceptions"
	"github.com/malikabdulaziz/tmn-backend/helpers"
	"github.com/malikabdulaziz/tmn-backend/models"
	"github.com/malikabdulaziz/tmn-backend/spreadsheets"
	"github.com/malikabdulaziz/tmn-backend/web"
)

// Import reads a buildings workbook. With dryRun it validates and counts, writing
// nothing, so the operator sees what an upload will do -- including what it will
// CLEAR -- before deciding.
//
// A bad row is reported and left out; the valid rows still apply, as in the price and
// project imports. Refusing a 3,771-row file for one bad cell would mean it could
// never be uploaded until it was perfect.
func (service *ServiceBuildingImpl) Import(ctx context.Context, fileBytes []byte, fileType string, dryRun bool, actor Actor) web.ImportResult {
	rows, err := spreadsheets.ParseSpreadsheet(fileBytes, fileType)
	if err != nil {
		panic(exceptions.NewBadRequest(err.Error()))
	}
	if len(rows) < 2 {
		panic(exceptions.NewBadRequest("the file has no data rows"))
	}

	colMap := spreadsheets.MapHeaderColumns(rows[0], BuildingColumns)
	if missing := spreadsheets.MissingRequiredColumns(colMap, BuildingColumns); len(missing) > 0 {
		panic(exceptions.NewBadRequest("the file is missing required columns: " + strings.Join(missing, ", ")))
	}

	tx, err := service.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	result := web.NewImportResult()
	batchId := newBatchId()

	// A building the file names twice is a mistake the operator must resolve:
	// applying both means the second silently wins, and which that is depends on row
	// order nobody thinks about.
	seenCode := map[string]int{}
	seenIris := map[string]int{}

	type pending struct {
		before models.Building
		after  models.Building
		isNew  bool
	}
	var queue []pending

	// Projects already resolved in this file, so a sheet naming the same project on
	// 31 towers costs one lookup rather than 31.
	projectCache := map[string]int{}

	for i, row := range rows[1:] {
		rowNumber := i + 2

		if isBlankRow(row) {
			continue
		}

		result.Rows++

		building, projectCode, rowErrs, rowWarnings := parseRow(row, colMap)
		if len(rowErrs) > 0 {
			addRowErrors(result, rowNumber, rowErrs)

			continue
		}

		// Accepted, but the operator should see it before applying.
		for _, warning := range rowWarnings {
			result.AddNotice(rowNumber, warning.column, warning.value, warning.message)
		}

		if firstRow, duplicate := seenCode[building.ExternalBuildingId]; duplicate {
			result.AddError(rowNumber, headerFor("external_building_id"), building.ExternalBuildingId,
				"This building appears twice in this file (also on row "+strconv.Itoa(firstRow)+")")

			continue
		}
		seenCode[building.ExternalBuildingId] = rowNumber

		// An IRIS code repeated in the file would pass the unique index one row at a
		// time and fail on the second, halfway through the apply.
		if building.IrisCode != "" {
			if firstRow, duplicate := seenIris[building.IrisCode]; duplicate {
				result.AddError(rowNumber, headerFor("iris_code"), building.IrisCode,
					"This IRIS Code appears twice in this file (also on row "+strconv.Itoa(firstRow)+")")

				continue
			}
			seenIris[building.IrisCode] = rowNumber
		}

		// An unknown project code creates a stub rather than rejecting the row, so the
		// two files can be uploaded in either order. The stub is deliberately thin:
		// it carries the code and nothing else, and the projects import fills it in.
		if projectCode != "" {
			projectId, created := service.resolveProject(ctx, tx, projectCode, projectCache, actor, batchId, dryRun)
			building.ProjectId = projectId
			building.ProjectIdIris = projectCode

			if created {
				result.AddNotice(rowNumber, headerFor("project_id_iris"), projectCode,
					"No project with this code yet, so an empty one will be created")
			}
		}

		existing, err := service.RepositoryBuildingInterface.FindByExternalId(ctx, tx, building.ExternalBuildingId)
		isNew := err != nil

		if isNew {
			result.Created++
			queue = append(queue, pending{after: building, isNew: true})

			continue
		}

		building.Id = existing.Id
		// Photos are not a column in this sheet, so carrying them across means an
		// upload cannot blank what it never offered to edit. project_name likewise:
		// it is the ERP correlation key the LOI dashboard joins on, and the sheet
		// speaks in project CODES instead.
		building.Images = existing.Images
		building.ProjectName = existing.ProjectName

		// A blank Project ID IRIS clears the link, like every other blank cell.
		if projectCode == "" {
			building.ProjectId = 0
			building.ProjectIdIris = ""
		}

		changes := DiffBuildings(existing, building, actor, models.BuildingSourceImport, batchId)
		if len(changes) == 0 {
			result.Unchanged++

			continue
		}

		for _, cleared := range ClearedFields(existing, building) {
			result.Cleared++
			result.AddNotice(rowNumber, headerFor(cleared.Field), cleared.Old,
				"This upload clears "+headerFor(cleared.Field))
		}

		result.Updated++
		queue = append(queue, pending{before: existing, after: building})
	}

	if dryRun {
		result.DryRun = true

		return *result
	}

	for _, item := range queue {
		if item.isNew {
			created, err := service.RepositoryBuildingInterface.Create(ctx, tx, item.after)
			helpers.PanicIfError(err)

			helpers.PanicIfError(service.RepositoryBuildingInterface.RecordChanges(ctx, tx,
				[]models.BuildingChange{CreationChange(created, actor, models.BuildingSourceImport, batchId)}))

			continue
		}

		// NOT Update(): that is the form's update and writes only sellable,
		// connectivity and resource_type, so the change log would record edits the
		// database never received.
		_, err := service.RepositoryBuildingInterface.UpdateFromImport(ctx, tx, item.after)
		helpers.PanicIfError(err)

		helpers.PanicIfError(service.RepositoryBuildingInterface.RecordChanges(ctx, tx,
			DiffBuildings(item.before, item.after, actor, models.BuildingSourceImport, batchId)))
	}

	result.Imported = true

	return *result
}

// Export writes the current table in the template's shape, so an export can be edited
// and uploaded back. The key column is always present: an export without it could
// only ever create duplicates on re-upload.
func (service *ServiceBuildingImpl) Export(ctx context.Context) ([]byte, error) {
	tx, err := service.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer helpers.CommitOrRollback(tx)

	list, err := service.RepositoryBuildingInterface.FindAllForExport(ctx, tx)
	if err != nil {
		return nil, err
	}

	rows := make([][]interface{}, len(list))
	for i, b := range list {
		rows[i] = []interface{}{
			b.ExternalBuildingId, b.Name, b.IrisCode,
			// Joined from building_projects, so an export can be edited and
			// re-uploaded without losing the link.
			b.ProjectIdIris,
			blankIfZeroFloat(b.Latitude), blankIfZeroFloat(b.Longitude),
			b.Subdistrict, b.Citytown, b.Province, b.CbdArea,
			b.BuildingType, b.GradeResource, blankIfZero(b.CompletionYear),
			b.BuildingStatus, yesNoCell(b.CompetitorPresence), yesNoCell(b.CompetitorExclusive),
			blankIfZero(b.Audience), blankIfZero(b.Impression),
			b.Sellable, b.Connectivity, b.ResourceType,
		}
	}

	return spreadsheets.BuildExport(BuildingSheetName, TemplateHeaders(), rows)
}

// resolveProject turns a project code into an id, creating an empty project when the
// code is unknown. Returns the id and whether it had to create one.
//
// On a dry run nothing is created: the preview reports what WOULD be created and
// returns 0, so a preview never leaves rows behind.
func (service *ServiceBuildingImpl) resolveProject(ctx context.Context, tx *sql.Tx, code string, cache map[string]int, actor Actor, batchId string, dryRun bool) (int, bool) {
	if id, seen := cache[code]; seen {
		return id, false
	}

	existing, err := service.RepositoryBuildingProject.FindByIris(ctx, tx, code, false)
	if err == nil {
		cache[code] = existing.Id

		return existing.Id, false
	}

	if dryRun {
		return 0, true
	}

	created, err := service.RepositoryBuildingProject.Create(ctx, tx, models.BuildingProject{
		ProjectIdIris: code,
		// Named after the code until the projects file fills it in. A blank name
		// would fail the NOT NULL check, and inventing a plausible name would be
		// worse -- it would look like real data.
		Name: code,
	})
	helpers.PanicIfError(err)

	helpers.PanicIfError(service.RepositoryBuildingProject.RecordChanges(ctx, tx,
		[]models.BuildingProjectChange{{
			ProjectId: created.Id, ProjectIdIris: code,
			ActorUserId: actor.UserId, ActorRole: actor.Role,
			Action:  models.BuildingProjectActionCreated,
			Source:  models.BuildingProjectSourceImport,
			BatchId: batchId, NewValue: code,
		}}))

	cache[code] = created.Id

	return created.Id, true
}

// Template is the empty workbook the operator downloads before filling one in.
func (service *ServiceBuildingImpl) Template() ([]byte, error) {
	return BuildTemplate()
}

// FindChanges is the history of one building, newest first.
func (service *ServiceBuildingImpl) FindChanges(ctx context.Context, buildingId int, take int, skip int) ([]models.BuildingChange, int) {
	tx, err := service.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	list, err := service.RepositoryBuildingInterface.FindChanges(ctx, tx, buildingId, take, skip)
	helpers.PanicIfError(err)

	total, err := service.RepositoryBuildingInterface.CountChanges(ctx, tx, buildingId)
	helpers.PanicIfError(err)

	return list, total
}

// blankIfZero writes an unset count as an empty cell rather than 0. Exporting 0 and
// re-importing it would turn "never filled in" into a real zero on the way back.
func blankIfZero(v int) interface{} {
	if v == 0 {
		return ""
	}

	return v
}

func blankIfZeroFloat(v float64) interface{} {
	if v == 0 {
		return ""
	}

	return v
}

// yesNoCell writes the words the column's vocabulary uses, so an exported file reads
// the same as one a person filled in.
func yesNoCell(v bool) interface{} {
	if v {
		return "yes"
	}

	return "no"
}
