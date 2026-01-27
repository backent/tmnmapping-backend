package poi

import (
	"context"
	"database/sql"

	"github.com/malikabdulaziz/tmn-backend/exceptions"
	"github.com/malikabdulaziz/tmn-backend/helpers"
	"github.com/malikabdulaziz/tmn-backend/models"
	repositoriesPOI "github.com/malikabdulaziz/tmn-backend/repositories/poi"
	webPOI "github.com/malikabdulaziz/tmn-backend/web/poi"
)

type ServicePOIImpl struct {
	DB                    *sql.DB
	RepositoryPOIInterface repositoriesPOI.RepositoryPOIInterface
}

func NewServicePOIImpl(
	db *sql.DB,
	repositoryPOI repositoriesPOI.RepositoryPOIInterface,
) ServicePOIInterface {
	return &ServicePOIImpl{
		DB:                    db,
		RepositoryPOIInterface: repositoryPOI,
	}
}

// Create creates a new POI with its points
func (service *ServicePOIImpl) Create(ctx context.Context, request webPOI.CreatePOIRequest) webPOI.POIResponse {
	tx, err := service.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	// Convert request to model
	poi := models.POI{
		Name:  request.Name,
		Color: request.Color,
		Points: make([]models.POIPoint, len(request.Points)),
	}

	// Convert points
	for i, pointReq := range request.Points {
		poi.Points[i] = models.POIPoint{
			PlaceName: pointReq.PlaceName,
			Address:   pointReq.Address,
			Latitude:  pointReq.Latitude,
			Longitude: pointReq.Longitude,
		}
	}

	// Create POI (which will create points)
	createdPOI, err := service.RepositoryPOIInterface.Create(ctx, tx, poi)
	helpers.PanicIfError(err)

	return service.poiModelToResponse(createdPOI)
}

// FindAll retrieves all POIs with their points, with pagination
func (service *ServicePOIImpl) FindAll(ctx context.Context, request webPOI.POIRequestFindAll) ([]webPOI.POIResponse, int) {
	tx, err := service.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	pois, err := service.RepositoryPOIInterface.FindAll(ctx, tx, request.GetTake(), request.GetSkip(), request.GetOrderBy(), request.GetOrderDirection())
	helpers.PanicIfError(err)

	total, err := service.RepositoryPOIInterface.CountAll(ctx, tx)
	helpers.PanicIfError(err)

	responses := make([]webPOI.POIResponse, len(pois))
	for i, poi := range pois {
		responses[i] = service.poiModelToResponse(poi)
	}

	return responses, total
}

// FindById retrieves a POI by ID with its points
func (service *ServicePOIImpl) FindById(ctx context.Context, id int) webPOI.POIResponse {
	tx, err := service.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	poi, err := service.RepositoryPOIInterface.FindById(ctx, tx, id)
	if err == sql.ErrNoRows {
		panic(exceptions.NewNotFoundError("POI not found"))
	}
	helpers.PanicIfError(err)

	return service.poiModelToResponse(poi)
}

// Update updates a POI and replaces all its points
func (service *ServicePOIImpl) Update(ctx context.Context, request webPOI.UpdatePOIRequest, id int) webPOI.POIResponse {
	tx, err := service.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	// Verify POI exists
	existingPOI, err := service.RepositoryPOIInterface.FindById(ctx, tx, id)
	if err == sql.ErrNoRows {
		panic(exceptions.NewNotFoundError("POI not found"))
	}
	helpers.PanicIfError(err)

	// Update POI fields
	existingPOI.Name = request.Name
	existingPOI.Color = request.Color

	// Convert points
	existingPOI.Points = make([]models.POIPoint, len(request.Points))
	for i, pointReq := range request.Points {
		existingPOI.Points[i] = models.POIPoint{
			PlaceName: pointReq.PlaceName,
			Address:   pointReq.Address,
			Latitude:  pointReq.Latitude,
			Longitude: pointReq.Longitude,
		}
	}

	// Update POI (which will replace points)
	updatedPOI, err := service.RepositoryPOIInterface.Update(ctx, tx, existingPOI)
	helpers.PanicIfError(err)

	return service.poiModelToResponse(updatedPOI)
}

// Delete deletes a POI (cascade will delete points)
func (service *ServicePOIImpl) Delete(ctx context.Context, id int) {
	tx, err := service.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	// Verify POI exists
	_, err = service.RepositoryPOIInterface.FindById(ctx, tx, id)
	if err == sql.ErrNoRows {
		panic(exceptions.NewNotFoundError("POI not found"))
	}
	helpers.PanicIfError(err)

	// Delete POI
	err = service.RepositoryPOIInterface.Delete(ctx, tx, id)
	helpers.PanicIfError(err)
}

// Helper function to convert model to response
func (service *ServicePOIImpl) poiModelToResponse(poi models.POI) webPOI.POIResponse {
	points := make([]webPOI.POIPointResponse, len(poi.Points))
	for i, point := range poi.Points {
		points[i] = webPOI.POIPointResponse{
			Id:        point.Id,
			PlaceName: point.PlaceName,
			Address:   point.Address,
			Latitude:  point.Latitude,
			Longitude: point.Longitude,
			CreatedAt: point.CreatedAt,
		}
	}

	return webPOI.POIResponse{
		Id:        poi.Id,
		Name:      poi.Name,
		Color:     poi.Color,
		Points:    points,
		CreatedAt: poi.CreatedAt,
		UpdatedAt: poi.UpdatedAt,
	}
}
