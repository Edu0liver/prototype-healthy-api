package services

import "github.com/Edu0liver/prototype-healthy-api/internal/platform/httputil"

// ErrConversationNotFound is returned when a conversation is absent in scope.
var ErrConversationNotFound = httputil.NotFound("conversation not found")
