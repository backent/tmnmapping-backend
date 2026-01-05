package erp

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// ERPBuilding represents the building data from Frappe ERP
type ERPBuilding struct {
	Name               string `json:"name"`
	BuildingName       string `json:"building_name"`
	BuildingId         string `json:"building_id"`
	IrisCode           string `json:"iris_code"`
	BuildingProject    string `json:"building_project"`
	AudienceActual     int    `json:"audience_actual"`
	AudienceProjection int    `json:"audience_projection"`
	CbdArea            string `json:"cbd_area"`
	Eligible           int    `json:"eligible"`
	CompetitorPresence int    `json:"competitor_presence"`
}

// ERPResponse represents the API response from Frappe
type ERPResponse struct {
	Data []ERPBuilding `json:"data"`
}

// ERPClient handles communication with Frappe ERP API
type ERPClient struct {
	BaseURL    string
	APIKey     string
	APISecret  string
	HTTPClient *http.Client
}

// NewERPClient creates a new ERP client instance
func NewERPClient(baseURL, apiKey, apiSecret string) *ERPClient {
	return &ERPClient{
		BaseURL:   baseURL,
		APIKey:    apiKey,
		APISecret: apiSecret,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// FetchBuildings fetches all buildings from the ERP API
func (c *ERPClient) FetchBuildings() ([]ERPBuilding, error) {
	// Build URL with query parameters to get all fields and all records
	url := fmt.Sprintf("%s/api/resource/Building?fields=[\"*\"]&limit_page_length=99999", c.BaseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set authorization header in format: Token API_KEY:API_SECRET
	if c.APIKey != "" && c.APISecret != "" {
		authValue := fmt.Sprintf("Token %s:%s", c.APIKey, c.APISecret)
		req.Header.Set("Authorization", authValue)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch buildings from ERP: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ERP API returned status %d", resp.StatusCode)
	}

	var erpResponse ERPResponse
	if err := json.NewDecoder(resp.Body).Decode(&erpResponse); err != nil {
		return nil, fmt.Errorf("failed to decode ERP response: %w", err)
	}

	return erpResponse.Data, nil
}
