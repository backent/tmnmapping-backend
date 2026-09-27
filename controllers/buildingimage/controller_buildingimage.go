package buildingimage

import (
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"github.com/malikabdulaziz/tmn-backend/exceptions"
	"github.com/malikabdulaziz/tmn-backend/helpers"
	"github.com/malikabdulaziz/tmn-backend/middlewares"
	servicesBuildingImage "github.com/malikabdulaziz/tmn-backend/services/buildingimage"
	"github.com/malikabdulaziz/tmn-backend/web"
)

type ControllerBuildingImageInterface interface {
	Upload(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	FindByBuilding(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	Delete(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	Serve(w http.ResponseWriter, r *http.Request, p httprouter.Params)
}

type ControllerBuildingImageImpl struct {
	service servicesBuildingImage.ServiceBuildingImageInterface
}

func NewControllerBuildingImageImpl(service servicesBuildingImage.ServiceBuildingImageInterface) ControllerBuildingImageInterface {
	return &ControllerBuildingImageImpl{service: service}
}

func pathId(p httprouter.Params, name string) int {
	id, err := strconv.Atoi(p.ByName(name))
	if err != nil {
		panic(exceptions.NewBadRequest("invalid " + name))
	}

	return id
}

func (c *ControllerBuildingImageImpl) Upload(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	// Capped before reading, so an oversized upload is refused rather than buffered
	// in full and then rejected.
	if err := r.ParseMultipartForm(servicesBuildingImage.MaxUploadBytes); err != nil {
		panic(exceptions.NewBadRequest("could not read the upload"))
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		panic(exceptions.NewBadRequest("no file was sent under the field name 'file'"))
	}
	defer file.Close()

	content := make([]byte, header.Size)
	if _, err := file.Read(content); err != nil && header.Size > 0 {
		panic(exceptions.NewBadRequest("could not read the uploaded file"))
	}

	userId := 0
	if raw, ok := r.Context().Value(middlewares.ContextKeyUserId).(string); ok {
		userId, _ = strconv.Atoi(raw)
	}

	image := c.service.Upload(r.Context(), pathId(p, "id"), p.ByName("slot"),
		content, header.Header.Get("Content-Type"), userId)

	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusCreated, Data: image})
}

func (c *ControllerBuildingImageImpl) FindByBuilding(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	images := c.service.FindByBuilding(r.Context(), pathId(p, "id"))
	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusOK, Data: images})
}

// Delete removes the photo this application hosts. ERP's becomes visible again, so
// this is "stop overriding", not "remove the picture".
func (c *ControllerBuildingImageImpl) Delete(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	c.service.Delete(r.Context(), pathId(p, "id"), p.ByName("slot"))
	helpers.ReturnReponseJSON(w, web.WebResponse{Status: "OK", Code: http.StatusOK, Data: "Image removed"})
}

// Serve returns one photo: ours when we host it, ERP's when we do not. The client
// asks for a building's front photo and never needs to know which side answered.
func (c *ControllerBuildingImageImpl) Serve(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	if !c.service.Serve(r.Context(), w, pathId(p, "id"), p.ByName("slot")) {
		http.NotFound(w, r)
	}
}
