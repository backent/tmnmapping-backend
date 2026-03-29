package branch

import (
	"context"

	webBranch "github.com/malikabdulaziz/tmn-backend/web/branch"
)

type ServiceBranchInterface interface {
	Create(ctx context.Context, request webBranch.CreateBranchRequest) webBranch.BranchResponse
	FindAll(ctx context.Context, request webBranch.BranchRequestFindAll) ([]webBranch.BranchResponse, int)
	FindById(ctx context.Context, id int) webBranch.BranchResponse
	Update(ctx context.Context, request webBranch.UpdateBranchRequest, id int) webBranch.BranchResponse
	Delete(ctx context.Context, id int)
	Import(ctx context.Context, fileBytes []byte, fileType string) []webBranch.BranchResponse
	Export(ctx context.Context, search string) ([]byte, error)
}
