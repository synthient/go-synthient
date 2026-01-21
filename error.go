package synthient

import "errors"

var (
	ErrNoToken    = errors.New("no token provided for client")
	ErrFileExists = errors.New("file already exists")

	// http related errors
	ErrBadRequest           = errors.New("invalid input parameters")
	ErrUnauthorized         = errors.New("no api key was provided or the key is invalid")
	ErrPaymentRequired      = errors.New("credits have run out")
	ErrInternalServerError  = errors.New("unexpected error occurred")
	ErrUnexpectedStatusCode = errors.New("returned status code did not match expected status code")
)
