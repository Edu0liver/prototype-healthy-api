package service

import "github.com/Edu0liver/prototype-healthy-api/pkg/httputil"

// Domain errors for the automation module.
var (
	ErrAutomationNotFound = httputil.NotFound("automation not found")
	ErrChannelNotFound    = httputil.BadRequest("channel not found in tenant")
	ErrAgentNotFound      = httputil.BadRequest("agent not found in tenant")
	// ErrActiveExists maps the "one active automation per channel" invariant.
	ErrActiveExists = httputil.Conflict("an active automation already exists for this channel; deactivate it first")
)
