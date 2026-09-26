package buildingproject

import (
	"context"
	"database/sql"
	"strings"

	"github.com/malikabdulaziz/tmn-backend/exceptions"
	"github.com/malikabdulaziz/tmn-backend/helpers"
	"github.com/malikabdulaziz/tmn-backend/models"
	repositoriesBuildingProject "github.com/malikabdulaziz/tmn-backend/repositories/buildingproject"
	webBuildingProject "github.com/malikabdulaziz/tmn-backend/web/buildingproject"
)

type ServiceBuildingProjectImpl struct {
	DB *sql.DB
	repositoriesBuildingProject.RepositoryBuildingProjectInterface
}

func NewServiceBuildingProjectImpl(
	db *sql.DB,
	repo repositoriesBuildingProject.RepositoryBuildingProjectInterface,
) ServiceBuildingProjectInterface {
	return &ServiceBuildingProjectImpl{DB: db, RepositoryBuildingProjectInterface: repo}
}

// canSeeFinance decides whether the landlord money columns are read at all. The
// permission is checked here rather than on the route because it gates COLUMNS, not
// access: everyone may open a project, only finance roles may see what TMN pays for it.
func canSeeFinance(actor Actor) bool {
	return models.RoleCan(actor.Role, models.PermissionBuildingProjectsFinance)
}

func (s *ServiceBuildingProjectImpl) Create(ctx context.Context, request webBuildingProject.SaveBuildingProjectRequest, actor Actor) webBuildingProject.BuildingProjectResponse {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	project := requestToProject(request)
	s.assertIrisFree(ctx, tx, project.ProjectIdIris, 0)

	created, err := s.RepositoryBuildingProjectInterface.Create(ctx, tx, project)
	helpers.PanicIfError(err)

	err = s.RepositoryBuildingProjectInterface.RecordChanges(ctx, tx,
		[]models.BuildingProjectChange{CreationChange(created, actor, models.BuildingProjectSourceForm, "")})
	helpers.PanicIfError(err)

	return s.responseById(ctx, tx, created.Id, actor)
}

func (s *ServiceBuildingProjectImpl) FindAll(ctx context.Context, request webBuildingProject.BuildingProjectRequestFindAll, actor Actor) ([]webBuildingProject.BuildingProjectResponse, int) {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	filter := repositoriesBuildingProject.ListFilter{
		Search:       request.GetSearch(),
		Status:       request.Status,
		ContractType: request.ContractType,
		Pic:          request.Pic,
	}

	finance := canSeeFinance(actor)
	list, err := s.RepositoryBuildingProjectInterface.FindAll(ctx, tx, request.GetTake(), request.GetSkip(),
		request.GetOrderBy(), request.GetOrderDirection(), filter, finance)
	helpers.PanicIfError(err)

	total, err := s.RepositoryBuildingProjectInterface.CountAll(ctx, tx, filter)
	helpers.PanicIfError(err)

	responses := make([]webBuildingProject.BuildingProjectResponse, len(list))
	for i, item := range list {
		responses[i] = projectToResponse(item, finance)
	}

	return responses, total
}

func (s *ServiceBuildingProjectImpl) FindById(ctx context.Context, id int, actor Actor) webBuildingProject.BuildingProjectResponse {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	return s.responseById(ctx, tx, id, actor)
}

func (s *ServiceBuildingProjectImpl) Update(ctx context.Context, request webBuildingProject.SaveBuildingProjectRequest, id int, actor Actor) webBuildingProject.BuildingProjectResponse {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	// Read WITH finance regardless of the caller: the diff must compare every field,
	// or an editor without the permission would appear to blank the rental they
	// cannot see. The values are used for comparison and logging, never returned --
	// projectToResponse re-applies the gate on the way out.
	existing, err := s.RepositoryBuildingProjectInterface.FindById(ctx, tx, id, true)
	if err != nil {
		panic(exceptions.NewNotFoundError("building project not found"))
	}

	updated := requestToProject(request)
	updated.Id = existing.Id

	// A caller without the finance permission cannot send those fields, so their
	// absence from the request is not an instruction to clear them.
	if !canSeeFinance(actor) {
		updated.AnnualRental = existing.AnnualRental
		updated.CompanyName = existing.CompanyName
		updated.ContractNo = existing.ContractNo
	}

	s.assertIrisFree(ctx, tx, updated.ProjectIdIris, id)

	changes := DiffProjects(existing, updated, actor, models.BuildingProjectSourceForm, "")
	if len(changes) == 0 {
		// Nothing changed: no write, no audit row, no updated_at bump. A save that
		// changes nothing should leave no trace, or the history fills with noise.
		return s.responseById(ctx, tx, id, actor)
	}

	_, err = s.RepositoryBuildingProjectInterface.Update(ctx, tx, updated)
	helpers.PanicIfError(err)

	helpers.PanicIfError(s.RepositoryBuildingProjectInterface.RecordChanges(ctx, tx, changes))

	return s.responseById(ctx, tx, id, actor)
}

func (s *ServiceBuildingProjectImpl) Delete(ctx context.Context, id int, actor Actor) {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	existing, err := s.RepositoryBuildingProjectInterface.FindById(ctx, tx, id, true)
	if err != nil {
		panic(exceptions.NewNotFoundError("building project not found"))
	}

	// Logged BEFORE the delete: the foreign key nulls project_id on the change rows,
	// so writing afterwards would file the row against a project that no longer
	// exists. project_id_iris and the name carry the evidence forward.
	err = s.RepositoryBuildingProjectInterface.RecordChanges(ctx, tx,
		[]models.BuildingProjectChange{DeletionChange(existing, actor, models.BuildingProjectSourceForm, "")})
	helpers.PanicIfError(err)

	helpers.PanicIfError(s.RepositoryBuildingProjectInterface.Delete(ctx, tx, id))
}

func (s *ServiceBuildingProjectImpl) FindChanges(ctx context.Context, projectId int, take int, skip int, actor Actor) ([]webBuildingProject.BuildingProjectChangeResponse, int) {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	list, err := s.RepositoryBuildingProjectInterface.FindChanges(ctx, tx, projectId, take, skip)
	helpers.PanicIfError(err)

	total, err := s.RepositoryBuildingProjectInterface.CountChanges(ctx, tx, projectId)
	helpers.PanicIfError(err)

	finance := canSeeFinance(actor)
	responses := make([]webBuildingProject.BuildingProjectChangeResponse, 0, len(list))
	for _, item := range list {
		responses = append(responses, changeToResponse(item, finance))
	}

	return responses, total
}

// assertIrisFree rejects a key already used by another project. The unique index would
// catch it anyway, but as a 500 naming a constraint; this names the field instead.
func (s *ServiceBuildingProjectImpl) assertIrisFree(ctx context.Context, tx *sql.Tx, iris string, selfId int) {
	existing, err := s.RepositoryBuildingProjectInterface.FindByIris(ctx, tx, iris, false)
	if err != nil {
		return // not found: the key is free
	}
	if existing.Id != selfId {
		panic(exceptions.NewBadRequest("project_id_iris " + iris + " already belongs to another project"))
	}
}

func (s *ServiceBuildingProjectImpl) responseById(ctx context.Context, tx *sql.Tx, id int, actor Actor) webBuildingProject.BuildingProjectResponse {
	finance := canSeeFinance(actor)
	project, err := s.RepositoryBuildingProjectInterface.FindById(ctx, tx, id, finance)
	if err != nil {
		panic(exceptions.NewNotFoundError("building project not found"))
	}

	return projectToResponse(project, finance)
}

// requestToProject canonicalises every closed vocabulary on the way in, so what is
// STORED is always the canonical spelling and a filter comparing exact strings works.
// An unknown value is rejected naming the field and what it accepts.
func requestToProject(request webBuildingProject.SaveBuildingProjectRequest) models.BuildingProject {
	return models.BuildingProject{
		ProjectIdIris:    strings.TrimSpace(request.ProjectIdIris),
		Name:             strings.TrimSpace(request.Name),
		BuildingType:     strings.TrimSpace(request.BuildingType),
		Grade:            strings.TrimSpace(request.Grade),
		Pic:              strings.TrimSpace(request.Pic),
		TmnProjectStatus: canonical("tmn_project_status", request.TmnProjectStatus),
		NoOfTower:        request.NoOfTower,
		NoOfScreen:       request.NoOfScreen,
		CreatedDate:      request.CreatedDate,
		Remark:           strings.TrimSpace(request.Remark),
		ContractType:     canonical("contract_type", request.ContractType),
		ContractNo:       strings.TrimSpace(request.ContractNo),
		ContractDate:     request.ContractDate,
		ContractStart:    request.ContractStart,
		ContractEnd:      request.ContractEnd,
		PeriodMonth:      request.PeriodMonth,
		AnnualRental:     request.AnnualRental,
		PaymentTerm:      canonical("payment_term", request.PaymentTerm),
		CompanyName:      strings.TrimSpace(request.CompanyName),
		Exclusivity:      canonical("exclusivity", request.Exclusivity),
		DocType:          canonical("doc_type", request.DocType),
		ContractStatus:   canonical("contract_status", request.ContractStatus),
		CancelledAt:      request.CancelledAt,
		CancelLastStatus: strings.TrimSpace(request.CancelLastStatus),
		CancelReason:     strings.TrimSpace(request.CancelReason),
	}
}

func canonical(field string, value string) string {
	canonicalised, ok := CanonicalVocabularyValue(field, value)
	if !ok {
		panic(exceptions.NewBadRequest(VocabularyError(field, value)))
	}

	return canonicalised
}
