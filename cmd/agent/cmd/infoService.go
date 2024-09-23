package cmd

import (
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"time"
)

// -----------------------------------
type ServiceReconciler struct {
	log *zap.Logger
}

func (s *ServiceReconciler) HandlerAdd(obj interface{}) {
	svc, ok := obj.(*corev1.Service)
	if !ok {
		s.log.Sugar().Warnf("HandlerAdd failed to get sevice obj: %v")
		return
	}
	s.log.Sugar().Infof("HandlerAdd get sevice: %+v", svc)
	return
}

func (s *ServiceReconciler) HandlerUpdate(oldObj, newObj interface{}) {
	oldSvc, ok1 := oldObj.(*corev1.Service)
	if !ok1 {
		s.log.Sugar().Warnf("HandlerUpdate failed to get old sevice obj: %v")
		return
	}
	newSvc, ok2 := newObj.(*corev1.Service)
	if !ok2 {
		s.log.Sugar().Warnf("HandlerUpdate failed to get new sevice obj: %v")
		return
	}

	s.log.Sugar().Infof("HandlerAdd get old sevice: %+v", oldSvc)
	s.log.Sugar().Infof("HandlerAdd get old sevice: %+v", newSvc)

	return
}

func (s *ServiceReconciler) HandlerDelete(obj interface{}) {
	svc, ok := obj.(*corev1.Service)
	if !ok {
		s.log.Sugar().Warnf("HandlerAdd failed to get sevice obj: %v")
		return
	}
	s.log.Sugar().Infof("HandlerDelete delete sevice: %+v", svc)
	return
}

func RunServiceInformer(Client kubernetes.Clientset) {

	kubeInformerFactory := kubeinformers.NewSharedInformerFactoryWithOptions(Client, time.Second*30)
	// service
	svcRes := corev1.SchemeGroupVersion.WithResource("services")
	srcInformer, e3 := kubeInformerFactory.ForResource(svcRes)
	if e3 != nil {
		rootLogger.Sugar().Fatalf("failed to create service informer: %v", e3)
	}

	r := ServiceReconciler{
		log: rootLogger.Named("service reconcile"),
	}
	srcInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    r.HandlerAdd,
		UpdateFunc: r.HandlerUpdate,
		DeleteFunc: r.HandlerDelete,
	})

	//
	// epsRes := discoveryv1.SchemeGroupVersion.WithResource("endpointslices")

}
