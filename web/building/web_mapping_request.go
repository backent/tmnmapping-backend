package building

type MappingBuildingRequest struct {
	buildingType  string
	buildingGrade string
	year          string
	subdistrict   string
	progress      string
	sellable      string
	connectivity  string
	lcdPresence   string
	lat           string
	lng           string
	radius        string
	poiId         string
}

func (r *MappingBuildingRequest) SetBuildingType(buildingType string) {
	r.buildingType = buildingType
}

func (r *MappingBuildingRequest) GetBuildingType() string {
	return r.buildingType
}

func (r *MappingBuildingRequest) SetBuildingGrade(buildingGrade string) {
	r.buildingGrade = buildingGrade
}

func (r *MappingBuildingRequest) GetBuildingGrade() string {
	return r.buildingGrade
}

func (r *MappingBuildingRequest) SetYear(year string) {
	r.year = year
}

func (r *MappingBuildingRequest) GetYear() string {
	return r.year
}

func (r *MappingBuildingRequest) SetSubdistrict(subdistrict string) {
	r.subdistrict = subdistrict
}

func (r *MappingBuildingRequest) GetSubdistrict() string {
	return r.subdistrict
}

func (r *MappingBuildingRequest) SetProgress(progress string) {
	r.progress = progress
}

func (r *MappingBuildingRequest) GetProgress() string {
	return r.progress
}

func (r *MappingBuildingRequest) SetSellable(sellable string) {
	r.sellable = sellable
}

func (r *MappingBuildingRequest) GetSellable() string {
	return r.sellable
}

func (r *MappingBuildingRequest) SetConnectivity(connectivity string) {
	r.connectivity = connectivity
}

func (r *MappingBuildingRequest) GetConnectivity() string {
	return r.connectivity
}

func (r *MappingBuildingRequest) SetLCDPresence(lcdPresence string) {
	r.lcdPresence = lcdPresence
}

func (r *MappingBuildingRequest) GetLCDPresence() string {
	return r.lcdPresence
}

func (r *MappingBuildingRequest) SetLat(lat string) {
	r.lat = lat
}

func (r *MappingBuildingRequest) GetLat() string {
	return r.lat
}

func (r *MappingBuildingRequest) SetLng(lng string) {
	r.lng = lng
}

func (r *MappingBuildingRequest) GetLng() string {
	return r.lng
}

func (r *MappingBuildingRequest) SetRadius(radius string) {
	r.radius = radius
}

func (r *MappingBuildingRequest) GetRadius() string {
	return r.radius
}

func (r *MappingBuildingRequest) SetPOIId(poiId string) {
	r.poiId = poiId
}

func (r *MappingBuildingRequest) GetPOIId() string {
	return r.poiId
}
