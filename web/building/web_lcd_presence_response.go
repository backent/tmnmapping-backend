package building

// LCDPresenceCitySummary holds the aggregated LCD presence counts and percentages for one city
type LCDPresenceCitySummary struct {
	Citytown    string             `json:"citytown"`
	Total       int                `json:"total"`
	ByStatus    map[string]int     `json:"by_status"`
	Percentages map[string]float64 `json:"percentages"`
}

// LCDPresenceTotals holds grand-total LCD presence counts and percentages across all cities
type LCDPresenceTotals struct {
	Total       int                `json:"total"`
	ByStatus    map[string]int     `json:"by_status"`
	Percentages map[string]float64 `json:"percentages"`
}

// LCDPresenceSummaryResponse is the full response for the dashboard LCD presence endpoint
type LCDPresenceSummaryResponse struct {
	Data   []LCDPresenceCitySummary `json:"data"`
	Totals LCDPresenceTotals        `json:"totals"`
}
