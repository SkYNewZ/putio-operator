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

package v1alpha1

import (
	"context"
	"net/url"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

var (
	// log is for logging in this package.
	feedlog = logf.Log.WithName("feed-resource")
	tracer  = otel.GetTracerProvider().Tracer("webhook")
)

const defaultParentDirID uint = 0

func (r *Feed) SetupWebhookWithManager(mgr ctrl.Manager) error {
	_, span := tracer.Start(context.Background(), "v1alpha1.Feed.SetupWebhookWithManager")
	defer span.End()

	//nolint:wrapcheck
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-putio-skynewz-dev-v1alpha1-feed,mutating=true,failurePolicy=fail,sideEffects=None,groups=putio.skynewz.dev,resources=feeds,verbs=create;update,versions=v1alpha1,name=mfeed.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &Feed{}

// Default implements webhook.Defaulter so a webhook will be registered for the type.
func (r *Feed) Default() {
	_, span := tracer.Start(context.Background(), "v1alpha1.Feed.Default")
	defer span.End()

	span.SetAttributes(attribute.String("name", r.Name))
	feedlog.Info("default", "name", r.Name)

	if r.Spec.ParentDirID == nil {
		r.Spec.ParentDirID = new(uint)
		*r.Spec.ParentDirID = defaultParentDirID
	}

	if r.Spec.DeleteOldFiles == nil {
		r.Spec.DeleteOldFiles = new(bool)
	}

	if r.Spec.DontProcessWholeFeed == nil {
		r.Spec.DontProcessWholeFeed = new(bool)
	}

	if r.Spec.Paused == nil {
		r.Spec.Paused = new(bool)
	}
}

//+kubebuilder:webhook:path=/validate-putio-skynewz-dev-v1alpha1-feed,mutating=false,failurePolicy=fail,sideEffects=None,groups=putio.skynewz.dev,resources=feeds,verbs=create;update,versions=v1alpha1,name=vfeed.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Feed{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type.
func (r *Feed) ValidateCreate() error {
	_, span := tracer.Start(context.Background(), "v1alpha1.Feed.ValidateCreate")
	defer span.End()

	span.SetAttributes(attribute.String("name", r.Name))
	feedlog.Info("validate create", "name", r.Name)
	return r.validateFeedSpec()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type.
func (r *Feed) ValidateUpdate(_ runtime.Object) error {
	_, span := tracer.Start(context.Background(), "v1alpha1.Feed.ValidateUpdate")
	defer span.End()

	span.SetAttributes(attribute.String("name", r.Name))
	feedlog.Info("validate update", "name", r.Name)
	return r.validateFeedSpec()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type.
func (r *Feed) ValidateDelete() error {
	_, span := tracer.Start(context.Background(), "v1alpha1.Feed.ValidateDelete")
	defer span.End()

	span.SetAttributes(attribute.String("name", r.Name))
	feedlog.Info("validate delete", "name", r.Name)
	return nil // nothing to validate on deletion
}

func (r *Feed) validateFeedSpec() error {
	_, span := tracer.Start(context.Background(), "v1alpha1.Feed.validateFeedSpec")
	defer span.End()

	// validate URL
	return r.validateRSSSourceURL(r.Spec.RssSourceURL, field.NewPath("spec").Child("rss_source_url"))
}

func (r *Feed) validateRSSSourceURL(u string, fldPath *field.Path) error {
	if _, err := url.ParseRequestURI(u); err != nil {
		return field.Invalid(fldPath, u, "invalid URL provided")
	}

	return nil
}
