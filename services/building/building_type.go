package building

import "strings"

// CanonicalBuildingTypes is the authoritative list of building_type values that
// downstream consumers (mapping page filter chips, exports, totals) rely on.
// Order is intentional: it drives the chip layout on the frontend (4 + 4 + 2).
// Any ERP value that does not match (case-insensitive, trimmed) is collapsed to
// "Other" by CanonicalizeBuildingType during sync.
var CanonicalBuildingTypes = []string{
	"Apartment",
	"Office",
	"Hotel",
	"Mall",
	"Golf Course",
	"Tennis & Padel",
	"Yoga Pilates",
	"Dining",
	"Spa & Reflexology",
	"Other",
}

// canonicalBuildingTypeFallback is returned for any ERP value not in CanonicalBuildingTypes.
const canonicalBuildingTypeFallback = "Other"

// canonicalBuildingTypeIndex maps lowercase canonical names to their canonical form
// for O(1) case-insensitive lookup.
var canonicalBuildingTypeIndex = func() map[string]string {
	idx := make(map[string]string, len(CanonicalBuildingTypes))
	for _, name := range CanonicalBuildingTypes {
		idx[strings.ToLower(name)] = name
	}
	return idx
}()

// CanonicalizeBuildingType maps a raw ERP building_type value to one of
// CanonicalBuildingTypes. Matching is case-insensitive and ignores surrounding
// whitespace. Anything not in the list (including empty string) returns "Other".
func CanonicalizeBuildingType(raw string) string {
	key := strings.ToLower(strings.TrimSpace(raw))
	if canonical, ok := canonicalBuildingTypeIndex[key]; ok {
		return canonical
	}
	return canonicalBuildingTypeFallback
}
