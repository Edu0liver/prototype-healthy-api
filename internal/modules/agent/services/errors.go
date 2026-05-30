package services

import "github.com/Edu0liver/prototype-healthy-api/internal/platform/httputil"

// ErrAgentNotFound is returned when an agent is not found in the tenant scope.
var ErrAgentNotFound = httputil.NotFound("agent not found")
