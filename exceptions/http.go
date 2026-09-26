package exceptions

import (
	"fmt"
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
			Extras: err.Extras,
		}
	} else if err, ok := i.(Unauthorized); ok {
		requestFields["status_code"] = http.StatusUnauthorized
		logger.WithFields(requestFields).WithField("error", err.Error).Warn("Unauthorized error")
		response = web.WebResponse{
			Code:   http.StatusUnauthorized,
			Status: "Unauthorized",
			Data:   err.Error,
		}
	} else if err, ok := i.(ForbiddenError); ok {
		requestFields["status_code"] = http.StatusForbidden
		logger.WithFields(requestFields).WithField("error", err.Error).Warn("Forbidden error")
		response = web.WebResponse{
			Code:   http.StatusForbidden,
			Status: "FORBIDDEN",
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
		reference := NewErrorReference()
		requestFields["status_code"] = http.StatusInternalServerError
		requestFields["error_reference"] = reference
		logger.WithFields(requestFields).WithField("error", err.Error()).Error("Internal server error")
		response = serverErrorResponse(reference, err.Error())
	} else {
		reference := NewErrorReference()
		requestFields["status_code"] = http.StatusInternalServerError
		requestFields["error_reference"] = reference
		logger.WithFields(requestFields).WithField("panic_data", i).Error("Unknown panic occurred")
		response = serverErrorResponse(reference, fmt.Sprintf("%v", i))
	}

	helpers.ReturnReponseJSON(w, response)
}

// serverErrorResponse builds the 500 an operator sees: a sentence they can act on
// and a reference they can quote, never the underlying error text.
//
// detail is the real error. It rides along under Extras only when DEBUG_ERRORS is
// explicitly on, which is for local work; staging and production leave it unset and
// the detail stays in the log.
func serverErrorResponse(reference string, detail string) web.WebResponse {
	extras := map[string]interface{}{"reference": reference}
	if debugErrorsEnabled() {
		extras["debug_detail"] = detail
	}

	return web.WebResponse{
		Code:   http.StatusInternalServerError,
		Status: "INTERNAL SERVER ERROR",
		Data:   GenericServerMessage(reference),
		Extras: extras,
	}
}
