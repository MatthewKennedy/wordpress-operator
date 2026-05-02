/*
Copyright 2018 Pressinfra SRL.

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
	"os"

	flag "github.com/spf13/pflag"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	"github.com/bitpoke/wordpress-operator/pkg/apis"
	"github.com/bitpoke/wordpress-operator/pkg/cmd/options"
	"github.com/bitpoke/wordpress-operator/pkg/controller"
)

const genericErrorExitCode = 1

var setupLog = ctrl.Log.WithName("wordpress-operator")

// +kubebuilder:rbac:groups=coordination.k8s.io,resources=leases,verbs=get;list;watch;create;update;patch;delete

func main() {
	options.AddToFlagSet(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New())
	setupLog.Info("Starting wordpress-operator...")

	// Get a config to talk to the apiserver
	cfg := ctrl.GetConfigOrDie()

	opt := ctrl.Options{
		LeaderElection:          options.LeaderElection,
		LeaderElectionID:        options.LeaderElectionID,
		LeaderElectionNamespace: options.LeaderElectionNamespace,
		Metrics:                 metricsserver.Options{BindAddress: options.MetricsBindAddress},
		HealthProbeBindAddress:  options.HealthProbeBindAddress,
	}

	// Create a new Cmd to provide shared dependencies and start components
	mgr, err := ctrl.NewManager(cfg, opt)
	if err != nil {
		setupLog.Error(err, "unable to create a new manager")
		os.Exit(genericErrorExitCode)
	}

	// Setup Scheme for all resources
	if err := apis.AddToScheme(mgr.GetScheme()); err != nil {
		setupLog.Error(err, "unable to register types to scheme")
		os.Exit(genericErrorExitCode)
	}

	// Setup all Controllers
	if err := controller.AddToManager(mgr); err != nil {
		setupLog.Error(err, "unable to setup controllers")
		os.Exit(genericErrorExitCode)
	}

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}

	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	// Start the Cmd
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "unable to start the manager")
		os.Exit(genericErrorExitCode)
	}
}
