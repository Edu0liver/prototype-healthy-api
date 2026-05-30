package service

import "github.com/Edu0liver/prototype-healthy-api/pkg/httputil"

// Domain errors for the iam module.
var (
	ErrInvalidCredentials = httputil.Unauthorized("invalid credentials")
	ErrUserNotFound       = httputil.NotFound("user not found")
	ErrUserDisabled       = httputil.Forbidden("user disabled")
	ErrEmailTaken         = httputil.Conflict("email already in use")
	ErrInvalidInvite      = httputil.BadRequest("invalid or expired invite")
	ErrCompanyNotFound    = httputil.NotFound("company not found")
	ErrAdminExists        = httputil.Conflict("company already has users")
)
