package putio

import (
	"errors"

	"github.com/putdotio/go-putio"
)

const (
	notFound string = "NotFound"
)

// IsNotFound check whether given error is a not found error.
func IsNotFound(err error) bool {
	var e *putio.ErrorResponse
	return errors.As(err, &e) && e.Type == notFound
}
