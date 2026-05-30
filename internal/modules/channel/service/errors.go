package service

import "github.com/Edu0liver/prototype-healthy-api/pkg/httputil"

// Domain errors for the channel module.
var (
	ErrChannelNotFound = httputil.NotFound("channel not found")
	ErrUnsupportedType = httputil.BadRequest("unsupported channel type")
	ErrNotWhatsApp     = httputil.BadRequest("operation only valid for whatsapp channels")
)
