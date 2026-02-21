package erp

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ERPBuilding represents the building data from Frappe ERP
type ERPBuilding struct {
	Name                string  `json:"name"`
	BuildingName        string  `json:"building_name"`
	BuildingId          string  `json:"building_id"`
	IrisCode            string  `json:"iris_code"`
	BuildingProject     string  `json:"building_project"`
	AudienceActual      int     `json:"audience_actual"`
	AudienceProjection  int     `json:"audience_projection"`
	CbdArea             string  `json:"cbd_area"`
	Subdistrict         string  `json:"subdistrict"`
	Citytown            string  `json:"citytown"`
	Province            string  `json:"province"`
	GradeResource       string  `json:"grade_resource"`
	BuildingType        string  `json:"building_type"`
	CompletionYear      int     `json:"completion_year"`
	Latitude            float64 `json:"latitude"`
	Longitude           float64 `json:"longitude"`
	Eligible            int     `json:"eligible"`
	CompetitorPresence  int     `json:"competitor_presence"`
	CompetitorExclusive int     `json:"competitor_exclusive"`
	FrontSidePhoto      string  `json:"front_side_photo"`
	BackSidePhoto       string  `json:"back_side_photo"`
	LeftSidePhoto       string  `json:"left_side_photo"`
	RightSidePhoto      string  `json:"right_side_photo"`
}

// ERPResponse represents the API response from Frappe
type ERPResponse struct {
	Data []ERPBuilding `json:"data"`
}

// ERPAcquisition represents the acquisition data from Frappe ERP.
// RawJSON holds the complete original record from the API (all fields).
type ERPAcquisition struct {
	Name              string          `json:"name"`
	BuildingProject   string          `json:"building_project"`
	Status            string          `json:"status"`
	WorkflowState     string          `json:"workflow_state"`
	AcquisitionPerson string          `json:"acquisition_person"`
	Creation          string          `json:"creation"`
	Modified          string          `json:"modified"`
	RawJSON           json.RawMessage `json:"-"`
}

// ERPBuildingProposal represents the building proposal data from Frappe ERP.
// RawJSON holds the complete original record from the API (all fields).
type ERPBuildingProposal struct {
	Name              string          `json:"name"`
	BuildingProject   string          `json:"building_project"`
	Status            string          `json:"status"`
	WorkflowState     string          `json:"workflow_state"`
	AcquisitionPerson string          `json:"acquisition_person"`
	NumberOfScreen    int             `json:"number_of_screen"`
	Creation          string          `json:"creation"`
	Modified          string          `json:"modified"`
	RawJSON           json.RawMessage `json:"-"`
}

// ERPLetterOfIntent represents the Letter of Intent data from Frappe ERP.
// RawJSON holds the complete original record from the API (all fields).
type ERPLetterOfIntent struct {
	Name              string          `json:"name"`
	BuildingProject   string          `json:"building_project"`
	Status            string          `json:"status"`
	WorkflowState     string          `json:"workflow_state"`
	AcquisitionPerson string          `json:"acquisition_person"`
	NumberOfScreen    int             `json:"number_of_screen"`
	Creation          string          `json:"creation"`
	Modified          string          `json:"modified"`
	RawJSON           json.RawMessage `json:"-"`
}

// erpRawResponse is used to decode ERP list responses into raw per-record JSON.
type erpRawResponse struct {
	Data []json.RawMessage `json:"data"`
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

// FetchAcquisitions fetches all acquisitions from the ERP API.
// Each record's RawJSON field contains the full original JSON object (all fields).
func (c *ERPClient) FetchAcquisitions() ([]ERPAcquisition, error) {
	url := fmt.Sprintf("%s/api/resource/Acquisition?fields=[\"*\"]&limit_page_length=99999", c.BaseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if c.APIKey != "" && c.APISecret != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Token %s:%s", c.APIKey, c.APISecret))
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch acquisitions from ERP: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ERP API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read ERP response body: %w", err)
	}

	var raw erpRawResponse
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("failed to decode ERP response: %w", err)
	}

	result := make([]ERPAcquisition, 0, len(raw.Data))
	for _, item := range raw.Data {
		var a ERPAcquisition
		if err := json.Unmarshal(item, &a); err != nil {
			continue
		}
		a.RawJSON = item
		result = append(result, a)
	}
	return result, nil
}

// FetchBuildingProposals fetches all building proposals from the ERP API.
// Each record's RawJSON field contains the full original JSON object (all fields).
func (c *ERPClient) FetchBuildingProposals() ([]ERPBuildingProposal, error) {
	url := fmt.Sprintf("%s/api/resource/Building Proposal?fields=[\"*\"]&limit_page_length=99999", c.BaseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if c.APIKey != "" && c.APISecret != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Token %s:%s", c.APIKey, c.APISecret))
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch building proposals from ERP: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ERP API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read ERP response body: %w", err)
	}

	var raw erpRawResponse
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("failed to decode ERP response: %w", err)
	}

	result := make([]ERPBuildingProposal, 0, len(raw.Data))
	for _, item := range raw.Data {
		var bp ERPBuildingProposal
		if err := json.Unmarshal(item, &bp); err != nil {
			continue
		}
		bp.RawJSON = item
		result = append(result, bp)
	}
	return result, nil
}

// FetchLOIs fetches all Letters of Intent from the ERP API.
// Each record's RawJSON field contains the full original JSON object (all fields).
func (c *ERPClient) FetchLOIs() ([]ERPLetterOfIntent, error) {
	url := fmt.Sprintf("%s/api/resource/Letter of Intent?fields=[\"*\"]&limit_page_length=99999", c.BaseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if c.APIKey != "" && c.APISecret != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Token %s:%s", c.APIKey, c.APISecret))
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch LOIs from ERP: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ERP API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read ERP response body: %w", err)
	}

	var raw erpRawResponse
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("failed to decode ERP LOI response: %w", err)
	}

	result := make([]ERPLetterOfIntent, 0, len(raw.Data))
	for _, item := range raw.Data {
		var l ERPLetterOfIntent
		if err := json.Unmarshal(item, &l); err != nil {
			continue
		}
		l.RawJSON = item
		result = append(result, l)
	}
	return result, nil
}
