package cmd

import (
	"fmt"
	"github.com/spidernet-io/rocktemplate/pkg/podBank"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

// -----------------------------------
type PodReconciler struct {
	log *zap.Logger
}

func (s *PodReconciler) HandlerAdd(obj interface{}) {
	pod, ok := obj.(*corev1.Pod)
	if !ok {
		s.log.Sugar().Warnf("HandlerAdd failed to get pod obj: %v")
		return
	}
	name := pod.Namespace + "/" + pod.Name
	logger := s.log.With(
		zap.String("pod", name),
	)
	logger.Sugar().Debugf("HandlerAdd process node %+v", name)

	s.log.Sugar().Info("update pod ip for pod %s", name)
	podBank.PodBankHander.Update(nil, pod)

	return
}

func (s *PodReconciler) HandlerUpdate(oldObj, newObj interface{}) {
	oldPod, ok1 := oldObj.(*corev1.Pod)
	if !ok1 {
		s.log.Sugar().Warnf("HandlerUpdate failed to get old pod obj %v")
		return
	}
	newPod, ok2 := newObj.(*corev1.Pod)
	if !ok2 {
		s.log.Sugar().Warnf("HandlerUpdate failed to get new pod obj %v")
		return
	}
	name := newPod.Namespace + "/" + newPod.Name
	logger := s.log.With(
		zap.String("pod", name),
	)

	s.log.Sugar().Info("update pod ip for pod %s/%s", newPod.Namespace, newPod.Name)
	podBank.PodBankHander.Update(oldPod, newPod)

	logger.Sugar().Debugf("HandlerUpdate process pod %s", name)

	return
}

func (s *PodReconciler) HandlerDelete(obj interface{}) {
	pod, ok := obj.(*corev1.Pod)
	if !ok {
		s.log.Sugar().Warnf("HandlerDelete failed to get pod obj: %v")
		return
	}
	name := pod.Namespace + "/" + pod.Name
	logger := s.log.With(
		zap.String("pod", name),
	)

	logger.Sugar().Infof("HandlerDelete process pod %s", name)
	podBank.PodBankHander.Update(pod, nil)

	return
}

func NewPodInformer(Client *kubernetes.Clientset, stopWatchCh chan struct{}, localNodeName string) {

	// call HandlerUpdate at an interval of 60s
	kubeInformerFactory := kubeinformers.NewSharedInformerFactoryWithOptions(Client, InformerListInvterval, kubeinformers.WithTweakListOptions(func(options *metav1.ListOptions) {
		options.FieldSelector = fmt.Sprintf("spec.nodeName=%s", localNodeName)
	}))
	res := corev1.SchemeGroupVersion.WithResource("pods")
	info, e3 := kubeInformerFactory.ForResource(res)
	if e3 != nil {
		rootLogger.Sugar().Fatalf("failed to create pod informer %v", e3)
	}

	r := PodReconciler{
		log: rootLogger.Named("PodReconciler"),
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
		rootLogger.Sugar().Fatalf("failed to WaitForCacheSync for pod ")
	}

	rootLogger.Sugar().Infof("succeeded to NewPodInformer ")
}
