/*

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
	"flag"
	installv1alpha1 "github.com/fyuan1316/test-cluster/api/v1alpha1"
	"github.com/fyuan1316/test-cluster/controllers"
	"github.com/fyuan1316/test-cluster/webhooks"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)

	_ = installv1alpha1.AddToScheme(scheme)
	_ = installv1alpha1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	flag.Parse()

	ctrl.SetLogger(zap.New(func(o *zap.Options) {
		o.Development = true
	}))
	hookServerCertDir := GetEnv("CertDir", "/tmp/k8s-webhook-server/serving-certs") //"/Users/max/cert"
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		LeaderElection:     enableLeaderElection,
		Port:               9443,
		CertDir:            hookServerCertDir,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&controllers.TiMatrixReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("TiMatrix"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "TiMatrix")
		os.Exit(1)
	}
	//if err = (&installv1alpha1.TiMatrix{}).SetupWebhookWithManager(mgr); err != nil {
	//	setupLog.Error(err, "unable to create webhook", "webhook", "TiMatrix")
	//	os.Exit(1)
	//}
	// +kubebuilder:scaffold:builder

	//hookServer := &webhook.Server{
	//	Port:    9443,
	//	CertDir: hookServerCertDir,
	//}
	//if err := mgr.Add(hookServer); err != nil {
	//	setupLog.Error(err, "unable to register webhook server with manager.")
	//	os.Exit(1)
	//}
	//mPath := "/mutate-timatrix"
	//vPath := "/validate-timatrix"
	//mgr.GetWebhookServer().Register(mPath, &webhook.Admission{Handler: &installv1alpha1.TiMatrixMutate{}})
	//mgr.GetWebhookServer().Register(vPath, &webhook.Admission{Handler: &installv1alpha1.TiMatrixValidate{
	//	Client: mgr.GetClient(),
	//}})

	webhooks.Register(mgr)

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

func GetEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}
