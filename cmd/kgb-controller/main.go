package main

import (
	"flag"
	"os"

	"github.com/akram/kgb/internal/controller"
	"github.com/akram/kgb/internal/scheme"
	"github.com/akram/kgb/pkg/approval"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

func main() {
	var metricsAddr string
	var probeAddr string
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "Prometheus metrics bind address.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "Health probe bind address.")
	opts := zap.Options{Development: true}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme.New(),
		Metrics:                metricsserver.Options{BindAddress: metricsAddr},
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         false,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	recorder := mgr.GetEventRecorderFor("kgb-controller")

	var nf approval.Notifier
	var chain approval.MultiNotifier
	if sl := approval.NewSlackFromEnv(); sl != nil {
		chain = append(chain, sl)
	}
	if cb := approval.NewRESTCallbackFromEnv(); cb != nil {
		chain = append(chain, cb)
	}
	if len(chain) > 0 {
		nf = chain
	}

	if err = (&controller.PolicyReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "policy controller")
		os.Exit(1)
	}
	if err = (&controller.ApprovalReconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Recorder: recorder,
		Notifier: nf,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "approval controller")
		os.Exit(1)
	}
	if err = (&controller.AgentSessionReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "session controller")
		os.Exit(1)
	}

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "healthz")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "readyz")
		os.Exit(1)
	}

	setupLog.Info("starting kgb-controller")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "manager exited")
		os.Exit(1)
	}
}

var setupLog = ctrl.Log.WithName("setup")
