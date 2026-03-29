package subcategory

import (
	"context"

	webSubCategory "github.com/malikabdulaziz/tmn-backend/web/subcategory"
)

type ServiceSubCategoryInterface interface {
	Create(ctx context.Context, request webSubCategory.CreateSubCategoryRequest) webSubCategory.SubCategoryResponse
	FindAll(ctx context.Context, request webSubCategory.SubCategoryRequestFindAll) ([]webSubCategory.SubCategoryResponse, int)
	FindById(ctx context.Context, id int) webSubCategory.SubCategoryResponse
	Update(ctx context.Context, request webSubCategory.UpdateSubCategoryRequest, id int) webSubCategory.SubCategoryResponse
	Delete(ctx context.Context, id int)
	Import(ctx context.Context, fileBytes []byte, fileType string) []webSubCategory.SubCategoryResponse
	Export(ctx context.Context, search string) ([]byte, error)
}
