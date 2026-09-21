//go:build wireinject
// +build wireinject

package injector

import (
	"github.com/google/wire"
	"github.com/julienschmidt/httprouter"
	controllersAuth "github.com/malikabdulaziz/tmn-backend/controllers/auth"
	controllersBranch "github.com/malikabdulaziz/tmn-backend/controllers/branch"
	controllersBrandAdv "github.com/malikabdulaziz/tmn-backend/controllers/brand"
	controllersBuilding "github.com/malikabdulaziz/tmn-backend/controllers/building"
	controllersBuildingPrice "github.com/malikabdulaziz/tmn-backend/controllers/buildingprice"
	controllersBuildingProject "github.com/malikabdulaziz/tmn-backend/controllers/buildingproject"
	controllersBuildingRestriction "github.com/malikabdulaziz/tmn-backend/controllers/buildingrestriction"
	controllersCategory "github.com/malikabdulaziz/tmn-backend/controllers/category"
	controllersCustomer "github.com/malikabdulaziz/tmn-backend/controllers/customer"
	controllersDashboard "github.com/malikabdulaziz/tmn-backend/controllers/dashboard"
	controllersImage "github.com/malikabdulaziz/tmn-backend/controllers/image"
	controllersMotherBrand "github.com/malikabdulaziz/tmn-backend/controllers/motherbrand"
	controllersPOI "github.com/malikabdulaziz/tmn-backend/controllers/poi"
	controllersQuotation "github.com/malikabdulaziz/tmn-backend/controllers/quotation"
	controllersRateCard "github.com/malikabdulaziz/tmn-backend/controllers/ratecard"
	controllersSalesAssignment "github.com/malikabdulaziz/tmn-backend/controllers/salesassignment"
	controllersSalesPackage "github.com/malikabdulaziz/tmn-backend/controllers/salespackage"
	controllersSavedPolygon "github.com/malikabdulaziz/tmn-backend/controllers/savedpolygon"
	controllersSubCategory "github.com/malikabdulaziz/tmn-backend/controllers/subcategory"
	controllersUser "github.com/malikabdulaziz/tmn-backend/controllers/user"
	"github.com/malikabdulaziz/tmn-backend/libs"
	"github.com/malikabdulaziz/tmn-backend/middlewares"
	repositoriesAuth "github.com/malikabdulaziz/tmn-backend/repositories/auth"
	repositoriesBranch "github.com/malikabdulaziz/tmn-backend/repositories/branch"
	repositoriesBrandAdv "github.com/malikabdulaziz/tmn-backend/repositories/brand"
	repositoriesBuilding "github.com/malikabdulaziz/tmn-backend/repositories/building"
	repositoriesBuildingPrice "github.com/malikabdulaziz/tmn-backend/repositories/buildingprice"
	repositoriesBuildingProject "github.com/malikabdulaziz/tmn-backend/repositories/buildingproject"
	repositoriesBuildingRestriction "github.com/malikabdulaziz/tmn-backend/repositories/buildingrestriction"
	repositoriesCategory "github.com/malikabdulaziz/tmn-backend/repositories/category"
	repositoriesCustomer "github.com/malikabdulaziz/tmn-backend/repositories/customer"
	repositoriesDashboard "github.com/malikabdulaziz/tmn-backend/repositories/dashboard"
	repositoriesMotherBrand "github.com/malikabdulaziz/tmn-backend/repositories/motherbrand"
	repositoriesPOI "github.com/malikabdulaziz/tmn-backend/repositories/poi"
	repositoriesQuotation "github.com/malikabdulaziz/tmn-backend/repositories/quotation"
	repositoriesRateCard "github.com/malikabdulaziz/tmn-backend/repositories/ratecard"
	repositoriesSalesAssignment "github.com/malikabdulaziz/tmn-backend/repositories/salesassignment"
	repositoriesSalesPackage "github.com/malikabdulaziz/tmn-backend/repositories/salespackage"
	repositoriesSavedPolygon "github.com/malikabdulaziz/tmn-backend/repositories/savedpolygon"
	repositoriesSubCategory "github.com/malikabdulaziz/tmn-backend/repositories/subcategory"
	repositoriesUser "github.com/malikabdulaziz/tmn-backend/repositories/user"
	servicesAcquisition "github.com/malikabdulaziz/tmn-backend/services/acquisition"
	servicesAuth "github.com/malikabdulaziz/tmn-backend/services/auth"
	servicesBranch "github.com/malikabdulaziz/tmn-backend/services/branch"
	servicesBrandAdv "github.com/malikabdulaziz/tmn-backend/services/brand"
	servicesBuilding "github.com/malikabdulaziz/tmn-backend/services/building"
	servicesBuildingPrice "github.com/malikabdulaziz/tmn-backend/services/buildingprice"
	servicesBuildingProject "github.com/malikabdulaziz/tmn-backend/services/buildingproject"
	servicesBuildingProposal "github.com/malikabdulaziz/tmn-backend/services/buildingproposal"
	servicesBuildingRestriction "github.com/malikabdulaziz/tmn-backend/services/buildingrestriction"
	servicesCategory "github.com/malikabdulaziz/tmn-backend/services/category"
	servicesCustomer "github.com/malikabdulaziz/tmn-backend/services/customer"
	servicesDashboard "github.com/malikabdulaziz/tmn-backend/services/dashboard"
	servicesLOI "github.com/malikabdulaziz/tmn-backend/services/loi"
	servicesMotherBrand "github.com/malikabdulaziz/tmn-backend/services/motherbrand"
	servicesPOI "github.com/malikabdulaziz/tmn-backend/services/poi"
	servicesQuotation "github.com/malikabdulaziz/tmn-backend/services/quotation"
	servicesRateCard "github.com/malikabdulaziz/tmn-backend/services/ratecard"
	servicesSalesAssignment "github.com/malikabdulaziz/tmn-backend/services/salesassignment"
	servicesSalesPackage "github.com/malikabdulaziz/tmn-backend/services/salespackage"
	servicesSavedPolygon "github.com/malikabdulaziz/tmn-backend/services/savedpolygon"
	servicesSubCategory "github.com/malikabdulaziz/tmn-backend/services/subcategory"
	servicesUser "github.com/malikabdulaziz/tmn-backend/services/user"
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

var advertiserSet = wire.NewSet(
	repositoriesCustomer.NewRepositoryCustomerImpl,
	repositoriesBrandAdv.NewRepositoryBrandImpl,
	repositoriesSalesAssignment.NewRepositorySalesAssignmentImpl,
	servicesCustomer.NewServiceCustomerImpl,
	servicesBrandAdv.NewServiceBrandImpl,
	servicesSalesAssignment.NewServiceSalesAssignmentImpl,
	controllersCustomer.NewControllerCustomerImpl,
	controllersBrandAdv.NewControllerBrandImpl,
	controllersSalesAssignment.NewControllerSalesAssignmentImpl,
)

var quotationSet = wire.NewSet(
	repositoriesQuotation.NewRepositoryQuotationImpl,
	servicesQuotation.NewServiceQuotationImpl,
	controllersQuotation.NewControllerQuotationImpl,
)

var rateCardSet = wire.NewSet(
	repositoriesRateCard.NewRepositoryRateCardImpl,
	servicesRateCard.NewServiceRateCardImpl,
	controllersRateCard.NewControllerRateCardImpl,
)

// Building prices: one per building, no versions -- see migration 021.
var buildingPriceSet = wire.NewSet(
	repositoriesBuildingPrice.NewRepositoryBuildingPriceImpl,
	servicesBuildingPrice.NewServiceBuildingPriceImpl,
	controllersBuildingPrice.NewControllerBuildingPriceImpl,
)

// Building projects: the landlord side of a building -- see migration 023.
var buildingProjectSet = wire.NewSet(
	repositoriesBuildingProject.NewRepositoryBuildingProjectImpl,
	servicesBuildingProject.NewServiceBuildingProjectImpl,
	controllersBuildingProject.NewControllerBuildingProjectImpl,
)

var userSet = wire.NewSet(
	servicesUser.NewServiceUserImpl,
	controllersUser.NewControllerUserImpl,
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
	middlewares.NewUserMiddleware,
	middlewares.NewCustomerMiddleware,
	middlewares.NewBrandMiddleware,
	middlewares.NewSalesAssignmentMiddleware,
	middlewares.NewRateCardMiddleware,
	middlewares.NewQuotationMiddleware,
	middlewares.NewBuildingPriceMiddleware,
	middlewares.NewBuildingProjectMiddleware,
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
		userSet,
		advertiserSet,
		rateCardSet,
		quotationSet,
		buildingPriceSet,
		buildingProjectSet,
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
