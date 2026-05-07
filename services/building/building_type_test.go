package building_test

import (
	"testing"

	"github.com/malikabdulaziz/tmn-backend/services/building"
)

func TestCanonicalBuildingTypesOrder(t *testing.T) {
	expected := []string{
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

	if len(building.CanonicalBuildingTypes) != len(expected) {
		t.Fatalf("expected %d canonical types, got %d", len(expected), len(building.CanonicalBuildingTypes))
	}
	for i, name := range expected {
		if building.CanonicalBuildingTypes[i] != name {
			t.Errorf("CanonicalBuildingTypes[%d] = %q, want %q", i, building.CanonicalBuildingTypes[i], name)
		}
	}
}

func TestCanonicalizeBuildingType(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{"exact match", "Apartment", "Apartment"},
		{"lowercase", "apartment", "Apartment"},
		{"uppercase", "APARTMENT", "Apartment"},
		{"surrounding whitespace", "  Hotel  ", "Hotel"},
		{"multi-word exact", "Tennis & Padel", "Tennis & Padel"},
		{"multi-word lowercase", "tennis & padel", "Tennis & Padel"},
		{"multi-word with case mix", "Spa & reflexology", "Spa & Reflexology"},
		{"unknown value falls back to Other", "Mixed Use", "Other"},
		{"legacy Office Building falls back to Other", "Office Building", "Other"},
		{"canonical Dining passes through", "Dining", "Dining"},
		{"non-canonical Dinning falls back to Other", "Dinning", "Other"},
		{"canonical Golf Course passes through", "Golf Course", "Golf Course"},
		{"non-canonical Golfcourse falls back to Other", "Golfcourse", "Other"},
		{"empty string falls back to Other", "", "Other"},
		{"whitespace-only falls back to Other", "   ", "Other"},
		{"Other passes through", "Other", "Other"},
		{"other lowercase passes through", "other", "Other"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := building.CanonicalizeBuildingType(tt.raw)
			if got != tt.want {
				t.Errorf("CanonicalizeBuildingType(%q) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}
