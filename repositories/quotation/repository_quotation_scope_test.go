package quotation

import (
	"regexp"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

var placeholderPattern = regexp.MustCompile(`\$(\d+)`)

// referenced returns the distinct placeholder numbers appearing in a clause.
func referenced(where string) map[int]bool {
	found := map[int]bool{}
	for _, match := range placeholderPattern.FindAllStringSubmatch(where, -1) {
		n, _ := strconv.Atoi(match[1])
		found[n] = true
	}

	return found
}

// The bug this guards against: a query that binds N arguments but never mentions
// $1. Postgres cannot infer a type for a parameter that appears nowhere and rejects
// the whole statement with 42P18, so the failure is total rather than subtle -- and
// it only shows up for scoped queries, which is every list a non-admin loads.
func TestBuildScopeNumbersEveryArgumentContiguously(t *testing.T) {
	cases := map[string]ListScope{
		"unscoped":            {},
		"owner":               {OwnedByUserId: 7},
		"queue":               {AwaitingApprovalByUserId: 9},
		"status":              {Status: "approved"},
		"search":              {Search: "acme"},
		"owner+status":        {OwnedByUserId: 7, Status: "draft"},
		"owner+search":        {OwnedByUserId: 7, Search: "acme"},
		"queue+status+search": {AwaitingApprovalByUserId: 9, Status: "pending_ceo", Search: "acme"},
		"everything":          {OwnedByUserId: 7, AwaitingApprovalByUserId: 9, Status: "draft", Search: "acme"},
	}

	for name, scope := range cases {
		t.Run(name, func(t *testing.T) {
			for _, startAt := range []int{1, 3} {
				where, args := buildScope(scope, startAt)
				used := referenced(where)

				// Every argument must be referenced, at exactly the offset the
				// caller reserved for it.
				for i := range args {
					assert.True(t, used[startAt+i],
						"argument %d has no $%d in %q", i, startAt+i, where)
				}

				// And nothing outside that range may be referenced, or the query
				// would bind fewer arguments than it names.
				for n := range used {
					assert.GreaterOrEqual(t, n, startAt, "placeholder $%d precedes startAt in %q", n, where)
					assert.Less(t, n, startAt+len(args), "placeholder $%d exceeds the arguments in %q", n, where)
				}
			}
		})
	}
}

// FindAll appends LIMIT and OFFSET after the scope, so its numbering has to
// continue from where buildScope stopped.
func TestBuildScopeLeavesRoomForLimitAndOffset(t *testing.T) {
	where, args := buildScope(ListScope{OwnedByUserId: 7, Search: "acme"}, 1)

	assert.Len(t, args, 2)
	assert.Contains(t, where, "$1")
	assert.Contains(t, where, "$2")
	assert.NotContains(t, where, "$3", "$3 and $4 belong to LIMIT and OFFSET")
}

func TestBuildScopeUnscopedBindsNothing(t *testing.T) {
	where, args := buildScope(ListScope{}, 1)

	assert.Empty(t, where)
	assert.Empty(t, args)
}

// The search term is one argument referenced three times, not three arguments.
func TestBuildScopeSearchReusesOnePlaceholder(t *testing.T) {
	where, args := buildScope(ListScope{Search: "acme"}, 1)

	assert.Equal(t, []interface{}{"%acme%"}, args)
	assert.Equal(t, 3, len(placeholderPattern.FindAllString(where, -1)))
	assert.Equal(t, map[int]bool{1: true}, referenced(where))
}
