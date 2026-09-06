package spreadsheets

import "errors"

// ErrUnsupportedFileType is returned for an upload that is neither xlsx nor csv.
var ErrUnsupportedFileType = errors.New("unsupported file type, use xlsx or csv")
