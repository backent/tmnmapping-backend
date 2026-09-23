package building

import (
	"context"
	"database/sql"
	"strings"

	"github.com/malikabdulaziz/tmn-backend/exceptions"
	"github.com/malikabdulaziz/tmn-backend/helpers"
	"github.com/malikabdulaziz/tmn-backend/models"
	webBuilding "github.com/malikabdulaziz/tmn-backend/web/building"
)

// Create raises a building from the form.
//
// Buildings used to arrive only from ERP, so there was no create path at all -- the
// form could edit three columns and nothing else. The spreadsheet owns them now, and
// a person adding one building should not have to prepare a file to do it.
func (service *ServiceBuildingImpl) Create(ctx context.Context, request webBuilding.SaveBuildingRequest, actor Actor) webBuilding.BuildingResponse {
	tx, err := service.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	building := requestToBuilding(request)

	service.assertCodeFree(ctx, tx, building.ExternalBuildingId, 0)
	service.assertIrisFree(ctx, tx, building.IrisCode, 0)
	service.linkProject(ctx, tx, &building, request.ProjectIdIris, actor)

	created, err := service.RepositoryBuildingInterface.Create(ctx, tx, building)
	helpers.PanicIfError(err)

	helpers.PanicIfError(service.RepositoryBuildingInterface.RecordChanges(ctx, tx,
		[]models.BuildingChange{CreationChange(created, actor, models.BuildingSourceForm, "")}))

	return service.responseById(ctx, tx, created.Id)
}

// Save replaces a building from the form.
//
// This is not the old Update, which wrote sellable, connectivity and resource_type
// and left everything else alone -- that was right when ERP owned the row. The form
// now REPLACES the record, so a blank clears, exactly as a blank cell does on the
// import. Two ways of editing the same building must not disagree about what an empty
// field means.
//
// Photos are carried across: the ERP sync owns them and the form does not offer them,
// so a save must not blank what it never showed.
func (service *ServiceBuildingImpl) Save(ctx context.Context, request webBuilding.SaveBuildingRequest, id int, actor Actor) webBuilding.BuildingResponse {
	tx, err := service.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	existing, err := service.RepositoryBuildingInterface.FindById(ctx, tx, id)
	if err == sql.ErrNoRows {
		panic(exceptions.NewNotFoundError("building not found"))
	}
	helpers.PanicIfError(err)

	updated := requestToBuilding(request)
	updated.Id = existing.Id
	updated.Images = existing.Images

	service.assertCodeFree(ctx, tx, updated.ExternalBuildingId, id)
	service.assertIrisFree(ctx, tx, updated.IrisCode, id)
	service.linkProject(ctx, tx, &updated, request.ProjectIdIris, actor)

	changes := DiffBuildings(existing, updated, actor, models.BuildingSourceForm, "")
	if len(changes) == 0 {
		// A save that changes nothing leaves no trace: no write, no audit row, no
		// updated_at. Otherwise the history fills with edits nobody made.
		return service.responseById(ctx, tx, id)
	}

	_, err = service.RepositoryBuildingInterface.UpdateFromImport(ctx, tx, updated)
	helpers.PanicIfError(err)

	helpers.PanicIfError(service.RepositoryBuildingInterface.RecordChanges(ctx, tx, changes))

	return service.responseById(ctx, tx, id)
}

// linkProject resolves the project CODE the form sends. An unknown code raises an
// empty project rather than refusing the save, matching the import, so the two files
// can arrive in either order and a person is not blocked by a project that has not
// been set up yet. A blank code clears the link.
func (service *ServiceBuildingImpl) linkProject(ctx context.Context, tx *sql.Tx, building *models.Building, code string, actor Actor) {
	code = strings.TrimSpace(code)
	if code == "" {
		building.ProjectId = 0
		building.ProjectIdIris = ""

		return
	}

	projectId, _ := service.resolveProject(ctx, tx, code, map[string]int{},
		actor, models.BuildingProjectSourceForm, "", false)

	building.ProjectId = projectId
	building.ProjectIdIris = code
}

// assertCodeFree rejects a Building Code already used by another building. The unique
// index would catch it too, but as a 500 naming a constraint; this names the field.
func (service *ServiceBuildingImpl) assertCodeFree(ctx context.Context, tx *sql.Tx, code string, selfId int) {
	if strings.TrimSpace(code) == "" {
		return
	}

	existing, err := service.RepositoryBuildingInterface.FindByExternalId(ctx, tx, code)
	if err == nil && existing.Id != selfId {
		panic(exceptions.NewBadRequest("Building Code " + code + " already belongs to another building"))
	}
}

// assertIrisFree does the same for the IRIS Code, which must be unique when present
// because the price import looks buildings up by it.
func (service *ServiceBuildingImpl) assertIrisFree(ctx context.Context, tx *sql.Tx, iris string, selfId int) {
	if strings.TrimSpace(iris) == "" {
		return
	}

	existing, err := service.RepositoryBuildingInterface.FindByIrisCode(ctx, tx, iris)
	if err == nil && existing.Id != selfId {
		panic(exceptions.NewBadRequest("IRIS Code " + iris + " already belongs to another building"))
	}
}

func (service *ServiceBuildingImpl) responseById(ctx context.Context, tx *sql.Tx, id int) webBuilding.BuildingResponse {
	building, err := service.RepositoryBuildingInterface.FindById(ctx, tx, id)
	if err != nil {
		panic(exceptions.NewNotFoundError("building not found"))
	}

	return webBuilding.BuildingModelToBuildingResponse(building)
}

// requestToBuilding maps the form payload, deriving what must not be typed.
func requestToBuilding(request webBuilding.SaveBuildingRequest) models.Building {
	building := models.Building{
		ExternalBuildingId: strings.TrimSpace(request.ExternalBuildingId),
		Name:               strings.TrimSpace(request.Name),
		IrisCode:           strings.TrimSpace(request.IrisCode),
		Latitude:           request.Latitude,
		Longitude:          request.Longitude,
		Subdistrict:        strings.TrimSpace(request.Subdistrict),
		Citytown:           strings.TrimSpace(request.Citytown),
		Province:           strings.TrimSpace(request.Province),
		CbdArea:            strings.TrimSpace(request.CbdArea),
		// Canonicalised as the import does, and a blank stays blank rather than
		// becoming "Other".
		BuildingType:        canonicalTypeOrBlank(request.BuildingType),
		GradeResource:       strings.TrimSpace(request.GradeResource),
		CompletionYear:      request.CompletionYear,
		CompetitorPresence:  request.CompetitorPresence,
		CompetitorExclusive: request.CompetitorExclusive,
		Audience:            request.Audience,
		Impression:          request.Impression,
		Sellable:            strings.TrimSpace(request.Sellable),
		Connectivity:        strings.TrimSpace(request.Connectivity),
		ResourceType:        strings.TrimSpace(request.ResourceType),
	}

	// An unfamiliar status is accepted and stored as written, as on the import: ERP
	// adds statuses without warning and only the map's progress filter is affected.
	building.BuildingStatus, _ = CanonicalBuildingStatus(request.BuildingStatus)

	// competitor_location has never been anything but a duplicate of presence, and
	// the map filters on it. Kept in step rather than left to diverge.
	building.CompetitorLocation = building.CompetitorPresence

	// Derived, never typed: one source of truth, so the form cannot disagree with the
	// status and competitor fields it is calculated from.
	building.LcdPresenceStatus = calculateLcdPresenceStatus(
		building.CompetitorPresence, building.CompetitorExclusive, building.BuildingStatus)

	return building
}
