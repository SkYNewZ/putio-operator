package controllers

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const finalizerAnnotation string = "feed.skynewz.dev/finalizer"

const (
	eventReconciliationStarted string = "ReconciliationStarted"
	eventUnableToGetAuthSecret        = "UnableToGetAuthSecret"

	// finalizer events
	eventAddedFinalizer          = "InstanceFinalizerAdded"
	eventUnableToAddFinalizer    = "UnableToAddFinalizer"
	eventUnableToDeleteFinalizer = "UnableToDeleteFinalizer"

	// creation/update events
	eventCreateOrUpdatedAtPutio             = "CreateOrUpdatedAtPutio"
	eventUnableToCreateOrUpdatedAtPutio     = "UnableToCreateOrUpdatedAtPutio"
	eventSuccessfullyCreateOrUpdatedAtPutio = "SuccessfullyCreateOrUpdatedAtPutio"

	// deletion event
	eventDeleteFeedAtPutio          = "DeleteFeedAtPutio"
	eventUnableToDeleteAtPutio      = "UnableToDeleteAtPutio"
	eventSuccessfullyDeletedAtPutio = "SuccessfullyDeletedAtPutio"

	// pause status update
	eventSetPauseStatus             = "SetPauseStatus"
	eventUnableToSetPauseStatus     = "UnableToSetPauseStatus"
	eventSuccessfullySetPauseStatus = "SuccessfullySetPauseStatus"

	// status update
	eventFeedStatus                    = "FeedStatus"
	eventUnableToUpdateFeedStatus      = "UnableToUpdateFeedStatus"
	eventFeedStatusSuccessfullyUpdated = "FeedStatusSuccessfullyUpdated"
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
