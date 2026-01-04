package helpers

import (
	"encoding/json"
	"net/http"

	"github.com/malikabdulaziz/tmn-backend/web"
)

func DecodeRequest(r *http.Request, requestVar interface{}) {
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(requestVar)
	PanicIfError(err)
}

func ReturnReponseJSON(w http.ResponseWriter, response web.WebResponse) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(response.Code)
	encoder := json.NewEncoder(w)
	encoder.Encode(response)
}

