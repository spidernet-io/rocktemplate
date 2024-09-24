package cmd

import (
	"github.com/spidernet-io/rocktemplate/pkg/ebpfWriter"
	"go.uber.org/zap"
	discoveryv1 "k8s.io/api/discovery/v1"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"reflect"
	"time"
)

// -----------------------------------
type EndpoingSliceReconciler struct {
	log    *zap.Logger
	writer ebpfWriter.EbpfWriter
}

func (s *EndpoingSliceReconciler) HandlerAdd(obj interface{}) {
	eds, ok := obj.(*discoveryv1.EndpointSlice)
	if !ok {
		s.log.Sugar().Warnf("HandlerAdd failed to get EndpointSlice obj: %v")
		return
	}
	name := eds.Namespace + "/" + eds.Name
	s.log.Sugar().Debugf("HandlerAdd process EndpointSlice: %+v", name)

	s.writer.UpdateEndpointSlice(eds)

	return
}

func (s *EndpoingSliceReconciler) HandlerUpdate(oldObj, newObj interface{}) {
	oldEds, ok1 := oldObj.(*discoveryv1.EndpointSlice)
	if !ok1 {
		s.log.Sugar().Warnf("HandlerUpdate failed to get old EndpointSlice obj: %v")
		return
	}
	newEds, ok2 := newObj.(*discoveryv1.EndpointSlice)
	if !ok2 {
		s.log.Sugar().Warnf("HandlerUpdate failed to get new EndpointSlice obj: %v")
		return
	}

	name := newEds.Namespace + "/" + newEds.Name
	if reflect.DeepEqual(oldEds.Endpoints, newEds.Endpoints) && reflect.DeepEqual(oldEds.Ports, newEds.Ports) {
		s.log.Sugar().Debugf("HandlerUpdate skip EndpointSlice: %+v", name)
		return
	}

	// s.log.Sugar().Debugf("HandlerUpdate get old EndpointSlice: %+v", oldEds)
	s.log.Sugar().Debugf("HandlerUpdate process EndpointSlice: %+v", newEds)
	s.writer.UpdateEndpointSlice(newEds)

	return
}

func (s *EndpoingSliceReconciler) HandlerDelete(obj interface{}) {
	eds, ok := obj.(*discoveryv1.EndpointSlice)
	if !ok {
		s.log.Sugar().Warnf("HandlerDelete failed to get EndpointSlice obj: %v")
		return
	}
	name := eds.Namespace + "/" + eds.Name
	s.log.Sugar().Debugf("HandlerDelete process EndpointSlice: %s", name)
	s.writer.DeleteEndpointSlice(eds)

	return
}

func NewEndpointSliceInformer(Client *kubernetes.Clientset, stopWatchCh chan struct{}, writer ebpfWriter.EbpfWriter) {

	// call HandlerUpdate at an interval of 60s
	kubeInformerFactory := kubeinformers.NewSharedInformerFactoryWithOptions(Client, time.Second*60)
	// service
	edsRes := discoveryv1.SchemeGroupVersion.WithResource("endpointslices")
	srcInformer, e3 := kubeInformerFactory.ForResource(edsRes)
	if e3 != nil {
		rootLogger.Sugar().Fatalf("failed to create service informer: %v", e3)
	}

	r := EndpoingSliceReconciler{
		log:    rootLogger.Named("EndpointSlice reconcile"),
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
		rootLogger.Sugar().Fatalf("failed to WaitForCacheSync for endpointslice ")
	}
	rootLogger.Sugar().Infof("succeeded to NewEndpointSliceInformer")

}
