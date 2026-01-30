package salespackage

import (
	"context"
	"database/sql"

	"github.com/malikabdulaziz/tmn-backend/exceptions"
	"github.com/malikabdulaziz/tmn-backend/helpers"
	"github.com/malikabdulaziz/tmn-backend/models"
	repositoriesBuilding "github.com/malikabdulaziz/tmn-backend/repositories/building"
	repositoriesSalesPackage "github.com/malikabdulaziz/tmn-backend/repositories/salespackage"
	webSalesPackage "github.com/malikabdulaziz/tmn-backend/web/salespackage"
)

type ServiceSalesPackageImpl struct {
	DB                            *sql.DB
	RepositorySalesPackageInterface repositoriesSalesPackage.RepositorySalesPackageInterface
	RepositoryBuildingInterface     repositoriesBuilding.RepositoryBuildingInterface
}

func NewServiceSalesPackageImpl(
	db *sql.DB,
	repoSalesPackage repositoriesSalesPackage.RepositorySalesPackageInterface,
	repoBuilding repositoriesBuilding.RepositoryBuildingInterface,
) ServiceSalesPackageInterface {
	return &ServiceSalesPackageImpl{
		DB:                            db,
		RepositorySalesPackageInterface: repoSalesPackage,
		RepositoryBuildingInterface:     repoBuilding,
	}
}

// validateBuildingIdsErr ensures all building ids exist; panics with BadRequest if any invalid
func (s *ServiceSalesPackageImpl) validateBuildingIdsErr(ctx context.Context, tx *sql.Tx, buildingIds []int) {
	for _, bid := range buildingIds {
		_, err := s.RepositoryBuildingInterface.FindById(ctx, tx, bid)
		if err == sql.ErrNoRows {
			panic(exceptions.NewBadRequest("building not found"))
		}
		helpers.PanicIfError(err)
	}
}

// Create creates a new sales package with building links
func (s *ServiceSalesPackageImpl) Create(ctx context.Context, request webSalesPackage.CreateSalesPackageRequest) webSalesPackage.SalesPackageResponse {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	s.validateBuildingIdsErr(ctx, tx, request.BuildingIds)

	pkg := models.SalesPackage{Name: request.Name}
	created, err := s.RepositorySalesPackageInterface.Create(ctx, tx, pkg, request.BuildingIds)
	helpers.PanicIfError(err)
	return s.modelToResponse(created)
}

// FindAll retrieves all sales packages with pagination
func (s *ServiceSalesPackageImpl) FindAll(ctx context.Context, request webSalesPackage.SalesPackageRequestFindAll) ([]webSalesPackage.SalesPackageResponse, int) {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	list, err := s.RepositorySalesPackageInterface.FindAll(ctx, tx, request.GetTake(), request.GetSkip(), request.GetOrderBy(), request.GetOrderDirection())
	helpers.PanicIfError(err)
	total, err := s.RepositorySalesPackageInterface.CountAll(ctx, tx)
	helpers.PanicIfError(err)

	responses := make([]webSalesPackage.SalesPackageResponse, len(list))
	for i, p := range list {
		responses[i] = s.modelToResponse(p)
	}
	return responses, total
}

// FindById retrieves a sales package by ID
func (s *ServiceSalesPackageImpl) FindById(ctx context.Context, id int) webSalesPackage.SalesPackageResponse {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	pkg, err := s.RepositorySalesPackageInterface.FindById(ctx, tx, id)
	if err == sql.ErrNoRows {
		panic(exceptions.NewNotFoundError("sales package not found"))
	}
	helpers.PanicIfError(err)
	return s.modelToResponse(pkg)
}

// Update updates a sales package and replaces building links
func (s *ServiceSalesPackageImpl) Update(ctx context.Context, request webSalesPackage.UpdateSalesPackageRequest, id int) webSalesPackage.SalesPackageResponse {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	existing, err := s.RepositorySalesPackageInterface.FindById(ctx, tx, id)
	if err == sql.ErrNoRows {
		panic(exceptions.NewNotFoundError("sales package not found"))
	}
	helpers.PanicIfError(err)

	s.validateBuildingIdsErr(ctx, tx, request.BuildingIds)

	existing.Name = request.Name
	updated, err := s.RepositorySalesPackageInterface.Update(ctx, tx, existing, request.BuildingIds)
	helpers.PanicIfError(err)
	return s.modelToResponse(updated)
}

// Delete deletes a sales package
func (s *ServiceSalesPackageImpl) Delete(ctx context.Context, id int) {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	_, err = s.RepositorySalesPackageInterface.FindById(ctx, tx, id)
	if err == sql.ErrNoRows {
		panic(exceptions.NewNotFoundError("sales package not found"))
	}
	helpers.PanicIfError(err)
	err = s.RepositorySalesPackageInterface.Delete(ctx, tx, id)
	helpers.PanicIfError(err)
}

func (s *ServiceSalesPackageImpl) modelToResponse(p models.SalesPackage) webSalesPackage.SalesPackageResponse {
	buildings := make([]webSalesPackage.BuildingRefResponse, len(p.Buildings))
	for i, b := range p.Buildings {
		buildings[i] = webSalesPackage.BuildingRefResponse{Id: b.Id, Name: b.Name}
	}
	return webSalesPackage.SalesPackageResponse{
		Id:        p.Id,
		Name:      p.Name,
		Buildings: buildings,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}
