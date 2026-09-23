package libs

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	controllersAuth "github.com/malikabdulaziz/tmn-backend/controllers/auth"
	controllersBranch "github.com/malikabdulaziz/tmn-backend/controllers/branch"
	controllersBrand "github.com/malikabdulaziz/tmn-backend/controllers/brand"
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
	"github.com/malikabdulaziz/tmn-backend/exceptions"
	"github.com/malikabdulaziz/tmn-backend/middlewares"
	"github.com/malikabdulaziz/tmn-backend/models"
)

func NewRouter(
	authMiddleware *middlewares.AuthMiddleware,
	buildingMiddleware *middlewares.BuildingMiddleware,
	poiMiddleware *middlewares.POIMiddleware,
	salesPackageMiddleware *middlewares.SalesPackageMiddleware,
	buildingRestrictionMiddleware *middlewares.BuildingRestrictionMiddleware,
	savedPolygonMiddleware *middlewares.SavedPolygonMiddleware,
	loggingMiddleware *middlewares.LoggingMiddleware,
	categoryMiddleware *middlewares.CategoryMiddleware,
	subCategoryMiddleware *middlewares.SubCategoryMiddleware,
	motherBrandMiddleware *middlewares.MotherBrandMiddleware,
	branchMiddleware *middlewares.BranchMiddleware,
	userMiddleware *middlewares.UserMiddleware,
	customerMiddleware *middlewares.CustomerMiddleware,
	brandMiddleware *middlewares.BrandMiddleware,
	salesAssignmentMiddleware *middlewares.SalesAssignmentMiddleware,
	rateCardMiddleware *middlewares.RateCardMiddleware,
	quotationMiddleware *middlewares.QuotationMiddleware,
	controllersAuth controllersAuth.ControllerAuthInterface,
	controllersBuilding controllersBuilding.ControllerBuildingInterface,
	controllersImage controllersImage.ControllerImageInterface,
	controllersPOI controllersPOI.ControllerPOIInterface,
	controllersSalesPackage controllersSalesPackage.ControllerSalesPackageInterface,
	controllersBuildingRestriction controllersBuildingRestriction.ControllerBuildingRestrictionInterface,
	controllersSavedPolygon controllersSavedPolygon.ControllerSavedPolygonInterface,
	controllersDashboard controllersDashboard.ControllerDashboardInterface,
	controllersCategory controllersCategory.ControllerCategoryInterface,
	controllersSubCategory controllersSubCategory.ControllerSubCategoryInterface,
	controllersMotherBrand controllersMotherBrand.ControllerMotherBrandInterface,
	controllersBranch controllersBranch.ControllerBranchInterface,
	controllersUser controllersUser.ControllerUserInterface,
	controllersCustomer controllersCustomer.ControllerCustomerInterface,
	controllersBrand controllersBrand.ControllerBrandInterface,
	controllersSalesAssignment controllersSalesAssignment.ControllerSalesAssignmentInterface,
	controllersRateCard controllersRateCard.ControllerRateCardInterface,
	controllersQuotation controllersQuotation.ControllerQuotationInterface,
	buildingPriceMiddleware *middlewares.BuildingPriceMiddleware,
	controllersBuildingPrice controllersBuildingPrice.ControllerBuildingPriceInterface,
	buildingProjectMiddleware *middlewares.BuildingProjectMiddleware,
	controllersBuildingProject controllersBuildingProject.ControllerBuildingProjectInterface,
) *httprouter.Router {
	router := httprouter.New()

	// Set panic handler
	router.PanicHandler = exceptions.RouterPanicHandler

	// Health check (no logging to avoid spam)
	router.GET("/health", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Public routes (with logging)
	router.POST("/login",
		loggingMiddleware.Log(
			authMiddleware.ValidateLogin(controllersAuth.Login)))

	router.POST("/logout",
		loggingMiddleware.Log(controllersAuth.Logout))

	// Protected routes (with logging)
	router.GET("/current-user",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(controllersAuth.CurrentUser)))

	// Building routes (protected)
	router.GET("/buildings",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequirePermission(models.PermissionBuildingsView)(controllersBuilding.FindAll))))

	router.GET("/buildings/:id",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequirePermission(models.PermissionBuildingsView)(controllersBuilding.FindById))))

	router.PUT("/buildings/:id",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequirePermission(models.PermissionBuildingsManage)(
				buildingMiddleware.ValidateUpdate(controllersBuilding.Update)))))

	router.POST("/buildings/sync",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequirePermission(models.PermissionBuildingsManage)(controllersBuilding.SyncManual))))

	// Spreadsheet maintenance for buildings. dry_run=true previews without writing.
	//
	// A blank cell CLEARS on this import, unlike the price and brand imports where a
	// blank leaves the value alone. Everything it changes is written to
	// building_changes, which is what makes that survivable.
	router.POST("/buildings-import",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(
				authMiddleware.RequirePermission(models.PermissionBuildingsManage)(controllersBuilding.Import))))

	router.GET("/buildings-export",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(
				authMiddleware.RequirePermission(models.PermissionBuildingsView)(controllersBuilding.Export))))

	router.GET("/buildings-template",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(
				authMiddleware.RequirePermission(models.PermissionBuildingsView)(controllersBuilding.Template))))

	// The recovery path after a destructive upload, so it is readable by anyone who
	// can read buildings rather than gated behind manage.
	router.GET("/buildings/:id/changes",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(
				authMiddleware.RequirePermission(models.PermissionBuildingsView)(controllersBuilding.FindChanges))))

	router.GET("/building-filter-options",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequirePermission(models.PermissionMappingView)(controllersBuilding.GetFilterOptions))))

	router.GET("/building-dropdown",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequirePermission(models.PermissionMappingView)(controllersBuilding.GetDropdownOptions))))

	router.POST("/mapping-buildings",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequirePermission(models.PermissionMappingView)(controllersBuilding.FindAllForMapping))))

	router.POST("/admin/mapping-building/export",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequirePermission(models.PermissionBuildingsManage)(controllersBuilding.ExportMappingBuildings))))

	// Image proxy route (protected)
	router.GET("/erp-images/*filepath",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequirePermission(models.PermissionMappingView)(controllersImage.ProxyImage))))

	// POI routes (protected)
	router.POST("/pois-import",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequirePermission(models.PermissionPOIsManage)(controllersPOI.Import))))

	router.GET("/pois-export",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequirePermission(models.PermissionPOIsManage)(controllersPOI.Export))))

	router.POST("/pois",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequirePermission(models.PermissionPOIsManage)(
				poiMiddleware.ValidateCreate(controllersPOI.Create)))))

	router.GET("/pois",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequirePermission(models.PermissionPOIsView)(controllersPOI.FindAll))))

	router.GET("/pois/:id",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequirePermission(models.PermissionPOIsView)(controllersPOI.FindById))))

	router.PUT("/pois/:id",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequirePermission(models.PermissionPOIsManage)(
				poiMiddleware.ValidateUpdate(controllersPOI.Update)))))

	router.DELETE("/pois/:id",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequirePermission(models.PermissionPOIsManage)(controllersPOI.Delete))))

	// Sales package routes (protected)
	router.POST("/sales-packages",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequirePermission(models.PermissionSalesPackagesManage)(
				salesPackageMiddleware.ValidateCreate(controllersSalesPackage.Create)))))

	router.GET("/sales-packages",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequirePermission(models.PermissionSalesPackagesView)(controllersSalesPackage.FindAll))))

	router.GET("/sales-packages/:id",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequirePermission(models.PermissionSalesPackagesView)(controllersSalesPackage.FindById))))

	router.PUT("/sales-packages/:id",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequirePermission(models.PermissionSalesPackagesManage)(
				salesPackageMiddleware.ValidateUpdate(controllersSalesPackage.Update)))))

	router.DELETE("/sales-packages/:id",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequirePermission(models.PermissionSalesPackagesManage)(controllersSalesPackage.Delete))))

	router.POST("/sales-packages-import",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequirePermission(models.PermissionSalesPackagesManage)(controllersSalesPackage.Import))))

	router.GET("/sales-packages-export",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequirePermission(models.PermissionSalesPackagesManage)(controllersSalesPackage.Export))))

	// Building restriction routes (protected)
	router.POST("/building-restrictions",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequirePermission(models.PermissionBuildingRestrictionsManage)(
				buildingRestrictionMiddleware.ValidateCreate(controllersBuildingRestriction.Create)))))

	router.GET("/building-restrictions",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequirePermission(models.PermissionBuildingRestrictionsView)(controllersBuildingRestriction.FindAll))))

	router.GET("/building-restrictions/:id",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequirePermission(models.PermissionBuildingRestrictionsView)(controllersBuildingRestriction.FindById))))

	router.PUT("/building-restrictions/:id",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequirePermission(models.PermissionBuildingRestrictionsManage)(
				buildingRestrictionMiddleware.ValidateUpdate(controllersBuildingRestriction.Update)))))

	router.DELETE("/building-restrictions/:id",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequirePermission(models.PermissionBuildingRestrictionsManage)(controllersBuildingRestriction.Delete))))

	router.POST("/building-restrictions-import",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequirePermission(models.PermissionBuildingRestrictionsManage)(controllersBuildingRestriction.Import))))

	router.GET("/building-restrictions-export",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequirePermission(models.PermissionBuildingRestrictionsManage)(controllersBuildingRestriction.Export))))

	// Saved polygon routes (protected)
	router.POST("/saved-polygons",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequirePermission(models.PermissionSavedPolygonsManage)(
				savedPolygonMiddleware.ValidateCreate(controllersSavedPolygon.Create)))))

	router.GET("/saved-polygons",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequirePermission(models.PermissionSavedPolygonsView)(controllersSavedPolygon.FindAll))))

	router.GET("/saved-polygons/:id",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequirePermission(models.PermissionSavedPolygonsView)(controllersSavedPolygon.FindById))))

	router.PUT("/saved-polygons/:id",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequirePermission(models.PermissionSavedPolygonsManage)(
				savedPolygonMiddleware.ValidateUpdate(controllersSavedPolygon.Update)))))

	router.DELETE("/saved-polygons/:id",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequirePermission(models.PermissionSavedPolygonsManage)(controllersSavedPolygon.Delete))))

	// Category routes (protected)
	router.POST("/categories",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequirePermission(models.PermissionMasterDataManage)(
				categoryMiddleware.ValidateCreate(controllersCategory.Create)))))

	router.GET("/categories",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequirePermission(models.PermissionMasterDataView)(controllersCategory.FindAll))))

	router.GET("/categories-dropdown",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequirePermission(models.PermissionMasterDataView)(controllersCategory.FindAllDropdown))))

	router.GET("/categories/:id",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequirePermission(models.PermissionMasterDataView)(controllersCategory.FindById))))

	router.PUT("/categories/:id",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequirePermission(models.PermissionMasterDataManage)(
				categoryMiddleware.ValidateUpdate(controllersCategory.Update)))))

	router.DELETE("/categories/:id",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequirePermission(models.PermissionMasterDataManage)(controllersCategory.Delete))))

	router.POST("/categories-import",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequirePermission(models.PermissionMasterDataManage)(controllersCategory.Import))))

	router.GET("/categories-export",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequirePermission(models.PermissionMasterDataManage)(controllersCategory.Export))))

	// Sub-Category routes (protected)
	router.POST("/sub-categories",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequirePermission(models.PermissionMasterDataManage)(
				subCategoryMiddleware.ValidateCreate(controllersSubCategory.Create)))))

	router.GET("/sub-categories",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequirePermission(models.PermissionMasterDataView)(controllersSubCategory.FindAll))))

	router.GET("/sub-categories-dropdown",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequirePermission(models.PermissionMasterDataView)(controllersSubCategory.FindAllDropdown))))

	router.GET("/sub-categories/:id",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequirePermission(models.PermissionMasterDataView)(controllersSubCategory.FindById))))

	router.PUT("/sub-categories/:id",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequirePermission(models.PermissionMasterDataManage)(
				subCategoryMiddleware.ValidateUpdate(controllersSubCategory.Update)))))

	router.DELETE("/sub-categories/:id",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequirePermission(models.PermissionMasterDataManage)(controllersSubCategory.Delete))))

	router.POST("/sub-categories-import",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequirePermission(models.PermissionMasterDataManage)(controllersSubCategory.Import))))

	router.GET("/sub-categories-export",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequirePermission(models.PermissionMasterDataManage)(controllersSubCategory.Export))))

	// Mother Brand routes (protected)
	router.POST("/mother-brands",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequirePermission(models.PermissionMasterDataManage)(
				motherBrandMiddleware.ValidateCreate(controllersMotherBrand.Create)))))

	router.GET("/mother-brands",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequirePermission(models.PermissionMasterDataView)(controllersMotherBrand.FindAll))))

	router.GET("/mother-brands-dropdown",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequirePermission(models.PermissionMasterDataView)(controllersMotherBrand.FindAllDropdown))))

	router.GET("/mother-brands/:id",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequirePermission(models.PermissionMasterDataView)(controllersMotherBrand.FindById))))

	router.PUT("/mother-brands/:id",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequirePermission(models.PermissionMasterDataManage)(
				motherBrandMiddleware.ValidateUpdate(controllersMotherBrand.Update)))))

	router.DELETE("/mother-brands/:id",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequirePermission(models.PermissionMasterDataManage)(controllersMotherBrand.Delete))))

	router.POST("/mother-brands-import",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequirePermission(models.PermissionMasterDataManage)(controllersMotherBrand.Import))))

	router.GET("/mother-brands-export",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequirePermission(models.PermissionMasterDataManage)(controllersMotherBrand.Export))))

	// Branch routes (protected)
	router.POST("/branches",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequirePermission(models.PermissionMasterDataManage)(
				branchMiddleware.ValidateCreate(controllersBranch.Create)))))

	router.GET("/branches",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequirePermission(models.PermissionMasterDataView)(controllersBranch.FindAll))))

	router.GET("/branches-dropdown",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequirePermission(models.PermissionMasterDataView)(controllersBranch.FindAllDropdown))))

	router.GET("/branches/:id",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequirePermission(models.PermissionMasterDataView)(controllersBranch.FindById))))

	router.PUT("/branches/:id",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequirePermission(models.PermissionMasterDataManage)(
				branchMiddleware.ValidateUpdate(controllersBranch.Update)))))

	router.DELETE("/branches/:id",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequirePermission(models.PermissionMasterDataManage)(controllersBranch.Delete))))

	router.POST("/branches-import",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequirePermission(models.PermissionMasterDataManage)(controllersBranch.Import))))

	router.GET("/branches-export",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequirePermission(models.PermissionMasterDataManage)(controllersBranch.Export))))

	// User management routes (admin only — these change who can do what)
	router.GET("/users",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequirePermission(models.PermissionUsersView)(controllersUser.FindAll))))

	router.GET("/users/:id",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequirePermission(models.PermissionUsersView)(controllersUser.FindById))))

	router.POST("/users",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequirePermission(models.PermissionUsersManage)(
				userMiddleware.ValidateCreate(controllersUser.Create)))))

	router.PUT("/users/:id",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequirePermission(models.PermissionUsersManage)(
				userMiddleware.ValidateUpdate(controllersUser.Update)))))

	router.DELETE("/users/:id",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequirePermission(models.PermissionUsersManage)(controllersUser.Delete))))

	// Customer routes
	router.GET("/customers",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(
				authMiddleware.RequirePermission(models.PermissionCustomersView)(controllersCustomer.FindAll))))

	router.GET("/customers/:id",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(
				authMiddleware.RequirePermission(models.PermissionCustomersView)(controllersCustomer.FindById))))

	router.POST("/customers",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(
				authMiddleware.RequirePermission(models.PermissionCustomersManage)(
					customerMiddleware.ValidateCreate(controllersCustomer.Create)))))

	router.PUT("/customers/:id",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(
				authMiddleware.RequirePermission(models.PermissionCustomersManage)(
					customerMiddleware.ValidateUpdate(controllersCustomer.Update)))))

	router.DELETE("/customers/:id",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(
				authMiddleware.RequirePermission(models.PermissionCustomersManage)(controllersCustomer.Delete))))

	router.POST("/customers-import",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(
				authMiddleware.RequirePermission(models.PermissionCustomersManage)(controllersCustomer.Import))))

	router.GET("/customers-export",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(
				authMiddleware.RequirePermission(models.PermissionCustomersView)(controllersCustomer.Export))))

	// The blank template is a read: anyone who may see the data may see its shape.
	router.GET("/customers-template",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(
				authMiddleware.RequirePermission(models.PermissionCustomersView)(controllersCustomer.Template))))

	// Brand routes
	router.GET("/brands",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(
				authMiddleware.RequirePermission(models.PermissionBrandsView)(controllersBrand.FindAll))))

	router.GET("/brands/:id",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(
				authMiddleware.RequirePermission(models.PermissionBrandsView)(controllersBrand.FindById))))

	router.POST("/brands",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(
				authMiddleware.RequirePermission(models.PermissionBrandsManage)(
					brandMiddleware.ValidateCreate(controllersBrand.Create)))))

	router.PUT("/brands/:id",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(
				authMiddleware.RequirePermission(models.PermissionBrandsManage)(
					brandMiddleware.ValidateUpdate(controllersBrand.Update)))))

	router.DELETE("/brands/:id",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(
				authMiddleware.RequirePermission(models.PermissionBrandsManage)(controllersBrand.Delete))))

	router.POST("/brands-import",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(
				authMiddleware.RequirePermission(models.PermissionBrandsManage)(controllersBrand.Import))))

	router.GET("/brands-export",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(
				authMiddleware.RequirePermission(models.PermissionBrandsView)(controllersBrand.Export))))

	// The blank template is a read: anyone who may see the data may see its shape.
	router.GET("/brands-template",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(
				authMiddleware.RequirePermission(models.PermissionBrandsView)(controllersBrand.Template))))

	// Sales assignment routes
	router.GET("/sales-assignments",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(
				authMiddleware.RequirePermission(models.PermissionSalesAssignmentsView)(controllersSalesAssignment.FindAll))))

	router.GET("/sales-assignments/:id",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(
				authMiddleware.RequirePermission(models.PermissionSalesAssignmentsView)(controllersSalesAssignment.FindById))))

	router.POST("/sales-assignments",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(
				authMiddleware.RequirePermission(models.PermissionSalesAssignmentsManage)(
					salesAssignmentMiddleware.ValidateCreate(controllersSalesAssignment.Create)))))

	router.PUT("/sales-assignments/:id",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(
				authMiddleware.RequirePermission(models.PermissionSalesAssignmentsManage)(
					salesAssignmentMiddleware.ValidateUpdate(controllersSalesAssignment.Update)))))

	router.DELETE("/sales-assignments/:id",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(
				authMiddleware.RequirePermission(models.PermissionSalesAssignmentsManage)(controllersSalesAssignment.Delete))))

	router.POST("/sales-assignments-import",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(
				authMiddleware.RequirePermission(models.PermissionSalesAssignmentsManage)(controllersSalesAssignment.Import))))

	router.GET("/sales-assignments-export",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(
				authMiddleware.RequirePermission(models.PermissionSalesAssignmentsView)(controllersSalesAssignment.Export))))

	// The blank template is a read: anyone who may see the data may see its shape.
	router.GET("/sales-assignments-template",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(
				authMiddleware.RequirePermission(models.PermissionSalesAssignmentsView)(controllersSalesAssignment.Template))))

	// Rate card routes.
	//
	// Prices are readable by every role -- a quotation cannot be priced otherwise.
	// Editing a draft is admin-only, and publishing is separate again because it
	// changes what every future quotation is priced against.
	router.GET("/rate-cards",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(
				authMiddleware.RequirePermission(models.PermissionRateCardsView)(controllersRateCard.FindAllVersions))))

	// Sibling path, not /rate-cards/current: httprouter refuses a static segment
	// and a wildcard (:id) at the same position. Matches the existing convention
	// used by /categories-dropdown and /pois-export.
	router.GET("/rate-cards-current",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(
				authMiddleware.RequirePermission(models.PermissionRateCardsView)(controllersRateCard.FindCurrentVersion))))

	router.GET("/rate-cards/:id",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(
				authMiddleware.RequirePermission(models.PermissionRateCardsView)(controllersRateCard.FindVersionById))))

	router.POST("/rate-cards",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(
				authMiddleware.RequirePermission(models.PermissionRateCardsManage)(
					rateCardMiddleware.ValidateCreateVersion(controllersRateCard.CreateVersion)))))

	router.PUT("/rate-cards/:id",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(
				authMiddleware.RequirePermission(models.PermissionRateCardsManage)(
					rateCardMiddleware.ValidateUpdateVersion(controllersRateCard.UpdateVersion)))))

	router.DELETE("/rate-cards/:id",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(
				authMiddleware.RequirePermission(models.PermissionRateCardsManage)(controllersRateCard.DeleteVersion))))

	router.POST("/rate-cards/:id/publish",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(
				authMiddleware.RequirePermission(models.PermissionRateCardsPublish)(controllersRateCard.PublishVersion))))

	// Building prices within a version
	router.GET("/rate-cards/:id/building-prices",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(
				authMiddleware.RequirePermission(models.PermissionRateCardsView)(controllersRateCard.FindBuildingPrices))))

	router.PUT("/rate-cards/:id/building-prices",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(
				authMiddleware.RequirePermission(models.PermissionRateCardsManage)(
					rateCardMiddleware.ValidateUpsertBuildingPrice(controllersRateCard.UpsertBuildingPrice)))))

	router.DELETE("/rate-cards/:id/building-prices/:buildingId",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(
				authMiddleware.RequirePermission(models.PermissionRateCardsManage)(controllersRateCard.DeleteBuildingPrice))))

	router.POST("/rate-cards/:id/building-prices-import",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(
				authMiddleware.RequirePermission(models.PermissionRateCardsManage)(controllersRateCard.ImportBuildingPrices))))

	router.GET("/rate-cards/:id/building-prices-export",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(
				authMiddleware.RequirePermission(models.PermissionRateCardsView)(controllersRateCard.ExportBuildingPrices))))

	// Package prices within a version
	router.GET("/rate-cards/:id/package-prices",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(
				authMiddleware.RequirePermission(models.PermissionRateCardsView)(controllersRateCard.FindPackagePrices))))

	router.PUT("/rate-cards/:id/package-prices",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(
				authMiddleware.RequirePermission(models.PermissionRateCardsManage)(
					rateCardMiddleware.ValidateUpsertPackagePrice(controllersRateCard.UpsertPackagePrice)))))

	router.DELETE("/rate-cards/:id/package-prices/:packageId",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(
				authMiddleware.RequirePermission(models.PermissionRateCardsManage)(controllersRateCard.DeletePackagePrice))))

	router.POST("/rate-cards/:id/package-prices-import",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(
				authMiddleware.RequirePermission(models.PermissionRateCardsManage)(controllersRateCard.ImportPackagePrices))))

	router.GET("/rate-cards/:id/package-prices-export",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(
				authMiddleware.RequirePermission(models.PermissionRateCardsView)(controllersRateCard.ExportPackagePrices))))

	// Blank templates are reads: anyone who may see the prices may see their shape.
	router.GET("/rate-card-building-prices-template",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(
				authMiddleware.RequirePermission(models.PermissionRateCardsView)(controllersRateCard.BuildingPriceTemplate))))

	router.GET("/rate-card-package-prices-template",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(
				authMiddleware.RequirePermission(models.PermissionRateCardsView)(controllersRateCard.PackagePriceTemplate))))

	// Building prices. One price per building, no versions -- see migration 021.
	router.GET("/building-prices",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(
				authMiddleware.RequirePermission(models.PermissionBuildingPricesView)(controllersBuildingPrice.FindAll))))

	router.PUT("/building-prices",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(
				authMiddleware.RequirePermission(models.PermissionBuildingPricesManage)(
					buildingPriceMiddleware.ValidateUpsert(controllersBuildingPrice.Upsert)))))

	router.DELETE("/building-prices/:buildingId",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(
				authMiddleware.RequirePermission(models.PermissionBuildingPricesManage)(controllersBuildingPrice.Delete))))

	// dry_run=true previews the file without writing anything.
	router.POST("/building-prices-import",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(
				authMiddleware.RequirePermission(models.PermissionBuildingPricesManage)(controllersBuildingPrice.Import))))

	router.GET("/building-prices-export",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(
				authMiddleware.RequirePermission(models.PermissionBuildingPricesView)(controllersBuildingPrice.Export))))

	router.GET("/building-prices-template",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(
				authMiddleware.RequirePermission(models.PermissionBuildingPricesView)(controllersBuildingPrice.Template))))

	// Building projects. The landlord side of a building -- see migration 023.
	//
	// Reads are open like the other master data; writes are admin. The landlord money
	// is NOT gated here: building-projects.finance decides which COLUMNS a caller
	// receives, inside the service, because everyone may open a project and only
	// finance may see what TMN pays for it. A route-level gate could only hide the
	// whole project.
	router.GET("/building-projects",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(
				authMiddleware.RequirePermission(models.PermissionBuildingProjectsView)(controllersBuildingProject.FindAll))))

	router.GET("/building-projects/:id",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(
				authMiddleware.RequirePermission(models.PermissionBuildingProjectsView)(controllersBuildingProject.FindById))))

	// Who changed what, and when. Finance values inside the history are withheld from
	// callers without the permission; the field and the actor are still named.
	router.GET("/building-projects/:id/changes",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(
				authMiddleware.RequirePermission(models.PermissionBuildingProjectsView)(controllersBuildingProject.FindChanges))))

	router.POST("/building-projects",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(
				authMiddleware.RequirePermission(models.PermissionBuildingProjectsManage)(
					buildingProjectMiddleware.ValidateSave(controllersBuildingProject.Create)))))

	router.PUT("/building-projects/:id",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(
				authMiddleware.RequirePermission(models.PermissionBuildingProjectsManage)(
					buildingProjectMiddleware.ValidateSave(controllersBuildingProject.Update)))))

	router.DELETE("/building-projects/:id",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(
				authMiddleware.RequirePermission(models.PermissionBuildingProjectsManage)(controllersBuildingProject.Delete))))

	// dry_run=true previews the file without writing. A blank cell CLEARS on this
	// import, so the preview reports cleared fields separately from updated rows.
	router.POST("/building-projects-import",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(
				authMiddleware.RequirePermission(models.PermissionBuildingProjectsManage)(controllersBuildingProject.Import))))

	router.GET("/building-projects-export",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(
				authMiddleware.RequirePermission(models.PermissionBuildingProjectsView)(controllersBuildingProject.Export))))

	router.GET("/building-projects-template",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(
				authMiddleware.RequirePermission(models.PermissionBuildingProjectsView)(controllersBuildingProject.Template))))

	// The form's dropdowns come from here, not from a copy in the client.
	router.GET("/building-projects-vocabulary",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(
				authMiddleware.RequirePermission(models.PermissionBuildingProjectsView)(controllersBuildingProject.Vocabulary))))

	// Quotation routes.
	//
	// View and manage are open to every role: the service scopes them per user, so a
	// salesperson sees only their own pipeline and an approver only their queue.
	// Restricting by role here would stop approvers reading what they must decide on.
	router.GET("/quotations",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(
				authMiddleware.RequirePermission(models.PermissionQuotationsView)(controllersQuotation.FindAll))))

	router.GET("/quotations-dashboard",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(
				authMiddleware.RequirePermission(models.PermissionQuotationsView)(controllersQuotation.DashboardCounts))))

	router.GET("/quotations/:id",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(
				authMiddleware.RequirePermission(models.PermissionQuotationsView)(controllersQuotation.FindById))))

	router.POST("/quotations",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(
				authMiddleware.RequirePermission(models.PermissionQuotationsManage)(
					quotationMiddleware.ValidateCreate(controllersQuotation.Create)))))

	router.PUT("/quotations/:id",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(
				authMiddleware.RequirePermission(models.PermissionQuotationsManage)(
					quotationMiddleware.ValidateUpdate(controllersQuotation.Update)))))

	router.DELETE("/quotations/:id",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(
				authMiddleware.RequirePermission(models.PermissionQuotationsManage)(controllersQuotation.Delete))))

	router.POST("/quotations/:id/submit",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(
				authMiddleware.RequirePermission(models.PermissionQuotationsManage)(controllersQuotation.Submit))))

	// Approving and returning are gated separately: only the approver roles, never
	// admin. Which specific person may act is enforced by the service.
	router.POST("/quotations/:id/approve",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(
				authMiddleware.RequirePermission(models.PermissionQuotationsApprove)(controllersQuotation.Approve))))

	router.POST("/quotations/:id/return",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(
				authMiddleware.RequirePermission(models.PermissionQuotationsApprove)(
					quotationMiddleware.ValidateReturn(controllersQuotation.Return)))))

	// Live pricing for the wizard. Runs the same functions submit will, so the
	// figure a seller sees is the figure they get.
	router.POST("/quotations-pricing-preview",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(
				authMiddleware.RequirePermission(models.PermissionQuotationsView)(
					quotationMiddleware.ValidatePricingPreview(controllersQuotation.PreviewPricing)))))

	router.GET("/dashboard/building-lcd-presence",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequirePermission(models.PermissionDashboardView)(controllersBuilding.GetLCDPresenceSummary))))

	// Dashboard report routes (protected)
	router.GET("/dashboard/acquisition",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequirePermission(models.PermissionDashboardView)(controllersDashboard.GetAcquisitionReport))))

	router.GET("/dashboard/building-proposal",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequirePermission(models.PermissionDashboardView)(controllersDashboard.GetBuildingProposalReport))))

	router.GET("/dashboard/loi",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequirePermission(models.PermissionDashboardView)(controllersDashboard.GetLOIReport))))

	return router
}
