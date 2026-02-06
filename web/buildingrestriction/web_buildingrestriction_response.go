package buildingrestriction

type BuildingRefResponse struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type BuildingRestrictionResponse struct {
	Id        int                      `json:"id"`
	Name      string                   `json:"name"`
	Buildings []BuildingRefResponse    `json:"buildings"`
	CreatedAt string                   `json:"created_at"`
	UpdatedAt string                   `json:"updated_at"`
}
