// Copyright 2022 Authors of spidernet-io
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"context"
	"fmt"
	"github.com/spidernet-io/rocktemplate/pkg/ebpfWriter"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	discovery "k8s.io/api/discovery/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	k8szap "sigs.k8s.io/controller-runtime/pkg/log/zap"
)

type reconciler struct {
	// client can be used to retrieve objects from the APIServer.
	client client.Client
	log    *zap.Logger
}

func (r *reconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	t := reconcile.Result{}

	r.log.Sugar().Infof("Reconcile: %v", req)

	return t, nil
}

var scheme = runtime.NewScheme()

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
}

func SetupController() {
	logger := rootLogger.Named("controller")

	config := ctrl.GetConfigOrDie()
	config.Burst = 100
	config.QPS = 50
	mgr, err := ctrl.NewManager(config, manager.Options{
		Scheme:                 scheme,
		Metrics:                metricsserver.Options{BindAddress: "0"},
		HealthProbeBindAddress: "0",
	})
	if err != nil {
		logger.Sugar().Fatalf("unable to set up controller: %v ", err)
	}

	ctrl.SetLogger(k8szap.New())

	r := reconciler{
		client: mgr.GetClient(),
		log:    logger,
	}
	// Setup a new controller to reconcile ReplicaSets
	logger.Sugar().Info("Setting up controller")
	c, err := controller.New("agent", mgr, controller.Options{
		Reconciler: &r,
	})
	if err != nil {
		logger.Sugar().Fatalf("unable to set up individual controller: %v", err)
	}

	// Watch ReplicaSets and enqueue ReplicaSet object key
	if err := c.Watch(source.Kind(mgr.GetCache(), &corev1.Service{}, &handler.TypedEnqueueRequestForObject[*corev1.Service]{})); err != nil {
		logger.Sugar().Fatalf("unable to watch service: %v", err)
	}
	if err := c.Watch(source.Kind(mgr.GetCache(), &discovery.EndpointSlice{}, &handler.TypedEnqueueRequestForObject[*discovery.EndpointSlice]{})); err != nil {
		logger.Sugar().Fatalf("unable to watch EndpointSlice: %v", err)
	}

	logger.Sugar().Info("starting manager")
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		logger.Sugar().Fatalf("unable to run manager: %v", err)
	}

}

// ------------------------------

var (
	KubeConfigPath = filepath.Join(os.Getenv("HOME"), ".kube", "config")
	ScInPodPath    = "/var/run/secrets/kubernetes.io/serviceaccount"
)

func existFile(filePath string) bool {
	if info, err := os.Stat(filePath); err == nil {
		if !info.IsDir() {
			return true
		}
	}
	return false
}

func ExistDir(dirPath string) bool {
	if info, err := os.Stat(dirPath); err == nil {
		if info.IsDir() {
			return true
		}
	}
	return false
}

func autoConfig() (*rest.Config, error) {
	var config *rest.Config
	var err error

	if existFile(KubeConfigPath) == true {
		config, err = clientcmd.BuildConfigFromFlags("", KubeConfigPath)
		if err != nil {
			return nil, fmt.Errorf("failed to get config from kube config=%v , info=%v", KubeConfigPath, err)
		}

	} else if ExistDir(ScInPodPath) == true {
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to get config from serviceaccount=%v , info=%v", ScInPodPath, err)
		}

	} else {
		return nil, fmt.Errorf("failed to get config ")
	}

	return config, nil
}

func RunReconciles() {

	writer := ebpfWriter.NewEbpfWriter(rootLogger.Named("data cacher"))
	// get clientset
	c, e1 := autoConfig()
	if e1 != nil {
		rootLogger.Sugar().Fatalf("failed to find client-go config: %v", e1)
	}
	Client, e2 := kubernetes.NewForConfig(c)
	if e2 != nil {
		rootLogger.Sugar().Fatalf("failed to NewForConfig: %v", e2)
	}

	// setup service informer
	stopWatchCh := make(chan struct{})
	NewServiceInformer(Client, stopWatchCh, writer)
	NewEndpointSliceInformer(Client, stopWatchCh, writer)

}
