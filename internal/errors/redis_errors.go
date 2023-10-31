package errors

var (
	RedisInternalServerError = func(e error) *Error {
		return &Error{Type: typeInternalServer, Err: e, DisplayString: "Internal Server Error. Please try again later."}
	}
	RedisNotFoundError = func(e error) *Error {
		return &Error{Type: typeNotFound, Err: e, DisplayString: "Not Found"}
	}
	MissingUserIdAndTokenError = func() *Error {
		return &Error{Type: typeBadRequest, DisplayString: "Missing user_id and token"}
	}
)
