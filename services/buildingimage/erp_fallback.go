package buildingimage

import (
	"io"
	"net/http"
	"os"
	"strings"
)

// FetchERPImage streams one ERP-hosted photo to w, returning whether it succeeded.
//
// The same thing GET /erp-images/* does, reachable from the service so one route can
// prefer a locally hosted photo and fall back to ERP without the caller choosing.
// ERP is still the source for the 552 buildings whose photos nobody has replaced.
func FetchERPImage(w http.ResponseWriter, erpPath string) bool {
	baseURL := strings.TrimSuffix(os.Getenv("ERP_API_BASE_URL"), "/")
	if baseURL == "" {
		return false
	}

	request, err := http.NewRequest(http.MethodGet, baseURL+"/"+strings.TrimPrefix(erpPath, "/"), nil)
	if err != nil {
		return false
	}

	if key, secret := os.Getenv("ERP_API_KEY"), os.Getenv("ERP_API_SECRET"); key != "" && secret != "" {
		request.Header.Set("Authorization", "Token "+key+":"+secret)
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return false
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return false
	}

	if contentType := response.Header.Get("Content-Type"); contentType != "" {
		w.Header().Set("Content-Type", contentType)
	}
	// ERP photos change rarely and cost a round trip to a third-party server.
	w.Header().Set("Cache-Control", "private, max-age=3600")

	_, err = io.Copy(w, response.Body)

	return err == nil
}
