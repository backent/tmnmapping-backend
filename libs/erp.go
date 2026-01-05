package libs

import (
	"os"

	"github.com/malikabdulaziz/tmn-backend/services/erp"
)

// ProvideERPClient provides the ERP client instance for dependency injection
func ProvideERPClient() *erp.ERPClient {
	baseURL := os.Getenv("ERP_API_BASE_URL")
	if baseURL == "" {
		baseURL = "https://erp.example.com"
	}

	apiKey := os.Getenv("ERP_API_KEY")
	apiSecret := os.Getenv("ERP_API_SECRET")

	return erp.NewERPClient(baseURL, apiKey, apiSecret)
}

