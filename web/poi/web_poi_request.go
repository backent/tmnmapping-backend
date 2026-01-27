package poi

type POIPointRequest struct {
	PlaceName string  `json:"place_name" validate:"required"`
	Address   string  `json:"address" validate:"required"`
	Latitude  float64 `json:"latitude" validate:"required"`
	Longitude float64 `json:"longitude" validate:"required"`
}

type CreatePOIRequest struct {
	Name   string           `json:"name" validate:"required"`
	Color  string           `json:"color" validate:"required"`
	Points []POIPointRequest `json:"points" validate:"required,min=1,dive"`
}

type UpdatePOIRequest struct {
	Name   string           `json:"name" validate:"required"`
	Color  string           `json:"color" validate:"required"`
	Points []POIPointRequest `json:"points" validate:"required,min=1,dive"`
}
