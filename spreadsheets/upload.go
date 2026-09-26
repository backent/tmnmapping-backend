package spreadsheets

import (
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/malikabdulaziz/tmn-backend/exceptions"
	"github.com/malikabdulaziz/tmn-backend/helpers"
)

// maxUploadBytes caps a master-data upload at 32MB, matching the existing importers.
const maxUploadBytes = 32 << 20

// ReadUpload pulls the single "file" field off a multipart request and returns its
// bytes plus lower-cased extension.
func ReadUpload(r *http.Request) ([]byte, string) {
	if err := r.ParseMultipartForm(maxUploadBytes); err != nil {
		panic(exceptions.NewBadRequestError("Failed to parse upload. Max file size is 32MB."))
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		panic(exceptions.NewBadRequestError("File is required. Use form field 'file'."))
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	helpers.PanicIfError(err)

	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(header.Filename), "."))
	if ext != "xlsx" && ext != "csv" {
		panic(exceptions.NewBadRequestError("Unsupported file type. Use .xlsx or .csv files."))
	}

	return fileBytes, ext
}

// WriteXLSX sends a spreadsheet as a download.
func WriteXLSX(w http.ResponseWriter, filename string, fileBytes []byte) {
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(fileBytes)
}
