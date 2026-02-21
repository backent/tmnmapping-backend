package libs

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	controllersAuth "github.com/malikabdulaziz/tmn-backend/controllers/auth"
	controllersBuilding "github.com/malikabdulaziz/tmn-backend/controllers/building"
	controllersDashboard "github.com/malikabdulaziz/tmn-backend/controllers/dashboard"
	controllersImage "github.com/malikabdulaziz/tmn-backend/controllers/image"
	controllersPOI "github.com/malikabdulaziz/tmn-backend/controllers/poi"
	controllersSalesPackage "github.com/malikabdulaziz/tmn-backend/controllers/salespackage"
	controllersBuildingRestriction "github.com/malikabdulaziz/tmn-backend/controllers/buildingrestriction"
	controllersSavedPolygon "github.com/malikabdulaziz/tmn-backend/controllers/savedpolygon"
	"github.com/malikabdulaziz/tmn-backend/exceptions"
	"github.com/malikabdulaziz/tmn-backend/middlewares"
)

func NewRouter(
	authMiddleware *middlewares.AuthMiddleware,
	buildingMiddleware *middlewares.BuildingMiddleware,
	poiMiddleware *middlewares.POIMiddleware,
	salesPackageMiddleware *middlewares.SalesPackageMiddleware,
	buildingRestrictionMiddleware *middlewares.BuildingRestrictionMiddleware,
	savedPolygonMiddleware *middlewares.SavedPolygonMiddleware,
	loggingMiddleware *middlewares.LoggingMiddleware,
	controllersAuth controllersAuth.ControllerAuthInterface,
	controllersBuilding controllersBuilding.ControllerBuildingInterface,
	controllersImage controllersImage.ControllerImageInterface,
	controllersPOI controllersPOI.ControllerPOIInterface,
	controllersSalesPackage controllersSalesPackage.ControllerSalesPackageInterface,
	controllersBuildingRestriction controllersBuildingRestriction.ControllerBuildingRestrictionInterface,
	controllersSavedPolygon controllersSavedPolygon.ControllerSavedPolygonInterface,
	controllersDashboard controllersDashboard.ControllerDashboardInterface,
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
			authMiddleware.RequireAuth(
				buildingMiddleware.ValidateUpdate(controllersBuilding.Update))))

	router.POST("/buildings/sync",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(controllersBuilding.SyncManual)))

	router.GET("/building-filter-options",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(controllersBuilding.GetFilterOptions)))

	router.GET("/mapping-buildings",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(controllersBuilding.FindAllForMapping)))

	router.POST("/admin/mapping-building/export",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(controllersBuilding.ExportMappingBuildings)))

	// Image proxy route (protected)
	// Using catch-all pattern - httprouter will match /erp-images/ and everything after
	router.GET("/erp-images/*filepath",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(controllersImage.ProxyImage)))

	// POI routes (protected)
	router.POST("/pois",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(
				poiMiddleware.ValidateCreate(controllersPOI.Create))))

	router.GET("/pois",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(controllersPOI.FindAll)))

	router.GET("/pois/:id",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(controllersPOI.FindById)))

	router.PUT("/pois/:id",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(
				poiMiddleware.ValidateUpdate(controllersPOI.Update))))

	router.DELETE("/pois/:id",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(controllersPOI.Delete)))

	// Sales package routes (protected)
	router.POST("/sales-packages",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(
				salesPackageMiddleware.ValidateCreate(controllersSalesPackage.Create))))

	router.GET("/sales-packages",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(controllersSalesPackage.FindAll)))

	router.GET("/sales-packages/:id",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(controllersSalesPackage.FindById)))

	router.PUT("/sales-packages/:id",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(
				salesPackageMiddleware.ValidateUpdate(controllersSalesPackage.Update))))

	router.DELETE("/sales-packages/:id",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(controllersSalesPackage.Delete)))

	// Building restriction routes (protected)
	router.POST("/building-restrictions",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(
				buildingRestrictionMiddleware.ValidateCreate(controllersBuildingRestriction.Create))))

	router.GET("/building-restrictions",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(controllersBuildingRestriction.FindAll)))

	router.GET("/building-restrictions/:id",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(controllersBuildingRestriction.FindById)))

	router.PUT("/building-restrictions/:id",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(
				buildingRestrictionMiddleware.ValidateUpdate(controllersBuildingRestriction.Update))))

	router.DELETE("/building-restrictions/:id",
		loggingMiddleware.Log(
			authMiddleware.RequireAuth(controllersBuildingRestriction.Delete)))

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
