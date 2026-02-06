package building

type MappingBuildingRequest struct {
	buildingType    string
	buildingGrade   string
	year            string
	subdistrict     string
	progress        string
	sellable        string
	connectivity    string
	lcdPresence     string
	salesPackageIds string
	lat             string
	lng             string
	radius          string
	poiId           string
	polygon         string
	minLat          string
	maxLat          string
	minLng          string
	maxLng          string
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

func (r *MappingBuildingRequest) SetPolygon(polygon string) {
	r.polygon = polygon
}

func (r *MappingBuildingRequest) GetPolygon() string {
	return r.polygon
}

func (r *MappingBuildingRequest) SetMinLat(minLat string) {
	r.minLat = minLat
}

func (r *MappingBuildingRequest) GetMinLat() string {
	return r.minLat
}

func (r *MappingBuildingRequest) SetMaxLat(maxLat string) {
	r.maxLat = maxLat
}

func (r *MappingBuildingRequest) GetMaxLat() string {
	return r.maxLat
}

func (r *MappingBuildingRequest) SetMinLng(minLng string) {
	r.minLng = minLng
}

func (r *MappingBuildingRequest) GetMinLng() string {
	return r.minLng
}

func (r *MappingBuildingRequest) SetMaxLng(maxLng string) {
	r.maxLng = maxLng
}

func (r *MappingBuildingRequest) GetMaxLng() string {
	return r.maxLng
}

func (r *MappingBuildingRequest) SetSalesPackageIds(salesPackageIds string) {
	r.salesPackageIds = salesPackageIds
}

func (r *MappingBuildingRequest) GetSalesPackageIds() string {
	return r.salesPackageIds
}
