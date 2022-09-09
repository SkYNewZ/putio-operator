package controllers

import (
	"fmt"
	"hash/crc32"
	"regexp"
	"strings"
)

// Allows for normalizing by collapsing newlines.
var sequentialNewlines = regexp.MustCompile("(?:\r?\n)+")

const (
	eventReconciliationStarted string = "ReconciliationStarted"
	eventUnableToGetAuthSecret        = "UnableToGetAuthSecret"

	// finalizer events
	eventAddedFinalizer          = "InstanceFinalizerAdded"
	eventUnableToAddFinalizer    = "UnableToAddFinalizer"
	eventUnableToDeleteFinalizer = "UnableToDeleteFinalizer"

	// creation/update events
	eventCreateOrUpdatedAtPutio         = "CreateOrUpdatedAtPutio"
	eventUnableToCreateOrUpdatedAtPutio = "UnableToCreateOrUpdatedAtPutio"

	// deletion event
	eventUnableToDeleteAtPutio      = "UnableToDeleteAtPutio"
	eventSuccessfullyDeletedAtPutio = "SuccessfullyDeletedAtPutio"

	// status update
	eventUnableToUpdateFeedStatus      = "UnableToUpdateFeedStatus"
	eventFeedStatusSuccessfullyUpdated = "FeedStatusSuccessfullyUpdated"
)

// Checksum generates a checksum for the given value to be compared against
// a respective annotation.
// Leading and trailing spaces are ignored.
func Checksum(value string) string {
	return fmt.Sprintf("%08x", crc32.ChecksumIEEE([]byte(sequentialNewlines.ReplaceAllString(strings.TrimSpace(value), `\n`))))
}
