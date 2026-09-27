package buildingproject

import (
	"context"

	"github.com/malikabdulaziz/tmn-backend/web"
	webBuildingProject "github.com/malikabdulaziz/tmn-backend/web/buildingproject"
)

type ServiceBuildingProjectInterface interface {
	// Every method takes the Actor: not for scoping -- any authenticated role may
	// read a project -- but because the landlord money is gated per field, and
	// because nothing may be written without recording who wrote it.
	Create(ctx context.Context, request webBuildingProject.SaveBuildingProjectRequest, actor Actor) webBuildingProject.BuildingProjectResponse
	FindAll(ctx context.Context, request webBuildingProject.BuildingProjectRequestFindAll, actor Actor) ([]webBuildingProject.BuildingProjectResponse, int)
	FindById(ctx context.Context, id int, actor Actor) webBuildingProject.BuildingProjectResponse
	Update(ctx context.Context, request webBuildingProject.SaveBuildingProjectRequest, id int, actor Actor) webBuildingProject.BuildingProjectResponse
	Delete(ctx context.Context, id int, actor Actor)

	// FindChanges is the history of one project, newest first.
	FindChanges(ctx context.Context, projectId int, take int, skip int, actor Actor) ([]webBuildingProject.BuildingProjectChangeResponse, int)

	// Import checks and counts every row; with dryRun it stops there and writes
	// nothing, so an upload can be previewed. Blank means CLEAR on this import, so
	// the result reports cleared fields separately from updated rows.
	Import(ctx context.Context, fileBytes []byte, fileType string, dryRun bool, actor Actor) web.ImportResult
	Export(ctx context.Context, actor Actor) ([]byte, error)
	Template() ([]byte, error)

	// Vocabulary serves the closed vocabularies to the form, so its dropdowns and
	// the importer's validation cannot drift apart. A hard-coded list in the client
	// would disagree with this one within a release or two, and the disagreement
	// would surface as a rejected upload rather than as a broken dropdown.
	Vocabulary() map[string][]string
}
