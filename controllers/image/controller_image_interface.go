package image

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type ControllerImageInterface interface {
	ProxyImage(w http.ResponseWriter, r *http.Request, p httprouter.Params)
}

