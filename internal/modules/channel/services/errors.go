package services

import "github.com/Edu0liver/prototype-healthy-api/internal/platform/httputil"

// Domain errors for the channel module.
var (
	ErrChannelNotFound = httputil.NotFound("channel not found")
	ErrUnsupportedType = httputil.BadRequest("unsupported channel type")
	ErrNotWhatsApp     = httputil.BadRequest("operation only valid for whatsapp channels")
)
