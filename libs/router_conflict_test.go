package libs_test

import (
	"net/http"
	"os"
	"regexp"
	"testing"

	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// httprouter panics at registration time when two routes conflict -- for example a
// static segment and a wildcard at the same position, as in
//
//	GET /rate-cards/current
//	GET /rate-cards/:id
//
// That panic happens when NewRouter runs, which is at process startup, so a conflict
// takes the whole service down on deploy rather than failing a test.
//
// NewRouter itself cannot be called here: it needs every controller, and building
// those needs a live database. So this reads the route table straight out of
// router.go and replays it into a bare router. It cannot drift from the real routes,
// because it is parsing them.
var routePattern = regexp.MustCompile(`router\.(GET|POST|PUT|DELETE|PATCH|HEAD)\("([^"]+)"`)

func TestRouterHasNoConflictingRoutes(t *testing.T) {
	source, err := os.ReadFile("router.go")
	require.NoError(t, err, "router.go must be readable")

	matches := routePattern.FindAllStringSubmatch(string(source), -1)
	require.NotEmpty(t, matches, "no routes found -- has router.go moved?")

	noop := func(http.ResponseWriter, *http.Request, httprouter.Params) {}
	router := httprouter.New()

	for _, match := range matches {
		method, path := match[1], match[2]

		// Handle() panics on a conflict; surface it as a readable failure naming
		// the offending route instead of a raw stack trace.
		func() {
			defer func() {
				if recovered := recover(); recovered != nil {
					t.Errorf("route %s %s conflicts: %v", method, path, recovered)
				}
			}()
			router.Handle(method, path, noop)
		}()
	}

	t.Logf("registered %d routes with no conflicts", len(matches))
}

// A duplicate method+path is silently wrong rather than a panic in some routers, and
// in httprouter it panics too -- either way it means one of them is dead.
func TestRouterHasNoDuplicateRoutes(t *testing.T) {
	source, err := os.ReadFile("router.go")
	require.NoError(t, err)

	seen := map[string]bool{}
	for _, match := range routePattern.FindAllStringSubmatch(string(source), -1) {
		key := match[1] + " " + match[2]
		assert.False(t, seen[key], "route %s is registered more than once", key)
		seen[key] = true
	}
}
