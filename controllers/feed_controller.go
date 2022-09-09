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
	"encoding/json"
	"errors"
	"fmt"
	"time"

	skynewzdevv1alpha1 "github.com/SkYNewZ/putio-operator/api/v1alpha1"
	"github.com/SkYNewZ/putio-operator/internal/http"
	"github.com/SkYNewZ/putio-operator/internal/putio"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	corev1 "k8s.io/api/core/v1"
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
	finalizerAnnotation string = "feed.skynewz.dev/finalizer"
	checksumAnnotation  string = "feed.skynewz.dev/checksum"

	titleSuffix string = " (managed by Kubernetes/putio-operator)"
)

var tracer = otel.GetTracerProvider().Tracer("controller")

// FeedReconciler reconciles a Feed object.
type FeedReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=skynewz.dev,resources=feeds,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=skynewz.dev,resources=feeds/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=skynewz.dev,resources=feeds/finalizers,verbs=update
//+kubebuilder:rbac:groups=skynewz.dev,resources=events,verbs=create;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
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
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	r.Recorder.Event(k8sFeed, corev1.EventTypeNormal, eventReconciliationStarted, "starting reconciliation")

	logger.Info("Setting up put.io client with instance secret")
	clientAuthSecret := &corev1.Secret{}
	if err := r.Get(ctx, types.NamespacedName{Name: k8sFeed.AuthSecretRef().Name, Namespace: req.Namespace}, clientAuthSecret); err != nil {
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
				return ctrl.Result{}, err
			}
			r.Recorder.Event(k8sFeed, corev1.EventTypeNormal, eventAddedFinalizer, "instance finalizer added")
		}
	} else {
		// The object is being deleted
		if controllerutil.ContainsFinalizer(k8sFeed, finalizerAnnotation) {
			// our finalizer is present, so lets handle any external dependency
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
				return ctrl.Result{}, err
			}
		}

		// Stop reconciliation as the item is being deleted
		return ctrl.Result{}, nil
	}

	putioFeed, err := r.createOrUpdateFeed(ctx, k8sFeed, putioClient)
	if err != nil {
		r.Recorder.Event(k8sFeed, corev1.EventTypeWarning, eventUnableToCreateOrUpdatedAtPutio, err.Error())
		span.RecordError(err)
		return ctrl.Result{}, err
	}

	r.Recorder.Event(k8sFeed, corev1.EventTypeNormal, eventCreateOrUpdatedAtPutio, "feed successfully created or updated")

	// refresh resource after updates and update its status
	k8sFeed = new(skynewzdevv1alpha1.Feed)
	if err := r.Get(ctx, req.NamespacedName, k8sFeed); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if err := r.updateFeedStatus(ctx, k8sFeed, putioFeed); err != nil {
		r.Recorder.Event(k8sFeed, corev1.EventTypeWarning, eventUnableToUpdateFeedStatus, err.Error())
		span.RecordError(err)
		return ctrl.Result{}, err
	}
	r.Recorder.Event(k8sFeed, corev1.EventTypeNormal, eventFeedStatusSuccessfullyUpdated, "feed status updated successfully")

	logger.Info("Feed successfully reconciled")
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *FeedReconciler) SetupWithManager(mgr ctrl.Manager) error {
	_, span := tracer.Start(context.Background(), "controllers.FeedReconciler.SetupWithManager")
	defer span.End()

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
		return ctrl.Result{Requeue: false}, errors.New("cannot delete Feed without its ID")
	}

	span.SetAttributes(attribute.Int("feed.status.id", int(*feed.Status.ID)))
	return ctrl.Result{RequeueAfter: time.Minute}, putioClient.Rss.Delete(ctx, *feed.Status.ID)
}

func (r *FeedReconciler) createOrUpdateFeed(ctx context.Context, feed *skynewzdevv1alpha1.Feed, putioClient *putio.Client) (*putio.Feed, error) {
	ctx, span := tracer.Start(ctx, "controllers.FeedReconciler.createOrUpdateFeed")
	defer span.End()

	var (
		logger    = log.FromContext(ctx)
		putioFeed *putio.Feed
		err       error
	)

	// search for existing feed
	if feed.Status.ID != nil {
		span.SetAttributes(attribute.Int("feed.status.id", int(*feed.Status.ID)))
		logger.Info("Searching Put.io feed from status ID", "id", feed.Status.ID)
		putioFeed, err = putioClient.Rss.Get(ctx, *feed.Status.ID)
		if err != nil && !putio.IsNotFound(err) {
			return nil, fmt.Errorf("unable to read Put.io feed: %w", err)
		}
	}

	// feed not found, creating it
	if putioFeed == nil {
		span.SetAttributes(attribute.String("action", "create"))
		logger.Info("Put.io feed not found, creating it", "title", feed.Spec.Title)

		// TODO: refacto code for creation/update, this is the same
		title := feed.Spec.Title + titleSuffix
		putioFeed, err = putioClient.Rss.Create(ctx, &putio.Feed{
			Title:                title,
			RssSourceURL:         feed.Spec.RssSourceURL,
			ParentDirID:          feed.Spec.ParentDirID,
			DeleteOldFiles:       feed.Spec.DeleteOldFiles,
			DontProcessWholeFeed: feed.Spec.DontProcessWholeFeed,
			Keyword:              feed.Spec.Keyword,
			UnwantedKeywords:     feed.Spec.UnwantedKeywords,
		})

		if err != nil {
			return nil, fmt.Errorf("unable to create feed to Put.io: %w", err)
		}

		span.SetAttributes(attribute.Int("feed.id", int(*putioFeed.ID)))

		// update pause/unpause
		switch feed.Spec.Paused {
		case true:
			err = putioClient.Rss.Pause(ctx, *putioFeed.ID)
		case false:
			err = putioClient.Rss.Resume(ctx, *putioFeed.ID)
		}

		if err != nil {
			return nil, fmt.Errorf("unable to update pause status to Put.io: %w", err)
		}

		// update checksum and update resource
		r.insertChecksum(ctx, feed)
		if err := r.Update(ctx, feed); err != nil {
			return nil, fmt.Errorf("unable to update Feed resource: %w", err)
		}

		// get new values
		// TODO: is really necessary ?
		putioFeed, err = putioClient.Rss.Get(ctx, *putioFeed.ID)
		if err != nil {
			return nil, fmt.Errorf("unable to refresh feed details from Put.io: %w", err)
		}

		logger.Info("Put.io feed successfully created", "id", putioFeed.ID)
		return putioFeed, nil
	}

	// feed exist, update it if needed
	if r.feedNeedUpdate(ctx, feed) {
		span.SetAttributes(attribute.String("action", "update"))
		span.SetAttributes(attribute.Int("feed.id", int(*putioFeed.ID)))
		logger.Info("Put.io feed found, updating", "id", putioFeed.ID)

		title := feed.Spec.Title + titleSuffix
		if err := putioClient.Rss.Update(ctx, &putio.Feed{
			Title:                title,
			RssSourceURL:         feed.Spec.RssSourceURL,
			ParentDirID:          feed.Spec.ParentDirID,
			DeleteOldFiles:       feed.Spec.DeleteOldFiles,
			DontProcessWholeFeed: feed.Spec.DontProcessWholeFeed,
			Keyword:              feed.Spec.Keyword,
			UnwantedKeywords:     feed.Spec.UnwantedKeywords,
		}, *putioFeed.ID); err != nil {
			return nil, fmt.Errorf("unable to update feed to Put.io: %w", err)
		}

		// update pause/unpause
		switch feed.Spec.Paused {
		case true:
			err = putioClient.Rss.Pause(ctx, *putioFeed.ID)
		case false:
			err = putioClient.Rss.Resume(ctx, *putioFeed.ID)
		}

		if err != nil {
			return nil, fmt.Errorf("unable to update pause status to Put.io: %w", err)
		}

		// get new values
		// TODO: is really necessary ?
		putioFeed, err = putioClient.Rss.Get(ctx, *putioFeed.ID)
		if err != nil {
			return nil, fmt.Errorf("unable to refresh feed details from Put.io: %w", err)
		}

		// update checksum and update resource
		r.insertChecksum(ctx, feed)
		if err := r.Update(ctx, feed); err != nil {
			return nil, fmt.Errorf("unable to update Feed resource: %w", err)
		}

		return putioFeed, nil
	}

	logger.Info("Put.io feed found, no change on spec, nothing to do")
	return putioFeed, nil
}

func (r *FeedReconciler) updateFeedStatus(ctx context.Context, feed *skynewzdevv1alpha1.Feed, putioFeed *putio.Feed) error {
	ctx, span := tracer.Start(ctx, "controllers.FeedReconciler.updateFeedStatus")
	defer span.End()

	logger := log.FromContext(ctx)

	// update status
	logger.Info("Updating feed status")
	feed.Status = skynewzdevv1alpha1.FeedStatus{
		ID:              putioFeed.ID,
		LastError:       putioFeed.LastError,
		FailedItemCount: putioFeed.FailedItemCount,
	}

	if !putioFeed.LastFetch.IsZero() {
		t := metav1.NewTime(putioFeed.LastFetch.GetTime())
		feed.Status.LastFetch = &t
	}

	if !putioFeed.PausedAt.IsZero() {
		t := metav1.NewTime(putioFeed.PausedAt.GetTime())
		feed.Status.PausedAt = &t
	}

	if !putioFeed.CreatedAt.IsZero() {
		t := metav1.NewTime(putioFeed.CreatedAt.GetTime())
		feed.Status.CreatedAt = &t
	}

	if !putioFeed.UpdatedAt.IsZero() {
		t := metav1.NewTime(putioFeed.UpdatedAt.GetTime())
		feed.Status.UpdatedAt = &t
	}

	return r.Client.Status().Update(ctx, feed)
}

func (r *FeedReconciler) feedNeedUpdate(ctx context.Context, feed *skynewzdevv1alpha1.Feed) bool {
	_, span := tracer.Start(ctx, "controllers.FeedReconciler.feedNeedUpdate")
	defer span.End()

	data, _ := json.Marshal(feed.Spec) //nolint:errchkjson
	checksum := Checksum(string(data))
	return feed.Annotations[checksumAnnotation] != checksum
}

func (r *FeedReconciler) insertChecksum(ctx context.Context, feed *skynewzdevv1alpha1.Feed) {
	_, span := tracer.Start(ctx, "controllers.FeedReconciler.insertChecksum")
	defer span.End()

	logger := log.FromContext(ctx)
	logger.Info("Inserting checksum annotation")

	data, _ := json.Marshal(feed.Spec) //nolint:errchkjson
	checksum := Checksum(string(data))
	feed.Annotations[checksumAnnotation] = checksum
}

func (r *FeedReconciler) makePutioClient(ctx context.Context, token string) *putio.Client {
	httpClient := http.NewHTTPClient(token)
	return putio.New(ctx, httpClient)
}
