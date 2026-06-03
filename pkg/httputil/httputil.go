// Package httputil standardizes JSON responses and maps domain errors to HTTP.
package httputil

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorResponse is the standard error envelope.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	Code    string `json:"code,omitempty"`
}

// DomainError is a typed error carrying an HTTP status and stable code.
type DomainError struct {
	Status  int
	Code    string
	Message string
}

func (e *DomainError) Error() string { return e.Message }

// NewDomainError constructs a DomainError.
func NewDomainError(status int, code, message string) *DomainError {
	return &DomainError{Status: status, Code: code, Message: message}
}

// Common domain error constructors.
func NotFound(msg string) *DomainError { return NewDomainError(http.StatusNotFound, "not_found", msg) }
func BadRequest(msg string) *DomainError {
	return NewDomainError(http.StatusBadRequest, "bad_request", msg)
}
func Unauthorized(msg string) *DomainError {
	return NewDomainError(http.StatusUnauthorized, "unauthorized", msg)
}
func Forbidden(msg string) *DomainError {
	return NewDomainError(http.StatusForbidden, "forbidden", msg)
}
func Conflict(msg string) *DomainError { return NewDomainError(http.StatusConflict, "conflict", msg) }

// PaymentRequired signals a plan quota/limit was hit (billing). Maps to 402.
func PaymentRequired(msg string) *DomainError {
	return NewDomainError(http.StatusPaymentRequired, "quota_exceeded", msg)
}

// OK writes a 200 with the payload.
func OK(c *gin.Context, payload any) { c.JSON(http.StatusOK, payload) }

// Created writes a 201 with the payload.
func Created(c *gin.Context, payload any) { c.JSON(http.StatusCreated, payload) }

// NoContent writes a 204.
func NoContent(c *gin.Context) { c.Status(http.StatusNoContent) }

// Fail writes an error response, mapping DomainError to its status. Unknown
// errors become 500 without leaking internals.
func Fail(c *gin.Context, err error) {
	var de *DomainError
	if errors.As(err, &de) {
		c.JSON(de.Status, ErrorResponse{Error: http.StatusText(de.Status), Message: de.Message, Code: de.Code})
		return
	}
	_ = c.Error(err) // record for logging middleware
	c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal_server_error", Message: "unexpected error"})
}

// BindJSON binds and validates the request body, writing a 400 on failure.
func BindJSON(c *gin.Context, dst any) bool {
	if err := c.ShouldBindJSON(dst); err != nil {
		Fail(c, BadRequest(err.Error()))
		return false
	}
	return true
}
