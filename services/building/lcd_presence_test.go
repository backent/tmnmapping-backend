package building

// Same package (not building_test) to access the unexported calculateLcdPresenceStatus function.

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculateLcdPresenceStatus(t *testing.T) {
	tests := []struct {
		name                string
		competitorPresence  bool
		competitorExclusive bool
		workflowState       string
		expected            string
	}{
		// TMN cases
		{
			name:                "TMN: bast signed, no competitors",
			competitorPresence:  false,
			competitorExclusive: false,
			workflowState:       "BAST Signed",
			expected:            "TMN",
		},
		{
			name:                "TMN: case-insensitive workflow state",
			competitorPresence:  false,
			competitorExclusive: false,
			workflowState:       "bast signed",
			expected:            "TMN",
		},
		{
			name:                "TMN: workflow state with surrounding spaces",
			competitorPresence:  false,
			competitorExclusive: false,
			workflowState:       "  BAST Signed  ",
			expected:            "TMN",
		},
		// Competitor cases
		{
			name:                "Competitor: presence true, empty workflow state",
			competitorPresence:  true,
			competitorExclusive: false,
			workflowState:       "",
			expected:            "Competitor",
		},
		{
			name:                "Competitor: exclusive true, empty workflow state",
			competitorPresence:  false,
			competitorExclusive: true,
			workflowState:       "",
			expected:            "Competitor",
		},
		{
			name:                "Competitor: both presence and exclusive, non-bast workflow",
			competitorPresence:  true,
			competitorExclusive: true,
			workflowState:       "In Progress",
			expected:            "Competitor",
		},
		// CoExist cases (no longer depends on screen count)
		{
			name:                "CoExist: bast signed, presence true, exclusive false",
			competitorPresence:  true,
			competitorExclusive: false,
			workflowState:       "BAST Signed",
			expected:            "CoExist",
		},
		// Opportunity cases
		{
			name:                "Opportunity: no competitors, empty workflow state",
			competitorPresence:  false,
			competitorExclusive: false,
			workflowState:       "",
			expected:            "Opportunity",
		},
		{
			name:                "Opportunity: no competitors, non-bast workflow",
			competitorPresence:  false,
			competitorExclusive: false,
			workflowState:       "In Progress",
			expected:            "Opportunity",
		},
		// Empty (contradictory data): BAST Signed combined with competitor_exclusive
		// is logically impossible — falls through to "" as anomaly indicator.
		{
			name:                "Empty: exclusive true with BAST Signed workflow (anomaly)",
			competitorPresence:  false,
			competitorExclusive: true,
			workflowState:       "BAST Signed",
			expected:            "",
		},
		{
			name:                "Empty: both competitors with BAST Signed workflow (anomaly)",
			competitorPresence:  true,
			competitorExclusive: true,
			workflowState:       "BAST Signed",
			expected:            "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateLcdPresenceStatus(
				tt.competitorPresence,
				tt.competitorExclusive,
				tt.workflowState,
			)
			assert.Equal(t, tt.expected, result)
		})
	}
}
