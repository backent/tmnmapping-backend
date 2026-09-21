package web

// ImportResult reports the outcome of a master-data upload.
//
// Imports are all-or-nothing: every row is validated before anything is written, so
// a non-empty Errors always means Imported is false and the database is untouched.
// A half-applied master-data import is worse than a rejected one — it leaves the
// operator unsure which half landed.
type ImportResult struct {
	Rows    int `json:"rows"`
	Created int `json:"created"`
	Updated int `json:"updated"`

	// Skipped counts rows that were read, were valid, and were deliberately not
	// applied -- a rate card row priced 0 for a building with no screens installed,
	// for example. Reported so the operator can reconcile the row count rather than
	// wondering where the difference went.
	Skipped int `json:"skipped"`

	// Unchanged counts rows whose price already matched, so a preview can say how
	// much an upload will actually change.
	Unchanged int `json:"unchanged"`

	// Cleared counts FIELDS, not rows, that an upload will blank. It exists for
	// imports where a blank cell means "clear this value" rather than "leave it
	// alone": there, clearing is the destructive operation, and a preview that only
	// said "120 projects updated" would bury it. Zero for merge-style imports.
	Cleared int `json:"cleared"`

	// DryRun marks a preview: every row was checked and counted, nothing written.
	DryRun   bool          `json:"dry_run"`
	Imported bool          `json:"imported"`
	Errors   []ImportError `json:"errors"`

	// Notices are per-row facts worth showing that do not reject the row -- above
	// all, which fields an upload is about to clear and what they hold today.
	Notices []ImportError `json:"notices"`
}

// ImportError points at one spreadsheet row. Row is the number the operator sees in
// Excel — row 1 is the header, so data starts at row 2.
type ImportError struct {
	Row     int    `json:"row"`
	Column  string `json:"column"`
	Value   string `json:"value"`
	Message string `json:"message"`
}

// NewImportResult starts an empty result with a non-nil Errors slice, so the JSON
// response carries [] rather than null.
func NewImportResult() *ImportResult {
	return &ImportResult{Errors: []ImportError{}, Notices: []ImportError{}}
}

// AddError records a rejection. Callers keep validating after the first failure so
// the operator sees every problem in one pass instead of fixing them one at a time.
func (r *ImportResult) AddError(row int, column string, value string, message string) {
	r.Errors = append(r.Errors, ImportError{Row: row, Column: column, Value: value, Message: message})
}

// AddNotice records something the operator should see before applying, such as a
// field about to be cleared. Value carries what the field holds TODAY, because after
// the upload that is the one thing no longer visible anywhere but the change log.
func (r *ImportResult) AddNotice(row int, column string, value string, message string) {
	r.Notices = append(r.Notices, ImportError{Row: row, Column: column, Value: value, Message: message})
}

func (r *ImportResult) HasErrors() bool {
	return len(r.Errors) > 0
}
