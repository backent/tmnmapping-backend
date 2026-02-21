package dashboard

// StatsSummary holds total count and per-status breakdown.
type StatsSummary struct {
	Total    int            `json:"total"`
	ByStatus map[string]int `json:"by_status"`
}

// PersonTypeStat holds building-type breakdown for one person.
type PersonTypeStat struct {
	Person string         `json:"person"`
	ByType map[string]int `json:"by_type"`
}

// PersonStatusStat holds workflow-state breakdown for one person.
type PersonStatusStat struct {
	Person   string         `json:"person"`
	ByStatus map[string]int `json:"by_status"`
}

// DashboardReport is the response for a single resource tab.
type DashboardReport struct {
	Stats          StatsSummary       `json:"stats"`
	ByPersonType   []PersonTypeStat   `json:"by_person_building_type"`
	ByPersonStatus []PersonStatusStat `json:"by_person_status"`
	PICs           []string           `json:"pics"`
}
