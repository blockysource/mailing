package mailprovider

// ErrAuth returns an authentication error.
func ErrAuth(err error) *AuthErr {
	return &AuthErr{
		Err: err,
	}
}

// AuthErr is an error that indicates a problem with the provider credentials / authentication.
type AuthErr struct {
	Err error
}

// Error returns the error.
func (e *AuthErr) Error() string {
	return e.Err.Error()
}

// Unwrap returns the error.
func (e *AuthErr) Unwrap() error {
	return e.Err
}

// Error is an error.
type Error struct {
	Temporary bool
	Err       error
}

// Error returns the error.
func (e *Error) Error() string {
	return e.Err.Error()
}

// Unwrap returns the error.
func (e *Error) Unwrap() error {
	return e.Err
}

// ErrTemporary returns a temporary error.
func ErrTemporary(err error) *Error {
	return &Error{
		Temporary: true,
		Err:       err,
	}
}

// ErrPermanent returns a permanent error.
func ErrPermanent(err error) *Error {
	return &Error{
		Temporary: false,
		Err:       err,
	}
}
