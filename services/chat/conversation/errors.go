package conversation

import "errors"

// The kinds of mistakes a caller can make. The api adapter turns each
// kind into the matching Connect error code.
var (
	// ErrNotFound: something the caller named doesn't exist.
	ErrNotFound = errors.New("not found")
	// ErrInvalidRequest: the request breaks a rule.
	ErrInvalidRequest = errors.New("invalid request")
)

// RequestError is a caller's mistake, with a message written for the
// person using the app, so it can be shown as it is.
type RequestError struct {
	// Kind is ErrNotFound or ErrInvalidRequest.
	Kind    error
	Message string
}

// Error returns the message. Every error type in Go has this method.
func (requestError *RequestError) Error() string {
	return requestError.Message
}

// Unwrap returns the kind, so errors.Is(err, ErrNotFound) works.
func (requestError *RequestError) Unwrap() error {
	return requestError.Kind
}

// notFound makes an ErrNotFound error with this message.
func notFound(message string) error {
	return &RequestError{Kind: ErrNotFound, Message: message}
}

// invalidRequest makes an ErrInvalidRequest error with this message.
func invalidRequest(message string) error {
	return &RequestError{Kind: ErrInvalidRequest, Message: message}
}
