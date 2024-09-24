package cmd

import (
	"github.com/spidernet-io/rocktemplate/pkg/ebpfWriter"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"reflect"
	"time"
)

// -----------------------------------
type ServiceReconciler struct {
	log    *zap.Logger
	writer ebpfWriter.EbpfWriter
}

func SkipServiceProcess(svc *corev1.Service) bool {
	switch svc.Spec.Type {
	case corev1.ServiceTypeClusterIP:
	case corev1.ServiceTypeNodePort:
	case corev1.ServiceTypeLoadBalancer:
		return false
	}
	return true
}

func (s *ServiceReconciler) HandlerAdd(obj interface{}) {
	svc, ok := obj.(*corev1.Service)
	if !ok {
		s.log.Sugar().Warnf("HandlerAdd failed to get sevice obj: %v")
		return
	}
	name := svc.Namespace + "/" + svc.Name
	if SkipServiceProcess(svc) {
		s.log.Sugar().Debugf("HandlerAdd skip sevice: %+v", name)
		return
	}

	s.log.Sugar().Debugf("HandlerAdd process sevice: %+v", name)
	s.writer.UpdateService(svc)

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

	name := newSvc.Namespace + "/" + newSvc.Name
	if SkipServiceProcess(newSvc) && SkipServiceProcess(oldSvc) {
		s.log.Sugar().Debugf("HandlerAdd skip service: %+v", name)
		return
	}
	if reflect.DeepEqual(oldSvc.Spec, newSvc.Spec) && reflect.DeepEqual(oldSvc.Status, newSvc.Status) {
		s.log.Sugar().Debugf("HandlerAdd skip service: %+v", name)
		return
	}

	s.log.Sugar().Debugf("HandlerUpdate process new sevice: %+v", name)
	s.writer.UpdateService(newSvc)

	return
}

func (s *ServiceReconciler) HandlerDelete(obj interface{}) {
	svc, ok := obj.(*corev1.Service)
	if !ok {
		s.log.Sugar().Warnf("HandlerDelete failed to get sevice obj: %v")
		return
	}
	name := svc.Namespace + "/" + svc.Name
	if SkipServiceProcess(svc) {
		s.log.Sugar().Debugf("HandlerAdd skip service: %+v", name)
		return
	}
	s.log.Sugar().Debugf("HandlerDelete process sevice: %+v", svc)
	s.writer.DeleteService(svc)

	return
}

func NewServiceInformer(Client *kubernetes.Clientset, stopWatchCh chan struct{}, writer ebpfWriter.EbpfWriter) {

	// call HandlerUpdate at an interval of 60s
	kubeInformerFactory := kubeinformers.NewSharedInformerFactoryWithOptions(Client, time.Second*60)
	// service
	svcRes := corev1.SchemeGroupVersion.WithResource("services")
	srcInformer, e3 := kubeInformerFactory.ForResource(svcRes)
	if e3 != nil {
		rootLogger.Sugar().Fatalf("failed to create service informer: %v", e3)
	}

	r := ServiceReconciler{
		log:    rootLogger.Named("service reconcile"),
		writer: writer,
	}
	srcInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    r.HandlerAdd,
		UpdateFunc: r.HandlerUpdate,
		DeleteFunc: r.HandlerDelete,
	})

	// notice that there is no need to run Start methods in a separate goroutine.
	// Start method is non-blocking and runs all registered informers in a dedicated goroutine.
	kubeInformerFactory.Start(stopWatchCh)

	if !cache.WaitForCacheSync(stopWatchCh, srcInformer.Informer().HasSynced) {
		rootLogger.Sugar().Fatalf("failed to WaitForCacheSync for serivce ")
	}

	rootLogger.Sugar().Infof("succeeded to NewServiceInformer ")
}
