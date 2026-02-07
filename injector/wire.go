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
	controllersSalesPackage "github.com/malikabdulaziz/tmn-backend/controllers/salespackage"
	controllersBuildingRestriction "github.com/malikabdulaziz/tmn-backend/controllers/buildingrestriction"
	controllersSavedPolygon "github.com/malikabdulaziz/tmn-backend/controllers/savedpolygon"
	"github.com/malikabdulaziz/tmn-backend/libs"
	"github.com/malikabdulaziz/tmn-backend/middlewares"
	repositoriesAuth "github.com/malikabdulaziz/tmn-backend/repositories/auth"
	repositoriesBuilding "github.com/malikabdulaziz/tmn-backend/repositories/building"
	repositoriesPOI "github.com/malikabdulaziz/tmn-backend/repositories/poi"
	repositoriesSalesPackage "github.com/malikabdulaziz/tmn-backend/repositories/salespackage"
	repositoriesBuildingRestriction "github.com/malikabdulaziz/tmn-backend/repositories/buildingrestriction"
	repositoriesSavedPolygon "github.com/malikabdulaziz/tmn-backend/repositories/savedpolygon"
	repositoriesUser "github.com/malikabdulaziz/tmn-backend/repositories/user"
	servicesAuth "github.com/malikabdulaziz/tmn-backend/services/auth"
	servicesBuilding "github.com/malikabdulaziz/tmn-backend/services/building"
	servicesPOI "github.com/malikabdulaziz/tmn-backend/services/poi"
	servicesSalesPackage "github.com/malikabdulaziz/tmn-backend/services/salespackage"
	servicesBuildingRestriction "github.com/malikabdulaziz/tmn-backend/services/buildingrestriction"
	servicesSavedPolygon "github.com/malikabdulaziz/tmn-backend/services/savedpolygon"
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

var salespackageSet = wire.NewSet(
	repositoriesSalesPackage.NewRepositorySalesPackageImpl,
	servicesSalesPackage.NewServiceSalesPackageImpl,
	controllersSalesPackage.NewControllerSalesPackageImpl,
)

var buildingrestrictionSet = wire.NewSet(
	repositoriesBuildingRestriction.NewRepositoryBuildingRestrictionImpl,
	servicesBuildingRestriction.NewServiceBuildingRestrictionImpl,
	controllersBuildingRestriction.NewControllerBuildingRestrictionImpl,
)

var savedpolygonSet = wire.NewSet(
	repositoriesSavedPolygon.NewRepositorySavedPolygonImpl,
	servicesSavedPolygon.NewServiceSavedPolygonImpl,
	controllersSavedPolygon.NewControllerSavedPolygonImpl,
)

var middlewareSet = wire.NewSet(
	middlewares.NewAuthMiddleware,
	middlewares.NewBuildingMiddleware,
	middlewares.NewPOIMiddleware,
	middlewares.NewSalesPackageMiddleware,
	middlewares.NewBuildingRestrictionMiddleware,
	middlewares.NewSavedPolygonMiddleware,
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
		salespackageSet,
		buildingrestrictionSet,
		savedpolygonSet,
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
		repositoriesPOI.NewRepositoryPOIImpl,
		servicesBuilding.NewServiceBuildingImpl,
	)
	return nil
}
