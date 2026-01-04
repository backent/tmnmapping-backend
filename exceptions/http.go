package exceptions

import (
	"net/http"
	"runtime/debug"

	"github.com/go-playground/validator/v10"
	"github.com/malikabdulaziz/tmn-backend/helpers"
	"github.com/malikabdulaziz/tmn-backend/web"
)

func RouterPanicHandler(w http.ResponseWriter, r *http.Request, i interface{}) {
	var response web.WebResponse
	logger := helpers.GetLogger()

	// Capture stack trace
	stackTrace := string(debug.Stack())

	// Build request context fields
	requestFields := map[string]interface{}{
		"method":      r.Method,
		"path":        r.URL.Path,
		"ip":          r.RemoteAddr,
		"stack_trace": stackTrace,
	}

	if err, ok := i.(validator.ValidationErrors); ok {
		requestFields["status_code"] = http.StatusBadRequest
		logger.WithFields(requestFields).WithField("error", err.Error()).Warn("Validation error")
		response = web.WebResponse{
			Code:   http.StatusBadRequest,
			Status: "BAD REQUEST",
			Data:   err.Error(),
		}
	} else if err, ok := i.(BadRequestError); ok {
		requestFields["status_code"] = http.StatusBadRequest
		logger.WithFields(requestFields).WithField("error", err.Error).Warn("Bad request error")
		response = web.WebResponse{
			Code:   http.StatusBadRequest,
			Status: "BAD REQUEST",
			Data:   err.Error,
		}
	} else if err, ok := i.(Unauthorized); ok {
		requestFields["status_code"] = http.StatusUnauthorized
		logger.WithFields(requestFields).WithField("error", err.Error).Warn("Unauthorized error")
		response = web.WebResponse{
			Code:   http.StatusUnauthorized,
			Status: "Unauthorized",
			Data:   err.Error,
		}
	} else if err, ok := i.(NotFoundError); ok {
		requestFields["status_code"] = http.StatusNotFound
		logger.WithFields(requestFields).WithField("error", err.Error).Warn("Not found error")
		response = web.WebResponse{
			Code:   http.StatusNotFound,
			Status: "NOT FOUND",
			Data:   err.Error,
		}
	} else if err, ok := i.(error); ok {
		requestFields["status_code"] = http.StatusInternalServerError
		logger.WithFields(requestFields).WithField("error", err.Error()).Error("Internal server error")
		response = web.WebResponse{
			Code:   http.StatusInternalServerError,
			Status: "INTERNAL SERVER ERROR",
			Data:   err.Error(),
		}
	} else {
		requestFields["status_code"] = http.StatusInternalServerError
		logger.WithFields(requestFields).WithField("panic_data", i).Error("Unknown panic occurred")
		response = web.WebResponse{
			Code:   http.StatusInternalServerError,
			Status: "INTERNAL SERVER ERROR",
			Data:   i,
		}
	}

	helpers.ReturnReponseJSON(w, response)
}

