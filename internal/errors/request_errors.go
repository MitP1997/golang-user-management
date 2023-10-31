package errors

var (
	BadRequestError   = func(str string) *Error { return &Error{Type: typeBadRequest, Err: nil, DisplayString: str} }
	UnauthedUserError = func(e error) *Error {
		return &Error{Type: typeForbidden, Err: e, DisplayString: "Unauthenticated user"}
	}
	InvalidOtpError = func(e error) *Error {
		return &Error{Type: typeBadRequest, Err: e, DisplayString: "Invalid OTP"}
	}
)
