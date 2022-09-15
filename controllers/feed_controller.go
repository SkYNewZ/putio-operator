/*
Copyright 2022 Quentin Lemaire <quentin@lemairepro.fr>.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	skynewzdevv1alpha1 "github.com/SkYNewZ/putio-operator/api/v1alpha1"
	"github.com/SkYNewZ/putio-operator/internal/http"
	"github.com/SkYNewZ/putio-operator/internal/putio"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	// <wanted title>|generation|managed by Kubernetes/putio-operator.
	titleFormat    = "%s|%d|managed by Kubernetes/putio-operator"
	titleSeparator = "|"
)

var tracer = otel.GetTracerProvider().Tracer("controller")

var errCannotDeleteFeedWithoutID = errors.New("cannot delete Feed without its ID")

// FeedReconciler reconciles a Feed object.
type FeedReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=putio.skynewz.dev,resources=feeds,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=putio.skynewz.dev,resources=feeds/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=putio.skynewz.dev,resources=feeds/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
//nolint:nestif,cyclop
func (r *FeedReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	ctx, span := tracer.Start(ctx, "controllers.FeedReconciler.Reconcile")
	defer span.End()

	span.SetAttributes(
		attribute.String("feed.name", req.Name),
		attribute.String("feed.namespace", req.Namespace),
	)

	logger := log.FromContext(ctx)

	// get the feed from Kubernetes
	k8sFeed := new(skynewzdevv1alpha1.Feed)
	if err := r.Get(ctx, req.NamespacedName, k8sFeed); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err) //nolint:wrapcheck
	}

	r.Recorder.Event(k8sFeed, corev1.EventTypeNormal, eventReconciliationStarted, "starting reconciliation")

	logger.Info("Setting up put.io client with feed secret")
	clientAuthSecret := &corev1.Secret{}
	if err := r.Get(ctx, types.NamespacedName{Name: k8sFeed.AuthSecretRef().Name, Namespace: req.Namespace}, clientAuthSecret); err != nil {
		span.RecordError(err)
		r.Recorder.Eventf(k8sFeed, corev1.EventTypeWarning, eventUnableToGetAuthSecret, err.Error())
		return ctrl.Result{}, fmt.Errorf("cannot get secret %q: %w", k8sFeed.AuthSecretRef().Name, err)
	}
	putioClient := r.makePutioClient(ctx, string(clientAuthSecret.Data[k8sFeed.AuthSecretRef().Key]))

	// examine DeletionTimestamp to determine if object is under deletion
	if k8sFeed.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object. This is equivalent
		// registering our finalizer.
		if !controllerutil.ContainsFinalizer(k8sFeed, finalizerAnnotation) {
			controllerutil.AddFinalizer(k8sFeed, finalizerAnnotation)
			if err := r.Update(ctx, k8sFeed); err != nil {
				r.Recorder.Eventf(k8sFeed, corev1.EventTypeWarning, eventUnableToAddFinalizer, err.Error())
				span.RecordError(err)
				return ctrl.Result{}, err //nolint:wrapcheck
			}
			r.Recorder.Event(k8sFeed, corev1.EventTypeNormal, eventAddedFinalizer, "feed finalizer added")
		}
	} else {
		// The object is being deleted
		if controllerutil.ContainsFinalizer(k8sFeed, finalizerAnnotation) {
			// our finalizer is present, so lets handle any external dependency
			r.Recorder.Event(k8sFeed, corev1.EventTypeNormal, eventDeleteFeedAtPutio, "deleting feed at putio")
			result, err := r.deleteFeed(ctx, k8sFeed, putioClient)
			if err != nil {
				// if fail to delete the external dependency here, return with error
				// so that it can be retried
				r.Recorder.Event(k8sFeed, corev1.EventTypeWarning, eventUnableToDeleteAtPutio, err.Error())
				span.RecordError(err)
				return result, err
			}

			r.Recorder.Event(k8sFeed, corev1.EventTypeNormal, eventSuccessfullyDeletedAtPutio, "feed successfully deleted")

			// remove our finalizer from the list and update it.
			controllerutil.RemoveFinalizer(k8sFeed, finalizerAnnotation)
			if err := r.Update(ctx, k8sFeed); err != nil {
				r.Recorder.Event(k8sFeed, corev1.EventTypeWarning, eventUnableToDeleteFinalizer, err.Error())
				span.RecordError(err)
				return ctrl.Result{}, err //nolint:wrapcheck
			}
		}

		// Stop reconciliation as the item is being deleted
		return ctrl.Result{}, nil
	}

	r.Recorder.Event(k8sFeed, corev1.EventTypeNormal, eventCreateOrUpdatedAtPutio, "handling feed creation/update")
	putioFeed, err := r.createOrUpdateFeed(ctx, k8sFeed, putioClient)
	if err != nil {
		r.Recorder.Event(k8sFeed, corev1.EventTypeWarning, eventUnableToCreateOrUpdatedAtPutio, err.Error())
		return ctrl.Result{}, err
	}
	r.Recorder.Event(k8sFeed, corev1.EventTypeNormal, eventSuccessfullyCreateOrUpdatedAtPutio, "feed successfully created or updated")

	r.Recorder.Event(k8sFeed, corev1.EventTypeNormal, eventFeedStatus, "update feed status")
	if err := r.updateFeedStatus(ctx, k8sFeed, putioFeed); err != nil {
		r.Recorder.Event(k8sFeed, corev1.EventTypeWarning, eventUnableToUpdateFeedStatus, err.Error())
		return ctrl.Result{}, err
	}
	r.Recorder.Event(k8sFeed, corev1.EventTypeNormal, eventFeedStatusSuccessfullyUpdated, "feed status successfully set")

	logger.Info("Feed successfully reconciled")
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *FeedReconciler) SetupWithManager(mgr ctrl.Manager) error {
	_, span := tracer.Start(context.Background(), "controllers.FeedReconciler.SetupWithManager")
	defer span.End()

	//nolint:wrapcheck
	return ctrl.NewControllerManagedBy(mgr).
		For(&skynewzdevv1alpha1.Feed{}).
		Complete(r)
}

func (r *FeedReconciler) deleteFeed(ctx context.Context, feed *skynewzdevv1alpha1.Feed, putioClient *putio.Client) (ctrl.Result, error) {
	ctx, span := tracer.Start(ctx, "controllers.FeedReconciler.deleteFeed")
	defer span.End()

	span.SetAttributes(attribute.String("action", "delete"))

	logger := log.FromContext(ctx)
	logger.Info("Deleting feed")

	if feed.Status.ID == nil {
		return ctrl.Result{Requeue: false}, errCannotDeleteFeedWithoutID
	}

	span.SetAttributes(attribute.Int("feed.status.id", int(*feed.Status.ID)))

	if err := putioClient.Rss.Delete(ctx, *feed.Status.ID); err != nil {
		return ctrl.Result{RequeueAfter: time.Minute}, fmt.Errorf("failed to delete feed: %w", err)
	}

	return ctrl.Result{}, nil
}

func (r *FeedReconciler) createOrUpdateFeed(ctx context.Context, feed *skynewzdevv1alpha1.Feed, putioClient *putio.Client) (*putio.Feed, error) {
	ctx, span := tracer.Start(ctx, "controllers.FeedReconciler.createOrUpdateFeed")
	defer span.End()

	span.SetAttributes(attribute.Int64("feed.generation", feed.GetGeneration()))

	var (
		logger    = log.FromContext(ctx)
		putioFeed *putio.Feed
		err       error
	)

	// search for existing feed
	if feed.Status.ID != nil {
		span.SetAttributes(attribute.Int("feed.id", int(*feed.Status.ID)))
		logger.Info("Searching Put.io feed from status ID", "id", *feed.Status.ID)
		putioFeed, err = putioClient.Rss.Get(ctx, *feed.Status.ID)
		if err != nil && !putio.IsNotFound(err) {
			span.RecordError(err)
			return nil, fmt.Errorf("unable to read Put.io feed: %w", err)
		}
	}

	// feed not found, creating it
	if putioFeed == nil {
		span.SetAttributes(attribute.String("action", "create"))
		logger.Info("Put.io feed not found, creating it", "title", feed.Spec.Title)

		putioFeed, err = putioClient.Rss.Create(ctx, makePutioFeedFromSpec(ctx, feed))
		if err != nil {
			span.RecordError(err)
			return nil, fmt.Errorf("unable to create feed to Put.io: %w", err)
		}

		span.SetAttributes(attribute.Int("feed.id", int(*putioFeed.ID)))

		if err := r.setPauseStatus(ctx, putioClient, feed, *putioFeed.ID); err != nil {
			span.RecordError(err)
			return nil, fmt.Errorf("unable to update pause status to Put.io: %w", err)
		}

		logger.Info("Put.io feed successfully created", "id", putioFeed.ID)
		return putioFeed, nil
	}

	// feed found, updating it if not already at the latest version
	if !isAlreadyProcessed(ctx, putioFeed, feed) {
		span.SetAttributes(attribute.String("action", "update"))
		logger.Info("Put.io feed found, updating", "id", putioFeed.ID)

		if err := putioClient.Rss.Update(ctx, makePutioFeedFromSpec(ctx, feed), *putioFeed.ID); err != nil {
			span.RecordError(err)
			return nil, fmt.Errorf("unable to update feed to Put.io: %w", err)
		}

		if err := r.setPauseStatus(ctx, putioClient, feed, *putioFeed.ID); err != nil {
			span.RecordError(err)
			return nil, fmt.Errorf("unable to update pause status to Put.io: %w", err)
		}

		return putioFeed, nil
	}

	logger.Info("Feed up to date")
	return putioFeed, nil
}

func (r *FeedReconciler) updateFeedStatus(ctx context.Context, feed *skynewzdevv1alpha1.Feed, putioFeed *putio.Feed) error {
	ctx, span := tracer.Start(ctx, "controllers.FeedReconciler.updateFeedStatus")
	defer span.End()

	logger := log.FromContext(ctx)

	// update status
	logger.Info("Updating feed status")
	feed.Status.ID = putioFeed.ID

	if putioFeed.LastError == "" {
		meta.SetStatusCondition(&feed.Status.Conditions, makeFeedAvailableCondition(metav1.ConditionTrue, FeedSuccessfullyDeployed, ""))
	} else {
		meta.SetStatusCondition(&feed.Status.Conditions, makeFeedAvailableCondition(metav1.ConditionFalse, FeedFailedToDeploy, putioFeed.LastError))
	}

	return r.Client.Status().Update(ctx, feed) //nolint:wrapcheck
}

func (r *FeedReconciler) makePutioClient(ctx context.Context, token string) *putio.Client {
	ctx, span := tracer.Start(ctx, "controllers.FeedReconciler.makePutioClient")
	defer span.End()

	httpClient := http.NewHTTPClient(token)
	return putio.New(ctx, httpClient)
}

func (r *FeedReconciler) setPauseStatus(ctx context.Context, putioClient *putio.Client, feed *skynewzdevv1alpha1.Feed, feedID uint) error {
	ctx, span := tracer.Start(ctx, "controllers.FeedReconciler.setPauseStatus")
	defer span.End()

	r.Recorder.Event(feed, corev1.EventTypeNormal, eventSetPauseStatus, "setting feed pause status")
	var err error
	switch *feed.Spec.Paused {
	case true:
		err = putioClient.Rss.Pause(ctx, feedID)
	case false:
		err = putioClient.Rss.Resume(ctx, feedID)
	}

	if err != nil {
		r.Recorder.Event(feed, corev1.EventTypeWarning, eventUnableToSetPauseStatus, err.Error())
	} else {
		r.Recorder.Event(feed, corev1.EventTypeNormal, eventSuccessfullySetPauseStatus, "feed pause status set")
	}

	return err //nolint:wrapcheck
}

// makeFeedTitleWithGenerationNumber to prevent infinite reconciliation, make a checksum of current spec
// and place it into title.
func makeFeedTitleWithGenerationNumber(ctx context.Context, feed *skynewzdevv1alpha1.Feed) string {
	_, span := tracer.Start(ctx, "controllers.makeFeedTitleWithGenerationNumber")
	defer span.End()

	return fmt.Sprintf(titleFormat, feed.Spec.Title, feed.GetGeneration())
}

func isAlreadyProcessed(ctx context.Context, putioFeed *putio.Feed, feed *skynewzdevv1alpha1.Feed) bool {
	_, span := tracer.Start(ctx, "controllers.isAlreadyProcessed")
	defer span.End()

	// parse current title
	parsed := strings.Split(putioFeed.Title, titleSeparator)
	if len(parsed) != 3 {
		return false
	}

	return parsed[1] == strconv.FormatInt(feed.GetGeneration(), 10)
}

func makePutioFeedFromSpec(ctx context.Context, feed *skynewzdevv1alpha1.Feed) *putio.Feed {
	ctx, span := tracer.Start(ctx, "controllers.makePutioFeedFromSpec")
	defer span.End()

	return &putio.Feed{
		Title:                makeFeedTitleWithGenerationNumber(ctx, feed),
		RssSourceURL:         feed.Spec.RssSourceURL,
		ParentDirID:          *feed.Spec.ParentDirID,
		DeleteOldFiles:       *feed.Spec.DeleteOldFiles,
		DontProcessWholeFeed: *feed.Spec.DontProcessWholeFeed,
		Keyword:              feed.Spec.Keyword,
		UnwantedKeywords:     feed.Spec.UnwantedKeywords,
	}
}
