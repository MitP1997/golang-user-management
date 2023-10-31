package errors

var (
	SendMailError = func(e error) *Error { return &Error{Type: typeInternalServer, Err: e, DisplayString: e.Error()} }
)
