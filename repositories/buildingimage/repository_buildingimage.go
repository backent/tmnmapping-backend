package buildingimage

import (
	"context"
	"database/sql"

	"github.com/malikabdulaziz/tmn-backend/models"
)

type RepositoryBuildingImageInterface interface {
	// Upsert replaces the photo in a slot. One row per building and slot, so
	// re-uploading overwrites rather than accumulating versions nobody asked for.
	// Returns the path that was there before, so its file can be deleted.
	Upsert(ctx context.Context, tx *sql.Tx, image models.HostedBuildingImage) (previousPath string, err error)
	FindByBuilding(ctx context.Context, tx *sql.Tx, buildingId int) ([]models.HostedBuildingImage, error)
	FindSlot(ctx context.Context, tx *sql.Tx, buildingId int, slot string) (models.HostedBuildingImage, error)
	Delete(ctx context.Context, tx *sql.Tx, buildingId int, slot string) (deletedPath string, err error)
}

type RepositoryBuildingImageImpl struct{}

func NewRepositoryBuildingImageImpl() RepositoryBuildingImageInterface {
	return &RepositoryBuildingImageImpl{}
}

const imageColumns = `i.id, i.building_id, i.slot, i.path, i.content_type, i.size_bytes,
	i.uploaded_by_user_id, u.name, i.created_at, i.updated_at`

const imageFrom = ` FROM ` + `building_images i LEFT JOIN users u ON u.id = i.uploaded_by_user_id`

func scanImage(scanner interface{ Scan(...interface{}) error }) (models.HostedBuildingImage, error) {
	var (
		image      models.HostedBuildingImage
		uploadedBy sql.NullInt64
		uploader   sql.NullString
	)

	err := scanner.Scan(&image.Id, &image.BuildingId, &image.Slot, &image.Path,
		&image.ContentType, &image.SizeBytes, &uploadedBy, &uploader,
		&image.CreatedAt, &image.UpdatedAt)
	if err != nil {
		return models.HostedBuildingImage{}, err
	}

	image.UploadedByUserId = int(uploadedBy.Int64)
	image.UploadedByName = uploader.String

	return image, nil
}

// Upsert returns the path it replaced so the caller can delete that file. Doing it
// here rather than in the service keeps the read and the write in one statement, so
// two uploads racing cannot both think they replaced nothing and leak a file.
func (r *RepositoryBuildingImageImpl) Upsert(ctx context.Context, tx *sql.Tx, image models.HostedBuildingImage) (string, error) {
	var previous sql.NullString

	err := tx.QueryRowContext(ctx, `
		INSERT INTO `+models.BuildingImageTable+`
			(building_id, slot, path, content_type, size_bytes, uploaded_by_user_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (building_id, slot) DO UPDATE SET
			path = EXCLUDED.path,
			content_type = EXCLUDED.content_type,
			size_bytes = EXCLUDED.size_bytes,
			uploaded_by_user_id = EXCLUDED.uploaded_by_user_id,
			updated_at = CURRENT_TIMESTAMP
		RETURNING (SELECT path FROM `+models.BuildingImageTable+`
		           WHERE building_id = $1 AND slot = $2)`,
		image.BuildingId, image.Slot, image.Path, image.ContentType,
		image.SizeBytes, nullIfZero(image.UploadedByUserId),
	).Scan(&previous)
	if err != nil {
		return "", err
	}

	if previous.String == image.Path {
		return "", nil
	}

	return previous.String, nil
}

func (r *RepositoryBuildingImageImpl) FindByBuilding(ctx context.Context, tx *sql.Tx, buildingId int) ([]models.HostedBuildingImage, error) {
	rows, err := tx.QueryContext(ctx,
		`SELECT `+imageColumns+imageFrom+` WHERE i.building_id = $1 ORDER BY i.slot`, buildingId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	images := []models.HostedBuildingImage{}
	for rows.Next() {
		image, err := scanImage(rows)
		if err != nil {
			return nil, err
		}
		images = append(images, image)
	}

	return images, rows.Err()
}

func (r *RepositoryBuildingImageImpl) FindSlot(ctx context.Context, tx *sql.Tx, buildingId int, slot string) (models.HostedBuildingImage, error) {
	return scanImage(tx.QueryRowContext(ctx,
		`SELECT `+imageColumns+imageFrom+` WHERE i.building_id = $1 AND i.slot = $2`,
		buildingId, slot))
}

func (r *RepositoryBuildingImageImpl) Delete(ctx context.Context, tx *sql.Tx, buildingId int, slot string) (string, error) {
	var path sql.NullString

	err := tx.QueryRowContext(ctx,
		`DELETE FROM `+models.BuildingImageTable+`
		 WHERE building_id = $1 AND slot = $2 RETURNING path`, buildingId, slot).Scan(&path)
	if err == sql.ErrNoRows {
		return "", nil
	}

	return path.String, err
}

func nullIfZero(v int) interface{} {
	if v == 0 {
		return nil
	}

	return v
}
