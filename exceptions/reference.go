package exceptions

import (
	"crypto/rand"
	"encoding/hex"
	"os"
	"strings"
)

// NewErrorReference returns a short, quotable id for one failure, e.g. "ERR-7F3A2C".
//
// The same id goes to the screen and to the log line carrying the real error, so
// support can move from what the operator read out to the stack trace with one
// grep. Six hex characters is 16 million values: not unique forever, but it only
// has to be unique among the failures someone might be looking at in the same
// afternoon, and it has to survive being read aloud over the phone.
func NewErrorReference() string {
	buf := make([]byte, 3)
	if _, err := rand.Read(buf); err != nil {
		// Never fail to report a failure just because we could not name it.
		return "ERR-UNKNOWN"
	}

	return "ERR-" + strings.ToUpper(hex.EncodeToString(buf))
}

// GenericServerMessage is what a 500 says to whoever is looking at the screen.
//
// It deliberately never carries err.Error(). That string is written for us at 2am
// with a stack trace beside it, and it has shipped constraint names, SQL and --
// through pgx connection errors -- the database host and username to the browser.
// The real error goes to the log under the same reference.
//
// It does not promise that nothing was saved. A panic inside a transaction does
// roll back (see helpers.CommitOrRollback), but not every failure is inside one:
// an image already written to disk, or a panic after commit, would make that
// promise a lie exactly when someone is relying on it.
func GenericServerMessage(reference string) string {
	return "Something went wrong on our side. Please try again — if it keeps happening, " +
		"quote reference " + reference + " so we can trace it."
}

// debugErrorsEnabled reports whether the technical detail may ride along in the
// response, for local work where nobody wants to tail a log to see a typo.
//
// It is opt-in through DEBUG_ERRORS rather than inferred from ENVIRONMENT on
// purpose: an unset or misspelled variable must fail closed. Defaulting to "show
// the raw error unless told otherwise" is how a forgotten variable in a new
// production environment turns into a disclosure.
func debugErrorsEnabled() bool {
	return strings.EqualFold(strings.TrimSpace(os.Getenv("DEBUG_ERRORS")), "true")
}
