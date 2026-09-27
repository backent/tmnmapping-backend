package injector

import (
	"context"
	"database/sql"

	"github.com/malikabdulaziz/tmn-backend/models"
	repositoriesBuilding "github.com/malikabdulaziz/tmn-backend/repositories/building"
	repositoriesBuildingImage "github.com/malikabdulaziz/tmn-backend/repositories/buildingimage"
	servicesBuildingImage "github.com/malikabdulaziz/tmn-backend/services/buildingimage"
)

// provideBuildingImageService assembles the photo service.
//
// Hand-written rather than left to Wire because the service needs two things Wire
// cannot supply on its own: the storage root, which comes from the environment, and
// the ERP fallback, which is a function so it can be replaced in a test.
func provideBuildingImageService(
	db *sql.DB,
	repository repositoriesBuildingImage.RepositoryBuildingImageInterface,
	buildings repositoriesBuilding.RepositoryBuildingInterface,
) servicesBuildingImage.ServiceBuildingImageInterface {
	service := servicesBuildingImage.NewServiceBuildingImageImpl(
		db, repository, servicesBuildingImage.NewStorageFromEnv())

	service.ERPFallback = servicesBuildingImage.FetchERPImage

	// The ERP paths live on the building row, written by the photo sync.
	service.BuildingImages = func(ctx context.Context, tx *sql.Tx, buildingId int) ([]models.BuildingImage, error) {
		building, err := buildings.FindById(ctx, tx, buildingId)
		if err != nil {
			return nil, err
		}

		return building.Images, nil
	}

	return service
}
