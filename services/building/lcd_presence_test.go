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
		screenCount         int
		expected            string
	}{
		// TMN cases
		{
			name:                "TMN: bast signed, no competitors, has screens",
			competitorPresence:  false,
			competitorExclusive: false,
			workflowState:       "BAST Signed",
			screenCount:         2,
			expected:            "TMN",
		},
		{
			name:                "TMN: case-insensitive workflow state",
			competitorPresence:  false,
			competitorExclusive: false,
			workflowState:       "bast signed",
			screenCount:         1,
			expected:            "TMN",
		},
		{
			name:                "TMN: workflow state with surrounding spaces",
			competitorPresence:  false,
			competitorExclusive: false,
			workflowState:       "  BAST Signed  ",
			screenCount:         1,
			expected:            "TMN",
		},
		// Competitor cases
		{
			name:                "Competitor: presence true, empty workflow state",
			competitorPresence:  true,
			competitorExclusive: false,
			workflowState:       "",
			screenCount:         0,
			expected:            "Competitor",
		},
		{
			name:                "Competitor: exclusive true, empty workflow state",
			competitorPresence:  false,
			competitorExclusive: true,
			workflowState:       "",
			screenCount:         0,
			expected:            "Competitor",
		},
		{
			// When workflowState is "BAST Signed", the Competitor check is skipped
			// (it requires workflowState=="" || !isBastSigned). No other condition matches either,
			// so the function returns "".
			name:                "Empty: exclusive true but BAST Signed workflow skips Competitor check",
			competitorPresence:  false,
			competitorExclusive: true,
			workflowState:       "BAST Signed",
			screenCount:         0,
			expected:            "",
		},
		{
			name:                "Competitor: both presence and exclusive, non-bast workflow",
			competitorPresence:  true,
			competitorExclusive: true,
			workflowState:       "In Progress",
			screenCount:         0,
			expected:            "Competitor",
		},
		// CoExist cases
		{
			name:                "CoExist: bast signed, presence true, exclusive false, has screens",
			competitorPresence:  true,
			competitorExclusive: false,
			workflowState:       "BAST Signed",
			screenCount:         3,
			expected:            "CoExist",
		},
		// Opportunity cases
		{
			name:                "Opportunity: no competitors, empty workflow state",
			competitorPresence:  false,
			competitorExclusive: false,
			workflowState:       "",
			screenCount:         0,
			expected:            "Opportunity",
		},
		{
			name:                "Opportunity: no competitors, non-bast workflow",
			competitorPresence:  false,
			competitorExclusive: false,
			workflowState:       "In Progress",
			screenCount:         0,
			expected:            "Opportunity",
		},
		// TMN does NOT require screenCount > 0 — it only checks isBastSigned + no competitors.
		{
			name:                "TMN: bast signed, no competitors, zero screens still returns TMN",
			competitorPresence:  false,
			competitorExclusive: false,
			workflowState:       "BAST Signed",
			screenCount:         0,
			expected:            "TMN",
		},
		// Empty (no conditions match): exclusive competitor + bast signed workflow
		// falls through all checks and returns ""
		{
			name:                "Empty: both competitors, BAST Signed workflow, has screens",
			competitorPresence:  true,
			competitorExclusive: true,
			workflowState:       "BAST Signed",
			screenCount:         5,
			expected:            "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateLcdPresenceStatus(
				tt.competitorPresence,
				tt.competitorExclusive,
				tt.workflowState,
				tt.screenCount,
			)
			assert.Equal(t, tt.expected, result)
		})
	}
}
