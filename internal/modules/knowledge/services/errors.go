package services

import (
	"errors"

	"github.com/Edu0liver/prototype-healthy-api/internal/platform/httputil"
)

// Domain errors for the knowledge module.
var (
	ErrKBNotFound       = httputil.NotFound("knowledge base not found")
	ErrDocumentNotFound = httputil.NotFound("document not found")
	// ErrUnsupportedFormat is internal (drives document.status=failed), not an HTTP error.
	ErrUnsupportedFormat = errors.New("knowledge: unsupported file format in v1 (txt/md/html supported)")
)
