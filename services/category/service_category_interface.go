package category

import (
	"context"

	webCategory "github.com/malikabdulaziz/tmn-backend/web/category"
)

type ServiceCategoryInterface interface {
	Create(ctx context.Context, request webCategory.CreateCategoryRequest) webCategory.CategoryResponse
	FindAll(ctx context.Context, request webCategory.CategoryRequestFindAll) ([]webCategory.CategoryResponse, int)
	FindById(ctx context.Context, id int) webCategory.CategoryResponse
	Update(ctx context.Context, request webCategory.UpdateCategoryRequest, id int) webCategory.CategoryResponse
	Delete(ctx context.Context, id int)
	Import(ctx context.Context, fileBytes []byte, fileType string) []webCategory.CategoryResponse
	Export(ctx context.Context, search string) ([]byte, error)
}
