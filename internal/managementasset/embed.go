package managementasset

import (
	_ "embed"
	"io"
	"net/http"
)

//go:embed embedded/management.html
var managementHTML []byte

// GetEmbeddedManagementHTML returns the embedded management.html content.
// This allows serving the management UI without runtime downloads.
func GetEmbeddedManagementHTML() []byte {
	return managementHTML
}

// ServeEmbeddedManagementHTML writes the embedded management.html to the response writer.
func ServeEmbeddedManagementHTML(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Content-Length", string(len(managementHTML)))
	w.WriteHeader(http.StatusOK)
	_, err := w.Write(managementHTML)
	return err
}

// ReadEmbeddedManagementHTML returns an io.Reader for the embedded management.html.
func ReadEmbeddedManagementHTML() io.Reader {
	return io.NewSectionReader(&embedReader{data: managementHTML}, 0, int64(len(managementHTML)))
}

// embedReader implements io.ReaderAt for the embedded data.
type embedReader struct {
	data []byte
}

func (r *embedReader) ReadAt(p []byte, off int64) (n int, err error) {
	if off >= int64(len(r.data)) {
		return 0, io.EOF
	}
	if off < 0 {
		return 0, io.ErrUnexpectedEOF
	}
	n = copy(p, r.data[off:])
	if n < len(p) {
		err = io.EOF
	}
	return
}
