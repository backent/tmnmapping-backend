//go:build wireinject
// +build wireinject

package injector

import (
	"github.com/google/wire"
	"github.com/julienschmidt/httprouter"
	controllersAuth "github.com/malikabdulaziz/tmn-backend/controllers/auth"
	controllersBuilding "github.com/malikabdulaziz/tmn-backend/controllers/building"
	controllersImage "github.com/malikabdulaziz/tmn-backend/controllers/image"
	controllersPOI "github.com/malikabdulaziz/tmn-backend/controllers/poi"
	"github.com/malikabdulaziz/tmn-backend/libs"
	"github.com/malikabdulaziz/tmn-backend/middlewares"
	repositoriesAuth "github.com/malikabdulaziz/tmn-backend/repositories/auth"
	repositoriesBuilding "github.com/malikabdulaziz/tmn-backend/repositories/building"
	repositoriesPOI "github.com/malikabdulaziz/tmn-backend/repositories/poi"
	repositoriesUser "github.com/malikabdulaziz/tmn-backend/repositories/user"
	servicesAuth "github.com/malikabdulaziz/tmn-backend/services/auth"
	servicesBuilding "github.com/malikabdulaziz/tmn-backend/services/building"
	servicesPOI "github.com/malikabdulaziz/tmn-backend/services/poi"
)

var authSet = wire.NewSet(
	repositoriesAuth.NewRepositoryAuthJWTImpl,
	repositoriesUser.NewRepositoryUserImpl,
	servicesAuth.NewServiceAuthImpl,
	controllersAuth.NewControllerAuthImpl,
)

var buildingSet = wire.NewSet(
	repositoriesBuilding.NewRepositoryBuildingImpl,
	servicesBuilding.NewServiceBuildingImpl,
	controllersBuilding.NewControllerBuildingImpl,
)

var imageSet = wire.NewSet(
	controllersImage.NewControllerImageImpl,
)

var poiSet = wire.NewSet(
	repositoriesPOI.NewRepositoryPOIImpl,
	servicesPOI.NewServicePOIImpl,
	controllersPOI.NewControllerPOIImpl,
)

var middlewareSet = wire.NewSet(
	middlewares.NewAuthMiddleware,
	middlewares.NewBuildingMiddleware,
	middlewares.NewPOIMiddleware,
	middlewares.NewLoggingMiddleware,
)

func InitializeRouter() *httprouter.Router {
	wire.Build(
		libs.NewDatabase,
		libs.NewValidator,
		libs.NewLogger,
		libs.ProvideERPClient,
		authSet,
		buildingSet,
		imageSet,
		poiSet,
		middlewareSet,
		libs.NewRouter,
	)
	return nil
}

func InitializeBuildingService() servicesBuilding.ServiceBuildingInterface {
	wire.Build(
		libs.NewDatabase,
		libs.NewLogger,
		libs.ProvideERPClient,
		repositoriesBuilding.NewRepositoryBuildingImpl,
		servicesBuilding.NewServiceBuildingImpl,
	)
	return nil
}
