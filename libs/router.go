package libs

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	controllersAuth "github.com/malikabdulaziz/tmn-backend/controllers/auth"
	controllersBranch "github.com/malikabdulaziz/tmn-backend/controllers/branch"
	controllersBuilding "github.com/malikabdulaziz/tmn-backend/controllers/building"
	controllersBuildingRestriction "github.com/malikabdulaziz/tmn-backend/controllers/buildingrestriction"
	controllersCategory "github.com/malikabdulaziz/tmn-backend/controllers/category"
	controllersDashboard "github.com/malikabdulaziz/tmn-backend/controllers/dashboard"
	controllersImage "github.com/malikabdulaziz/tmn-backend/controllers/image"
	controllersMotherBrand "github.com/malikabdulaziz/tmn-backend/controllers/motherbrand"
	controllersPOI "github.com/malikabdulaziz/tmn-backend/controllers/poi"
	controllersSalesPackage "github.com/malikabdulaziz/tmn-backend/controllers/salespackage"
	controllersSavedPolygon "github.com/malikabdulaziz/tmn-backend/controllers/savedpolygon"
	controllersSubCategory "github.com/malikabdulaziz/tmn-backend/controllers/subcategory"
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
			authMiddleware.RequireAuth(controllersBuilding.FindAll)))

	router.GET("/buildings/:id",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(controllersBuilding.FindById)))

	router.PUT("/buildings/:id",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequireRole(models.RoleAdmin)(
				buildingMiddleware.ValidateUpdate(controllersBuilding.Update)))))

	router.POST("/buildings/sync",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequireRole(models.RoleAdmin)(controllersBuilding.SyncManual))))

	router.GET("/building-filter-options",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(controllersBuilding.GetFilterOptions)))

	router.GET("/building-dropdown",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(controllersBuilding.GetDropdownOptions)))

	router.POST("/mapping-buildings",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(controllersBuilding.FindAllForMapping)))

	router.POST("/admin/mapping-building/export",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequireRole(models.RoleAdmin)(controllersBuilding.ExportMappingBuildings))))

	// Image proxy route (protected)
	router.GET("/erp-images/*filepath",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(controllersImage.ProxyImage)))

	// POI routes (protected)
	router.POST("/pois-import",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequireRole(models.RoleAdmin)(controllersPOI.Import))))

	router.GET("/pois-export",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequireRole(models.RoleAdmin)(controllersPOI.Export))))

	router.POST("/pois",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequireRole(models.RoleAdmin)(
				poiMiddleware.ValidateCreate(controllersPOI.Create)))))

	router.GET("/pois",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(controllersPOI.FindAll)))

	router.GET("/pois/:id",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(controllersPOI.FindById)))

	router.PUT("/pois/:id",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequireRole(models.RoleAdmin)(
				poiMiddleware.ValidateUpdate(controllersPOI.Update)))))

	router.DELETE("/pois/:id",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequireRole(models.RoleAdmin)(controllersPOI.Delete))))

	// Sales package routes (protected)
	router.POST("/sales-packages",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequireRole(models.RoleAdmin)(
				salesPackageMiddleware.ValidateCreate(controllersSalesPackage.Create)))))

	router.GET("/sales-packages",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(controllersSalesPackage.FindAll)))

	router.GET("/sales-packages/:id",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(controllersSalesPackage.FindById)))

	router.PUT("/sales-packages/:id",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequireRole(models.RoleAdmin)(
				salesPackageMiddleware.ValidateUpdate(controllersSalesPackage.Update)))))

	router.DELETE("/sales-packages/:id",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequireRole(models.RoleAdmin)(controllersSalesPackage.Delete))))

	router.POST("/sales-packages-import",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequireRole(models.RoleAdmin)(controllersSalesPackage.Import))))

	router.GET("/sales-packages-export",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequireRole(models.RoleAdmin)(controllersSalesPackage.Export))))

	// Building restriction routes (protected)
	router.POST("/building-restrictions",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequireRole(models.RoleAdmin)(
				buildingRestrictionMiddleware.ValidateCreate(controllersBuildingRestriction.Create)))))

	router.GET("/building-restrictions",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(controllersBuildingRestriction.FindAll)))

	router.GET("/building-restrictions/:id",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(controllersBuildingRestriction.FindById)))

	router.PUT("/building-restrictions/:id",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequireRole(models.RoleAdmin)(
				buildingRestrictionMiddleware.ValidateUpdate(controllersBuildingRestriction.Update)))))

	router.DELETE("/building-restrictions/:id",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequireRole(models.RoleAdmin)(controllersBuildingRestriction.Delete))))

	router.POST("/building-restrictions-import",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequireRole(models.RoleAdmin)(controllersBuildingRestriction.Import))))

	router.GET("/building-restrictions-export",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequireRole(models.RoleAdmin)(controllersBuildingRestriction.Export))))

	// Saved polygon routes (protected)
	router.POST("/saved-polygons",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(
				savedPolygonMiddleware.ValidateCreate(controllersSavedPolygon.Create))))

	router.GET("/saved-polygons",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(controllersSavedPolygon.FindAll)))

	router.GET("/saved-polygons/:id",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(controllersSavedPolygon.FindById)))

	router.PUT("/saved-polygons/:id",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(
				savedPolygonMiddleware.ValidateUpdate(controllersSavedPolygon.Update))))

	router.DELETE("/saved-polygons/:id",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(controllersSavedPolygon.Delete)))

	// Category routes (protected)
	router.POST("/categories",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequireRole(models.RoleAdmin)(
				categoryMiddleware.ValidateCreate(controllersCategory.Create)))))

	router.GET("/categories",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(controllersCategory.FindAll)))

	router.GET("/categories-dropdown",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(controllersCategory.FindAllDropdown)))

	router.GET("/categories/:id",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(controllersCategory.FindById)))

	router.PUT("/categories/:id",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequireRole(models.RoleAdmin)(
				categoryMiddleware.ValidateUpdate(controllersCategory.Update)))))

	router.DELETE("/categories/:id",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequireRole(models.RoleAdmin)(controllersCategory.Delete))))

	router.POST("/categories-import",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequireRole(models.RoleAdmin)(controllersCategory.Import))))

	router.GET("/categories-export",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequireRole(models.RoleAdmin)(controllersCategory.Export))))

	// Sub-Category routes (protected)
	router.POST("/sub-categories",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequireRole(models.RoleAdmin)(
				subCategoryMiddleware.ValidateCreate(controllersSubCategory.Create)))))

	router.GET("/sub-categories",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(controllersSubCategory.FindAll)))

	router.GET("/sub-categories-dropdown",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(controllersSubCategory.FindAllDropdown)))

	router.GET("/sub-categories/:id",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(controllersSubCategory.FindById)))

	router.PUT("/sub-categories/:id",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequireRole(models.RoleAdmin)(
				subCategoryMiddleware.ValidateUpdate(controllersSubCategory.Update)))))

	router.DELETE("/sub-categories/:id",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequireRole(models.RoleAdmin)(controllersSubCategory.Delete))))

	router.POST("/sub-categories-import",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequireRole(models.RoleAdmin)(controllersSubCategory.Import))))

	router.GET("/sub-categories-export",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequireRole(models.RoleAdmin)(controllersSubCategory.Export))))

	// Mother Brand routes (protected)
	router.POST("/mother-brands",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequireRole(models.RoleAdmin)(
				motherBrandMiddleware.ValidateCreate(controllersMotherBrand.Create)))))

	router.GET("/mother-brands",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(controllersMotherBrand.FindAll)))

	router.GET("/mother-brands-dropdown",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(controllersMotherBrand.FindAllDropdown)))

	router.GET("/mother-brands/:id",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(controllersMotherBrand.FindById)))

	router.PUT("/mother-brands/:id",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequireRole(models.RoleAdmin)(
				motherBrandMiddleware.ValidateUpdate(controllersMotherBrand.Update)))))

	router.DELETE("/mother-brands/:id",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequireRole(models.RoleAdmin)(controllersMotherBrand.Delete))))

	router.POST("/mother-brands-import",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequireRole(models.RoleAdmin)(controllersMotherBrand.Import))))

	router.GET("/mother-brands-export",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequireRole(models.RoleAdmin)(controllersMotherBrand.Export))))

	// Branch routes (protected)
	router.POST("/branches",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequireRole(models.RoleAdmin)(
				branchMiddleware.ValidateCreate(controllersBranch.Create)))))

	router.GET("/branches",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(controllersBranch.FindAll)))

	router.GET("/branches-dropdown",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(controllersBranch.FindAllDropdown)))

	router.GET("/branches/:id",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(controllersBranch.FindById)))

	router.PUT("/branches/:id",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequireRole(models.RoleAdmin)(
				branchMiddleware.ValidateUpdate(controllersBranch.Update)))))

	router.DELETE("/branches/:id",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequireRole(models.RoleAdmin)(controllersBranch.Delete))))

	router.POST("/branches-import",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequireRole(models.RoleAdmin)(controllersBranch.Import))))

	router.GET("/branches-export",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(authMiddleware.RequireRole(models.RoleAdmin)(controllersBranch.Export))))

	router.GET("/dashboard/building-lcd-presence",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(controllersBuilding.GetLCDPresenceSummary)))

	// Dashboard report routes (protected)
	router.GET("/dashboard/acquisition",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(controllersDashboard.GetAcquisitionReport)))

	router.GET("/dashboard/building-proposal",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(controllersDashboard.GetBuildingProposalReport)))

	router.GET("/dashboard/loi",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(controllersDashboard.GetLOIReport)))

	return router
}
