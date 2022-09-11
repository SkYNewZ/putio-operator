package controllers

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const finalizerAnnotation string = "feed.skynewz.dev/finalizer"

const (
	eventReconciliationStarted string = "ReconciliationStarted"
	eventUnableToGetAuthSecret string = "UnableToGetAuthSecret" //nolint:gosec

	// finalizer events.
	eventAddedFinalizer          string = "InstanceFinalizerAdded"
	eventUnableToAddFinalizer    string = "UnableToAddFinalizer"
	eventUnableToDeleteFinalizer string = "UnableToDeleteFinalizer"

	// creation/update events.
	eventCreateOrUpdatedAtPutio             string = "CreateOrUpdatedAtPutio"
	eventUnableToCreateOrUpdatedAtPutio     string = "UnableToCreateOrUpdatedAtPutio"
	eventSuccessfullyCreateOrUpdatedAtPutio string = "SuccessfullyCreateOrUpdatedAtPutio"

	// deletion event.
	eventDeleteFeedAtPutio          string = "DeleteFeedAtPutio"
	eventUnableToDeleteAtPutio      string = "UnableToDeleteAtPutio"
	eventSuccessfullyDeletedAtPutio string = "SuccessfullyDeletedAtPutio"

	// pause status update.
	eventSetPauseStatus             string = "SetPauseStatus"
	eventUnableToSetPauseStatus     string = "UnableToSetPauseStatus"
	eventSuccessfullySetPauseStatus string = "SuccessfullySetPauseStatus"

	// status update.
	eventFeedStatus                    string = "FeedStatus"
	eventUnableToUpdateFeedStatus      string = "UnableToUpdateFeedStatus"
	eventFeedStatusSuccessfullyUpdated string = "FeedStatusSuccessfullyUpdated"
)

type FeedConditionType string

const (
	FeedAvailable FeedConditionType = "Available"
)

type FeedConditionReason string

const (
	FeedSuccessfullyDeployed FeedConditionReason = "FeedSuccessfullyDeployed"
	FeedFailedToDeploy       FeedConditionReason = "FeedFailedToDeploy"
)

func makeFeedAvailableCondition(status metav1.ConditionStatus, reason FeedConditionReason, message string) metav1.Condition {
	return metav1.Condition{
		Type:    string(FeedAvailable),
		Status:  status,
		Reason:  string(reason),
		Message: message,
	}
}
