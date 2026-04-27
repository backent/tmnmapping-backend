//go:build wireinject
// +build wireinject

package injector

import (
	"github.com/google/wire"
	"github.com/julienschmidt/httprouter"
	controllersAuth "github.com/malikabdulaziz/tmn-backend/controllers/auth"
	controllersBuilding "github.com/malikabdulaziz/tmn-backend/controllers/building"
	controllersBranch "github.com/malikabdulaziz/tmn-backend/controllers/branch"
	controllersBuildingRestriction "github.com/malikabdulaziz/tmn-backend/controllers/buildingrestriction"
	controllersCategory "github.com/malikabdulaziz/tmn-backend/controllers/category"
	controllersDashboard "github.com/malikabdulaziz/tmn-backend/controllers/dashboard"
	controllersImage "github.com/malikabdulaziz/tmn-backend/controllers/image"
	controllersMotherBrand "github.com/malikabdulaziz/tmn-backend/controllers/motherbrand"
	controllersPOI "github.com/malikabdulaziz/tmn-backend/controllers/poi"
	controllersSalesPackage "github.com/malikabdulaziz/tmn-backend/controllers/salespackage"
	controllersSavedPolygon "github.com/malikabdulaziz/tmn-backend/controllers/savedpolygon"
	controllersSubCategory "github.com/malikabdulaziz/tmn-backend/controllers/subcategory"
	"github.com/malikabdulaziz/tmn-backend/libs"
	"github.com/malikabdulaziz/tmn-backend/middlewares"
	repositoriesAuth "github.com/malikabdulaziz/tmn-backend/repositories/auth"
	repositoriesBranch "github.com/malikabdulaziz/tmn-backend/repositories/branch"
	repositoriesBuilding "github.com/malikabdulaziz/tmn-backend/repositories/building"
	repositoriesBuildingRestriction "github.com/malikabdulaziz/tmn-backend/repositories/buildingrestriction"
	repositoriesCategory "github.com/malikabdulaziz/tmn-backend/repositories/category"
	repositoriesDashboard "github.com/malikabdulaziz/tmn-backend/repositories/dashboard"
	repositoriesMotherBrand "github.com/malikabdulaziz/tmn-backend/repositories/motherbrand"
	repositoriesPOI "github.com/malikabdulaziz/tmn-backend/repositories/poi"
	repositoriesSalesPackage "github.com/malikabdulaziz/tmn-backend/repositories/salespackage"
	repositoriesSavedPolygon "github.com/malikabdulaziz/tmn-backend/repositories/savedpolygon"
	repositoriesSubCategory "github.com/malikabdulaziz/tmn-backend/repositories/subcategory"
	repositoriesUser "github.com/malikabdulaziz/tmn-backend/repositories/user"
	servicesAcquisition "github.com/malikabdulaziz/tmn-backend/services/acquisition"
	servicesAuth "github.com/malikabdulaziz/tmn-backend/services/auth"
	servicesBranch "github.com/malikabdulaziz/tmn-backend/services/branch"
	servicesBuilding "github.com/malikabdulaziz/tmn-backend/services/building"
	servicesBuildingProposal "github.com/malikabdulaziz/tmn-backend/services/buildingproposal"
	servicesBuildingRestriction "github.com/malikabdulaziz/tmn-backend/services/buildingrestriction"
	servicesCategory "github.com/malikabdulaziz/tmn-backend/services/category"
	servicesDashboard "github.com/malikabdulaziz/tmn-backend/services/dashboard"
	servicesLOI "github.com/malikabdulaziz/tmn-backend/services/loi"
	servicesMotherBrand "github.com/malikabdulaziz/tmn-backend/services/motherbrand"
	servicesPOI "github.com/malikabdulaziz/tmn-backend/services/poi"
	servicesSalesPackage "github.com/malikabdulaziz/tmn-backend/services/salespackage"
	servicesSavedPolygon "github.com/malikabdulaziz/tmn-backend/services/savedpolygon"
	servicesSubCategory "github.com/malikabdulaziz/tmn-backend/services/subcategory"
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

var categorySet = wire.NewSet(
	repositoriesCategory.NewRepositoryCategoryImpl,
	servicesCategory.NewServiceCategoryImpl,
	controllersCategory.NewControllerCategoryImpl,
)

var subCategorySet = wire.NewSet(
	repositoriesSubCategory.NewRepositorySubCategoryImpl,
	servicesSubCategory.NewServiceSubCategoryImpl,
	controllersSubCategory.NewControllerSubCategoryImpl,
)

var motherBrandSet = wire.NewSet(
	repositoriesMotherBrand.NewRepositoryMotherBrandImpl,
	servicesMotherBrand.NewServiceMotherBrandImpl,
	controllersMotherBrand.NewControllerMotherBrandImpl,
)

var branchSet = wire.NewSet(
	repositoriesBranch.NewRepositoryBranchImpl,
	servicesBranch.NewServiceBranchImpl,
	controllersBranch.NewControllerBranchImpl,
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

var dashboardSet = wire.NewSet(
	repositoriesDashboard.NewRepositoryDashboardImpl,
	servicesDashboard.NewServiceDashboardImpl,
	controllersDashboard.NewControllerDashboardImpl,
)

var middlewareSet = wire.NewSet(
	middlewares.NewAuthMiddleware,
	middlewares.NewBuildingMiddleware,
	middlewares.NewPOIMiddleware,
	middlewares.NewSalesPackageMiddleware,
	middlewares.NewBuildingRestrictionMiddleware,
	middlewares.NewSavedPolygonMiddleware,
	middlewares.NewLoggingMiddleware,
	middlewares.NewCategoryMiddleware,
	middlewares.NewSubCategoryMiddleware,
	middlewares.NewMotherBrandMiddleware,
	middlewares.NewBranchMiddleware,
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
		categorySet,
		subCategorySet,
		motherBrandSet,
		branchSet,
		poiSet,
		salespackageSet,
		buildingrestrictionSet,
		savedpolygonSet,
		dashboardSet,
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

func InitializeAcquisitionService() servicesAcquisition.ServiceAcquisitionInterface {
	wire.Build(
		libs.NewDatabase,
		libs.NewLogger,
		libs.ProvideERPClient,
		servicesAcquisition.NewServiceAcquisitionImpl,
	)
	return nil
}

func InitializeBuildingProposalService() servicesBuildingProposal.ServiceBuildingProposalInterface {
	wire.Build(
		libs.NewDatabase,
		libs.NewLogger,
		libs.ProvideERPClient,
		servicesBuildingProposal.NewServiceBuildingProposalImpl,
	)
	return nil
}

func InitializeLOIService() servicesLOI.ServiceLOIInterface {
	wire.Build(
		libs.NewDatabase,
		libs.NewLogger,
		libs.ProvideERPClient,
		servicesLOI.NewServiceLOIImpl,
	)
	return nil
}
