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

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	skynewzdevv1alpha1 "github.com/SkYNewZ/putio-operator/api/v1alpha1"
	"github.com/SkYNewZ/putio-operator/controllers"
	"github.com/SkYNewZ/putio-operator/internal/logger"
	"github.com/SkYNewZ/putio-operator/internal/sentry"
	"github.com/SkYNewZ/putio-operator/internal/tracing"
	"github.com/go-logr/zapr"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	//+kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")

	serviceVersion = "dev"
)

const (
	serviceName string = "putio-operator"
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(skynewzdevv1alpha1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

//nolint:cyclop
func main() {
	var (
		configFile string
		version    bool
	)

	flag.BoolVar(&version, "version", false, "Show current version")
	flag.StringVar(&configFile, "config", "",
		"The controller will load its initial configuration from this file. "+
			"Omit this flag to use the default configuration values. "+
			"Command-line flags override configuration from this file.")

	opts := zap.Options{Development: os.Getenv("DEBUG") == "1"}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	if version {
		fmt.Printf("%s %s\n", serviceName, serviceVersion) //nolint:forbidigo
		os.Exit(0)
	}

	// make the default logger for setup log
	l := zap.NewRaw(zap.UseFlagOptions(&opts))
	ctrl.SetLogger(zapr.NewLogger(l))

	setupLog.Info("configure sentry")
	sentryClient, err := sentry.ConfigureSentry(serviceName, serviceVersion)
	if err != nil {
		setupLog.Error(err, "unable to configure sentry")
		os.Exit(1)
	}

	// Flush buffered events before the program terminates.
	// Set the timeout to the maximum duration the program can afford to wait.
	defer sentryClient.Flush(time.Second * 2)

	setupLog.Info("configure logger")
	newLogger, err := logger.ConfigureLogger(sentryClient, l)
	if err != nil {
		setupLog.Error(err, "unable to configure loggerRaw")
		os.Exit(1)
	}

	ctrl.SetLogger(newLogger) // reset the final logger as default

	options := ctrl.Options{Scheme: scheme}
	if configFile != "" {
		options, err = options.AndFrom(ctrl.ConfigFile().AtPath(configFile))
		if err != nil {
			setupLog.Error(err, "unable to load the config file")
			os.Exit(1)
		}
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), options)
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&controllers.FeedReconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Recorder: mgr.GetEventRecorderFor("feed-reconciler"),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Feed")
		os.Exit(1)
	}
	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("configure tracer")
	tracerProvider, err := tracing.ConfigureTracing(context.Background(), serviceName, serviceVersion)
	if err != nil {
		setupLog.Error(err, "unable to setup tracer")
		os.Exit(1)
	}

	// Handle shutdown properly so nothing leaks.
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		setupLog.Info("Stopping and waiting for tracer")
		if err := tracerProvider.Shutdown(ctx); err != nil {
			setupLog.Error(err, "problem shutting down tracer provider")
		}
	}()

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
