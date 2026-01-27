package libs

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	controllersAuth "github.com/malikabdulaziz/tmn-backend/controllers/auth"
	controllersBuilding "github.com/malikabdulaziz/tmn-backend/controllers/building"
	controllersImage "github.com/malikabdulaziz/tmn-backend/controllers/image"
	controllersPOI "github.com/malikabdulaziz/tmn-backend/controllers/poi"
	"github.com/malikabdulaziz/tmn-backend/exceptions"
	"github.com/malikabdulaziz/tmn-backend/middlewares"
)

func NewRouter(
	authMiddleware *middlewares.AuthMiddleware,
	buildingMiddleware *middlewares.BuildingMiddleware,
	poiMiddleware *middlewares.POIMiddleware,
	loggingMiddleware *middlewares.LoggingMiddleware,
	controllersAuth controllersAuth.ControllerAuthInterface,
	controllersBuilding controllersBuilding.ControllerBuildingInterface,
	controllersImage controllersImage.ControllerImageInterface,
	controllersPOI controllersPOI.ControllerPOIInterface,
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

	return router
}
