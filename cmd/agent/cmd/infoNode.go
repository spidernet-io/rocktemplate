package cmd

import (
	"github.com/google/go-cmp/cmp"
	"github.com/spidernet-io/rocktemplate/pkg/ebpfWriter"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"reflect"
)

// -----------------------------------
type NodeReconciler struct {
	log    *zap.Logger
	writer ebpfWriter.EbpfWriter
}

func (s *NodeReconciler) HandlerAdd(obj interface{}) {
	node, ok := obj.(*corev1.Node)
	if !ok {
		s.log.Sugar().Warnf("HandlerAdd failed to get node obj: %v")
		return
	}
	logger := s.log.With(
		zap.String("node", node.Name),
	)

	logger.Sugar().Debugf("HandlerAdd process node %+v", node.Name)
	s.writer.UpdateNode(logger, node, false)

	return
}

func (s *NodeReconciler) HandlerUpdate(oldObj, newObj interface{}) {
	oldNode, ok1 := oldObj.(*corev1.Node)
	if !ok1 {
		s.log.Sugar().Warnf("HandlerUpdate failed to get old node obj %v")
		return
	}
	newNode, ok2 := newObj.(*corev1.Node)
	if !ok2 {
		s.log.Sugar().Warnf("HandlerUpdate failed to get new node obj %v")
		return
	}

	logger := s.log.With(
		zap.String("node", newNode.Name),
	)

	logger.Sugar().Debugf("HandlerUpdate process node %+v", newNode.Name)

	onlyUpdateTime := false
	if t := cmp.Diff(oldNode.Status.Addresses, newNode.Status.Addresses); len(t) > 0 {
		logger.Sugar().Debugf("node address: %s", t)
	}
	if reflect.DeepEqual(oldNode.Status.Addresses, newNode.Status.Addresses) {
		onlyUpdateTime = true
	}
	s.writer.UpdateNode(logger, newNode, onlyUpdateTime)

	return
}

func (s *NodeReconciler) HandlerDelete(obj interface{}) {
	node, ok := obj.(*corev1.Node)
	if !ok {
		s.log.Sugar().Warnf("HandlerDelete failed to get node obj: %v")
		return
	}
	logger := s.log.With(
		zap.String("node", node.Name),
	)

	logger.Sugar().Infof("HandlerDelete process node %+v", node.Name)
	s.writer.DeleteNode(logger, node)

	return
}

func NewNodeInformer(Client *kubernetes.Clientset, stopWatchCh chan struct{}, writer ebpfWriter.EbpfWriter) {

	// call HandlerUpdate at an interval of 60s
	kubeInformerFactory := kubeinformers.NewSharedInformerFactoryWithOptions(Client, InformerListInvterval)
	res := corev1.SchemeGroupVersion.WithResource("nodes")
	info, e3 := kubeInformerFactory.ForResource(res)
	if e3 != nil {
		rootLogger.Sugar().Fatalf("failed to create node informer %v", e3)
	}

	r := NodeReconciler{
		log:    rootLogger.Named("nodeReconciler"),
		writer: writer,
	}
	info.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    r.HandlerAdd,
		UpdateFunc: r.HandlerUpdate,
		DeleteFunc: r.HandlerDelete,
	})

	// notice that there is no need to run Start methods in a separate goroutine.
	// Start method is non-blocking and runs all registered informers in a dedicated goroutine.
	kubeInformerFactory.Start(stopWatchCh)

	if !cache.WaitForCacheSync(stopWatchCh, info.Informer().HasSynced) {
		rootLogger.Sugar().Fatalf("failed to WaitForCacheSync for node ")
	}

	rootLogger.Sugar().Infof("succeeded to NewNodeInformer ")
}
