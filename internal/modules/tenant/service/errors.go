package service

import "github.com/Edu0liver/prototype-healthy-api/pkg/httputil"

// Domain errors for the tenant module, mapped to HTTP by httputil.Fail.
var (
	ErrCompanyNotFound  = httputil.NotFound("company not found")
	ErrBrandingNotFound = httputil.NotFound("branding not found")
	ErrSlugTaken        = httputil.Conflict("slug already in use")
	ErrDomainTaken      = httputil.Conflict("domain already registered")
)
