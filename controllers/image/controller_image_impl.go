package image

import (
	"io"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/malikabdulaziz/tmn-backend/helpers"
)

type ControllerImageImpl struct{}

func NewControllerImageImpl() ControllerImageInterface {
	return &ControllerImageImpl{}
}

// ProxyImage proxies image requests to the ERP server
// Route: GET /erp-images/*
func (controller *ControllerImageImpl) ProxyImage(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	// Extract the image path from the request URL
	// The route is /erp-images/*, so we need to get everything after /erp-images/
	requestPath := r.URL.Path
	prefix := "/erp-images/"

	if !strings.HasPrefix(requestPath, prefix) {
		http.Error(w, "Invalid image path", http.StatusBadRequest)
		return
	}

	imagePath := strings.TrimPrefix(requestPath, prefix)
	if imagePath == "" {
		http.Error(w, "Image path is required", http.StatusBadRequest)
		return
	}

	// Get ERP base URL and credentials from environment
	erpBaseURL := os.Getenv("ERP_API_BASE_URL")
	if erpBaseURL == "" {
		http.Error(w, "ERP_API_BASE_URL not configured", http.StatusInternalServerError)
		return
	}

	apiKey := os.Getenv("ERP_API_KEY")
	apiSecret := os.Getenv("ERP_API_SECRET")

	// Construct the full URL to the ERP image server
	// Remove leading slash from imagePath if present
	imagePath = strings.TrimPrefix(imagePath, "/")
	erpBaseURL = strings.TrimSuffix(erpBaseURL, "/")
	erpImageURL := erpBaseURL + "/" + imagePath

	// Create request to ERP server
	req, err := http.NewRequest("GET", erpImageURL, nil)
	if err != nil {
		helpers.GetLogger().WithError(err).Error("Failed to create request to ERP image server")
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}

	// Add authorization header if credentials are available
	if apiKey != "" && apiSecret != "" {
		req.Header.Set("Authorization", "Token "+apiKey+":"+apiSecret)
	}

	// Forward other headers that might be useful
	req.Header.Set("Accept", r.Header.Get("Accept"))
	req.Header.Set("User-Agent", r.Header.Get("User-Agent"))

	// Make request to ERP server
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		helpers.GetLogger().WithError(err).WithField("url", erpImageURL).Error("Failed to fetch image from ERP")
		http.Error(w, "Failed to fetch image", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Check if the response is successful
	if resp.StatusCode != http.StatusOK {
		helpers.GetLogger().WithFields(map[string]interface{}{
			"status": resp.StatusCode,
			"url":    erpImageURL,
		}).Warn("ERP image server returned non-200 status")
		http.Error(w, "Image not found", resp.StatusCode)
		return
	}

	// Set content type from response or infer from file extension
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		ext := strings.ToLower(path.Ext(imagePath))
		switch ext {
		case ".jpg", ".jpeg":
			contentType = "image/jpeg"
		case ".png":
			contentType = "image/png"
		case ".gif":
			contentType = "image/gif"
		case ".webp":
			contentType = "image/webp"
		default:
			contentType = "application/octet-stream"
		}
	}

	// Set response headers
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "public, max-age=3600") // Cache for 1 hour

	// Copy the image data to the response
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		helpers.GetLogger().WithError(err).Error("Failed to copy image data to response")
		return
	}
}
