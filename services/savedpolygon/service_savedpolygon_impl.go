package savedpolygon

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/malikabdulaziz/tmn-backend/exceptions"
	"github.com/malikabdulaziz/tmn-backend/helpers"
	"github.com/malikabdulaziz/tmn-backend/models"
	repositoriesSavedPolygon "github.com/malikabdulaziz/tmn-backend/repositories/savedpolygon"
	webSavedPolygon "github.com/malikabdulaziz/tmn-backend/web/savedpolygon"
)

type ServiceSavedPolygonImpl struct {
	DB                            *sql.DB
	RepositorySavedPolygonInterface repositoriesSavedPolygon.RepositorySavedPolygonInterface
}

func NewServiceSavedPolygonImpl(
	db *sql.DB,
	repositorySavedPolygon repositoriesSavedPolygon.RepositorySavedPolygonInterface,
) ServiceSavedPolygonInterface {
	return &ServiceSavedPolygonImpl{
		DB:                            db,
		RepositorySavedPolygonInterface: repositorySavedPolygon,
	}
}

func (s *ServiceSavedPolygonImpl) validatePoints(points []webSavedPolygon.SavedPolygonPointRequest) {
	if len(points) < 3 {
		panic(exceptions.NewBadRequest("polygon must have at least 3 points"))
	}
	for i, p := range points {
		if p.Lat < -90 || p.Lat > 90 {
			panic(exceptions.NewBadRequest("invalid lat at point " + strconv.Itoa(i+1)))
		}
		if p.Lng < -180 || p.Lng > 180 {
			panic(exceptions.NewBadRequest("invalid lng at point " + strconv.Itoa(i+1)))
		}
	}
}

// Create creates a new saved polygon with its points
func (s *ServiceSavedPolygonImpl) Create(ctx context.Context, request webSavedPolygon.CreateSavedPolygonRequest) webSavedPolygon.SavedPolygonResponse {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	s.validatePoints(request.Points)

	polygon := models.SavedPolygon{Name: request.Name}
	points := make([]models.SavedPolygonPoint, len(request.Points))
	for i, p := range request.Points {
		points[i] = models.SavedPolygonPoint{Ord: i, Lat: p.Lat, Lng: p.Lng}
	}

	created, err := s.RepositorySavedPolygonInterface.Create(ctx, tx, polygon, points)
	helpers.PanicIfError(err)
	return s.modelToResponse(created)
}

// FindAll retrieves all saved polygons with pagination
func (s *ServiceSavedPolygonImpl) FindAll(ctx context.Context, request webSavedPolygon.SavedPolygonRequestFindAll) ([]webSavedPolygon.SavedPolygonResponse, int) {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	list, err := s.RepositorySavedPolygonInterface.FindAll(ctx, tx, request.GetTake(), request.GetSkip(), request.GetOrderBy(), request.GetOrderDirection())
	helpers.PanicIfError(err)
	total, err := s.RepositorySavedPolygonInterface.CountAll(ctx, tx)
	helpers.PanicIfError(err)

	responses := make([]webSavedPolygon.SavedPolygonResponse, len(list))
	for i, p := range list {
		responses[i] = s.modelToResponse(p)
	}
	return responses, total
}

// FindById retrieves a saved polygon by ID
func (s *ServiceSavedPolygonImpl) FindById(ctx context.Context, id int) webSavedPolygon.SavedPolygonResponse {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	polygon, err := s.RepositorySavedPolygonInterface.FindById(ctx, tx, id)
	if err == sql.ErrNoRows {
		panic(exceptions.NewNotFoundError("saved polygon not found"))
	}
	helpers.PanicIfError(err)
	return s.modelToResponse(polygon)
}

// Update updates a saved polygon and replaces its points
func (s *ServiceSavedPolygonImpl) Update(ctx context.Context, request webSavedPolygon.UpdateSavedPolygonRequest, id int) webSavedPolygon.SavedPolygonResponse {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	s.validatePoints(request.Points)

	existing, err := s.RepositorySavedPolygonInterface.FindById(ctx, tx, id)
	if err == sql.ErrNoRows {
		panic(exceptions.NewNotFoundError("saved polygon not found"))
	}
	helpers.PanicIfError(err)

	existing.Name = request.Name
	points := make([]models.SavedPolygonPoint, len(request.Points))
	for i, p := range request.Points {
		points[i] = models.SavedPolygonPoint{Ord: i, Lat: p.Lat, Lng: p.Lng}
	}

	updated, err := s.RepositorySavedPolygonInterface.Update(ctx, tx, existing, points)
	helpers.PanicIfError(err)
	return s.modelToResponse(updated)
}

// Delete deletes a saved polygon
func (s *ServiceSavedPolygonImpl) Delete(ctx context.Context, id int) {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	_, err = s.RepositorySavedPolygonInterface.FindById(ctx, tx, id)
	if err == sql.ErrNoRows {
		panic(exceptions.NewNotFoundError("saved polygon not found"))
	}
	helpers.PanicIfError(err)
	err = s.RepositorySavedPolygonInterface.Delete(ctx, tx, id)
	helpers.PanicIfError(err)
}

func (s *ServiceSavedPolygonImpl) modelToResponse(p models.SavedPolygon) webSavedPolygon.SavedPolygonResponse {
	points := make([]webSavedPolygon.SavedPolygonPointResponse, len(p.Points))
	for i, pt := range p.Points {
		points[i] = webSavedPolygon.SavedPolygonPointResponse{Ord: pt.Ord, Lat: pt.Lat, Lng: pt.Lng}
	}
	return webSavedPolygon.SavedPolygonResponse{
		Id:        p.Id,
		Name:      p.Name,
		Points:    points,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}
