package libs

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	controllersAuth "github.com/malikabdulaziz/tmn-backend/controllers/auth"
	controllersBranch "github.com/malikabdulaziz/tmn-backend/controllers/branch"
	controllersBrand "github.com/malikabdulaziz/tmn-backend/controllers/brand"
	controllersBuilding "github.com/malikabdulaziz/tmn-backend/controllers/building"
	controllersBuildingRestriction "github.com/malikabdulaziz/tmn-backend/controllers/buildingrestriction"
	controllersCategory "github.com/malikabdulaziz/tmn-backend/controllers/category"
	controllersCustomer "github.com/malikabdulaziz/tmn-backend/controllers/customer"
	controllersDashboard "github.com/malikabdulaziz/tmn-backend/controllers/dashboard"
	controllersImage "github.com/malikabdulaziz/tmn-backend/controllers/image"
	controllersMotherBrand "github.com/malikabdulaziz/tmn-backend/controllers/motherbrand"
	controllersPOI "github.com/malikabdulaziz/tmn-backend/controllers/poi"
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
