package putio

import (
	"errors"
	"fmt"

	"github.com/putdotio/go-putio"
)

const notFound string = "NotFound"

var _ error = (*InvalidStatusReceivedError)(nil)

// InvalidStatusReceivedError error from Putio when an update action is failed.
type InvalidStatusReceivedError struct {
	Status string
}

func newErrInvalidStatusReceived(status string) *InvalidStatusReceivedError {
	return &InvalidStatusReceivedError{Status: status}
}

func (e InvalidStatusReceivedError) Error() string {
	return fmt.Sprintf("invalid %q status received", e.Status)
}

// IsNotFound check whether given error is a not found error.
func IsNotFound(err error) bool {
	var e *putio.ErrorResponse
	return errors.As(err, &e) && e.Type == notFound
}
