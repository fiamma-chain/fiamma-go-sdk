package ginctx

import (
	"net/http"
)

// Code code
type Code string

// all codes
const (
	// * request
	ErrRequestAccessDenied   Code = "ErrRequestAccessDenied"
	ErrRequestMethodNotFound      = "ErrRequestMethodNotFound"
	ErrRequestParamInvalid        = "ErrRequestParamInvalid"
	// * resource
	ErrResourceNotFound        = "ErrResourceNotFound"
	ErrResourceAccessForbidden = "ErrResourceAccessForbidden"
	ErrResourceConflict        = "ErrResourceConflict"
	ErrResourceDeleteForbidden = "ErrResourceDeleteForbidden"
	ErrResourceHasBeenUsed     = "ErrResourceHasBeenUsed"
	ErrInvalidToken            = "ErrInvalidToken"

	// * unknown
	ErrUnknown = "UnknownError"
)

func getHTTPStatus(c Code) int {
	switch c {
	case ErrResourceNotFound, ErrRequestMethodNotFound:
		return http.StatusNotFound
	case ErrRequestAccessDenied:
		return http.StatusUnauthorized
	case ErrResourceHasBeenUsed:
		return http.StatusForbidden
	case ErrUnknown:
		return http.StatusInternalServerError
	default:
		return http.StatusBadRequest
	}
}
