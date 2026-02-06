package buildingrestriction

import (
	"context"
	"database/sql"

	"github.com/malikabdulaziz/tmn-backend/exceptions"
	"github.com/malikabdulaziz/tmn-backend/helpers"
	"github.com/malikabdulaziz/tmn-backend/models"
	repositoriesBuilding "github.com/malikabdulaziz/tmn-backend/repositories/building"
	repositoriesBuildingRestriction "github.com/malikabdulaziz/tmn-backend/repositories/buildingrestriction"
	webBuildingRestriction "github.com/malikabdulaziz/tmn-backend/web/buildingrestriction"
)

type ServiceBuildingRestrictionImpl struct {
	DB                                    *sql.DB
	RepositoryBuildingRestrictionInterface repositoriesBuildingRestriction.RepositoryBuildingRestrictionInterface
	RepositoryBuildingInterface           repositoriesBuilding.RepositoryBuildingInterface
}

func NewServiceBuildingRestrictionImpl(
	db *sql.DB,
	repoBuildingRestriction repositoriesBuildingRestriction.RepositoryBuildingRestrictionInterface,
	repoBuilding repositoriesBuilding.RepositoryBuildingInterface,
) ServiceBuildingRestrictionInterface {
	return &ServiceBuildingRestrictionImpl{
		DB:                                    db,
		RepositoryBuildingRestrictionInterface: repoBuildingRestriction,
		RepositoryBuildingInterface:           repoBuilding,
	}
}

// validateBuildingIdsErr ensures all building ids exist; panics with BadRequest if any invalid
func (s *ServiceBuildingRestrictionImpl) validateBuildingIdsErr(ctx context.Context, tx *sql.Tx, buildingIds []int) {
	for _, bid := range buildingIds {
		_, err := s.RepositoryBuildingInterface.FindById(ctx, tx, bid)
		if err == sql.ErrNoRows {
			panic(exceptions.NewBadRequest("building not found"))
		}
		helpers.PanicIfError(err)
	}
}

// Create creates a new building restriction with building links
func (s *ServiceBuildingRestrictionImpl) Create(ctx context.Context, request webBuildingRestriction.CreateBuildingRestrictionRequest) webBuildingRestriction.BuildingRestrictionResponse {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	s.validateBuildingIdsErr(ctx, tx, request.BuildingIds)

	restriction := models.BuildingRestriction{Name: request.Name}
	created, err := s.RepositoryBuildingRestrictionInterface.Create(ctx, tx, restriction, request.BuildingIds)
	helpers.PanicIfError(err)
	return s.modelToResponse(created)
}

// FindAll retrieves all building restrictions with pagination
func (s *ServiceBuildingRestrictionImpl) FindAll(ctx context.Context, request webBuildingRestriction.BuildingRestrictionRequestFindAll) ([]webBuildingRestriction.BuildingRestrictionResponse, int) {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	list, err := s.RepositoryBuildingRestrictionInterface.FindAll(ctx, tx, request.GetTake(), request.GetSkip(), request.GetOrderBy(), request.GetOrderDirection())
	helpers.PanicIfError(err)
	total, err := s.RepositoryBuildingRestrictionInterface.CountAll(ctx, tx)
	helpers.PanicIfError(err)

	responses := make([]webBuildingRestriction.BuildingRestrictionResponse, len(list))
	for i, r := range list {
		responses[i] = s.modelToResponse(r)
	}
	return responses, total
}

// FindById retrieves a building restriction by ID
func (s *ServiceBuildingRestrictionImpl) FindById(ctx context.Context, id int) webBuildingRestriction.BuildingRestrictionResponse {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	restriction, err := s.RepositoryBuildingRestrictionInterface.FindById(ctx, tx, id)
	if err == sql.ErrNoRows {
		panic(exceptions.NewNotFoundError("building restriction not found"))
	}
	helpers.PanicIfError(err)
	return s.modelToResponse(restriction)
}

// Update updates a building restriction and replaces building links
func (s *ServiceBuildingRestrictionImpl) Update(ctx context.Context, request webBuildingRestriction.UpdateBuildingRestrictionRequest, id int) webBuildingRestriction.BuildingRestrictionResponse {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	existing, err := s.RepositoryBuildingRestrictionInterface.FindById(ctx, tx, id)
	if err == sql.ErrNoRows {
		panic(exceptions.NewNotFoundError("building restriction not found"))
	}
	helpers.PanicIfError(err)

	s.validateBuildingIdsErr(ctx, tx, request.BuildingIds)

	existing.Name = request.Name
	updated, err := s.RepositoryBuildingRestrictionInterface.Update(ctx, tx, existing, request.BuildingIds)
	helpers.PanicIfError(err)
	return s.modelToResponse(updated)
}

// Delete deletes a building restriction
func (s *ServiceBuildingRestrictionImpl) Delete(ctx context.Context, id int) {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	_, err = s.RepositoryBuildingRestrictionInterface.FindById(ctx, tx, id)
	if err == sql.ErrNoRows {
		panic(exceptions.NewNotFoundError("building restriction not found"))
	}
	helpers.PanicIfError(err)
	err = s.RepositoryBuildingRestrictionInterface.Delete(ctx, tx, id)
	helpers.PanicIfError(err)
}

func (s *ServiceBuildingRestrictionImpl) modelToResponse(r models.BuildingRestriction) webBuildingRestriction.BuildingRestrictionResponse {
	buildings := make([]webBuildingRestriction.BuildingRefResponse, len(r.Buildings))
	for i, b := range r.Buildings {
		buildings[i] = webBuildingRestriction.BuildingRefResponse{Id: b.Id, Name: b.Name}
	}
	return webBuildingRestriction.BuildingRestrictionResponse{
		Id:        r.Id,
		Name:      r.Name,
		Buildings: buildings,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
}
