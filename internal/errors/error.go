package errors

import "net/http"

const (
	typeInternalServer = "internal_server"
	typeBadRequest     = "bad_request"
	typeNotFound       = "not_found"
	typeForbidden      = "forbidden"
)

type Error struct {
	Type          string
	Err           error
	DisplayString string
}

func (e *Error) Error() error {
	if e == nil {
		return nil
	}
	if e.Err != nil {
		return e.Err
	}
	return nil
}

func (e *Error) UserStatusError() int {
	if e.Type == typeInternalServer {
		return http.StatusInternalServerError
	}
	if e.Type == typeBadRequest {
		return http.StatusBadRequest
	}
	if e.Type == typeNotFound {
		return http.StatusNotFound
	}
	if e.Type == typeForbidden {
		return http.StatusForbidden
	}
	// Handle more error scenarios
	return -1
}

func (e *Error) UserErrorString() string {
	if e.DisplayString != "" {
		return e.DisplayString
	}
	return "Unknown error"
}

func (e *Error) IsNotFound() bool {
	return e.Type == typeNotFound
}
