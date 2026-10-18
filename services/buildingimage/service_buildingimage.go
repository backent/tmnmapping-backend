package buildingimage

import (
	"context"
	"database/sql"
	"io"
	"net/http"
	"strings"

	"github.com/malikabdulaziz/tmn-backend/exceptions"
	"github.com/malikabdulaziz/tmn-backend/helpers"
	"github.com/malikabdulaziz/tmn-backend/models"
	repositoriesBuildingImage "github.com/malikabdulaziz/tmn-backend/repositories/buildingimage"
)

type ServiceBuildingImageInterface interface {
	Upload(ctx context.Context, buildingId int, slot string, content []byte, contentType string, userId int) models.HostedBuildingImage
	FindByBuilding(ctx context.Context, buildingId int) []models.HostedBuildingImage
	Delete(ctx context.Context, buildingId int, slot string)

	// Serve writes one photo to w: ours when we host it, ERP's when we do not.
	// Reports whether anything was written, so the caller can 404 rather than send
	// an empty 200.
	Serve(ctx context.Context, w http.ResponseWriter, buildingId int, slot string) bool
}

type ServiceBuildingImageImpl struct {
	DB         *sql.DB
	Repository repositoriesBuildingImage.RepositoryBuildingImageInterface
	Storage    *Storage
	// ERPFallback serves the ERP photo for a slot when this application hosts none.
	// A function rather than a client so the fallback can be tested without ERP.
	ERPFallback func(w http.ResponseWriter, erpPath string) bool
	// BuildingImages returns the ERP paths recorded on the building row.
	BuildingImages func(ctx context.Context, tx *sql.Tx, buildingId int) ([]models.BuildingImage, error)
}

func NewServiceBuildingImageImpl(
	db *sql.DB,
	repository repositoriesBuildingImage.RepositoryBuildingImageInterface,
	storage *Storage,
) *ServiceBuildingImageImpl {
	return &ServiceBuildingImageImpl{DB: db, Repository: repository, Storage: storage}
}

func (s *ServiceBuildingImageImpl) Upload(ctx context.Context, buildingId int, slot string, content []byte, contentType string, userId int) models.HostedBuildingImage {
	if !models.IsValidBuildingImageSlot(slot) {
		panic(exceptions.NewBadRequest("slot must be one of " + strings.Join(models.BuildingImageSlots, ", ")))
	}
	if !s.Storage.Configured() {
		// Said plainly rather than failing mid-write: without a volume the upload
		// would appear to work and vanish at the next deploy.
		panic(exceptions.NewBadRequest(ErrNoStorageDir.Error()))
	}
	if len(content) == 0 {
		panic(exceptions.NewBadRequest("the uploaded file is empty"))
	}
	if len(content) > MaxUploadBytes {
		panic(exceptions.NewBadRequest("the image is larger than 8 MB"))
	}

	// Sniffed, not trusted: the browser's Content-Type is whatever the client says.
	extension, ok := ExtensionFor(http.DetectContentType(content))
	if !ok {
		panic(exceptions.NewBadRequest("the file must be one of " + strings.Join(AllowedContentTypes(), ", ")))
	}

	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	path, err := s.Storage.Save(buildingId, slot, extension, content)
	helpers.PanicIfError(err)

	previous, err := s.Repository.Upsert(ctx, tx, models.HostedBuildingImage{
		BuildingId: buildingId, Slot: slot, Path: path,
		ContentType: http.DetectContentType(content),
		SizeBytes:   int64(len(content)), UploadedByUserId: userId,
	})
	if err != nil {
		// The row did not land, so the file we just wrote is orphaned. Remove it
		// rather than leave the volume filling with files nothing points at.
		_ = s.Storage.Remove(path)
		helpers.PanicIfError(err)
	}

	// Only once the row is safely replaced. Deleting first would lose the old photo
	// if the write failed.
	if previous != "" {
		_ = s.Storage.Remove(previous)
	}

	stored, err := s.Repository.FindSlot(ctx, tx, buildingId, slot)
	helpers.PanicIfError(err)

	return stored
}

func (s *ServiceBuildingImageImpl) FindByBuilding(ctx context.Context, buildingId int) []models.HostedBuildingImage {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	images, err := s.Repository.FindByBuilding(ctx, tx, buildingId)
	helpers.PanicIfError(err)

	return images
}

// Delete removes the photo this application hosts. The ERP one, if any, becomes
// visible again -- deleting ours is "stop overriding", not "remove the picture".
func (s *ServiceBuildingImageImpl) Delete(ctx context.Context, buildingId int, slot string) {
	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	path, err := s.Repository.Delete(ctx, tx, buildingId, slot)
	helpers.PanicIfError(err)

	if path != "" {
		_ = s.Storage.Remove(path)
	}
}

// Serve prefers the photo this application hosts and falls back to ERP's.
//
// One route for both, so the client asks for a building's front photo without
// knowing or caring which side of the cutover it comes from.
func (s *ServiceBuildingImageImpl) Serve(ctx context.Context, w http.ResponseWriter, buildingId int, slot string) bool {
	if !models.IsValidBuildingImageSlot(slot) {
		return false
	}

	tx, err := s.DB.Begin()
	helpers.PanicIfError(err)
	defer helpers.CommitOrRollback(tx)

	if s.Storage.Configured() {
		if hosted, err := s.Repository.FindSlot(ctx, tx, buildingId, slot); err == nil {
			if file, err := s.Storage.Open(hosted.Path); err == nil {
				defer file.Close()
				w.Header().Set("Content-Type", hosted.ContentType)
				// Short, because replacing a photo changes its stored filename but
				// not this URL.
				w.Header().Set("Cache-Control", "private, max-age=60")
				_, _ = io.Copy(w, file)

				return true
			}
			// The row says there is a file and the volume disagrees -- most likely a
			// deploy that lost an unmounted directory. Fall through to ERP rather
			// than show nothing.
		}
	}

	if s.BuildingImages == nil || s.ERPFallback == nil {
		return false
	}

	erpImages, err := s.BuildingImages(ctx, tx, buildingId)
	if err != nil {
		return false
	}

	for _, image := range erpImages {
		if image.Name == slot && image.Path != "" {
			return s.ERPFallback(w, image.Path)
		}
	}

	return false
}
