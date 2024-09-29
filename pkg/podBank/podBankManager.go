package podBank

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"strings"
)

type PidBankManager interface {
	Update(*corev1.Pod, *corev1.Pod)
	LookupPodByPid(uint32) (string, string, bool, error)
}

type podBankManager struct {
	client   *kubernetes.Clientset
	log      *zap.Logger
	NodeName string
	// key : podId and containerID
	// values: pod name and namespace
	podInfo *PodRegistry
}

var _ PidBankManager = (*podBankManager)(nil)

var PodBankHander PidBankManager

func InitPodBankManager(c *kubernetes.Clientset, log *zap.Logger, nodeName string) {
	if _, ok := PodBankHander.(*podBankManager); !ok {
		t := &podBankManager{
			client:   c,
			log:      log,
			NodeName: nodeName,
			// each node running pod with a total of max 1000
			podInfo: NewPodRegistry(1000),
		}
		t.initPodBank()
		PodBankHander = t
		log.Sugar().Info("finish initialize PodBankHander")
	} else {
		log.Sugar().Errorf("secondary calling for PodBankHander")
	}
}

// -----------------------------------

func (s *podBankManager) updatePodInfo(pod *corev1.Pod) error {
	getContaineridFunc := func(line string) string {
		// ContainerID is the ID of the container in the format '<type>://<container_id>'.
		index := strings.Index(line, "//")
		if index == -1 {
			return ""
		}
		return strings.TrimSpace(line[index+2:])
	}

	if len(pod.Status.ContainerStatuses) > 0 {
		l := len(pod.Status.ContainerStatuses)
		containerId := getContaineridFunc(pod.Status.ContainerStatuses[l-1].ContainerID)
		if len(containerId) < 0 {
			return fmt.Errorf("failed to get container id")
		}
		key := PodName{
			Podname:   pod.Name,
			Namespace: pod.Namespace,
		}
		value := PodID{
			PodUuid:     string(pod.ObjectMeta.UID),
			ContainerId: containerId,
		}
		s.podInfo.Set(key, value)
		return nil
	}
	return fmt.Errorf("no ContainerStatuses")
}

func (s *podBankManager) deletePodInfo(pod *corev1.Pod) {
	key := PodName{
		Podname:   pod.Name,
		Namespace: pod.Namespace,
	}
	s.podInfo.Delete(key)
}

// -----------------------------------

// before pod informer, build local database firstly for serving ebpf in case of missing event
func (s *podBankManager) initPodBank() {

	s.log.Sugar().Infof("initPodBank")

	namespaces, err := s.client.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		s.log.Sugar().Fatalf("failed to list namespaces: %v", err)
	}
	for _, ns := range namespaces.Items {
		// only get the pod of local node
		pods, err := s.client.CoreV1().Pods(ns.Name).List(context.TODO(), metav1.ListOptions{
			FieldSelector: fmt.Sprintf("spec.nodeName=%s", s.NodeName),
		})
		if err != nil {
			s.log.Sugar().Fatalf("Error listing pods in namespace %s: %v\n", ns.Name, err)
		}
		for _, pod := range pods.Items {
			s.log.Sugar().Debugf("save id information of pod %s/%s", pod.Namespace, pod.Name)
			if err := s.updatePodInfo(&pod); err != nil {
				s.log.Sugar().Errorf("error: %s", err)
			}

		}
	}

	s.log.Sugar().Infof("succeeded to get all pod uid information, total %d: %+v", s.podInfo.Count(), s.podInfo.GetAll())
	return
}

func (s *podBankManager) Update(oldPod, newPod *corev1.Pod) {
	if newPod == nil && oldPod == nil {
		return

	} else if newPod == nil && oldPod != nil {
		// delete
		s.log.Sugar().Debugf("delete pod id for pod %s/%s", oldPod.Namespace, oldPod.Name)
		s.deletePodInfo(oldPod)
		s.log.Sugar().Debugf("pod uid information, total %d: %+v", s.podInfo.Count(), s.podInfo.GetAll())
	} else {
		// add
		if newPod.Status.ContainerStatuses != nil && len(newPod.Status.ContainerStatuses) > 0 {
			s.log.Sugar().Debugf("add pod id for pod %s/%s", newPod.Namespace, newPod.Name)
			s.updatePodInfo(newPod)
			s.log.Sugar().Debugf("pod uid information, total %d: %+v", s.podInfo.Count(), s.podInfo.GetAll())
		}
	}
	return
}

// pid 用于查询 关联 pod name
func (s *podBankManager) LookupPodByPid(pid uint32) (podName, namespace string, host bool, err error) {
	if pid == 0 {
		return "", "", false, fmt.Errorf("empty input")
	}

	podId, containerId, host, e := getPodAndContainerID(pid)
	if e != nil {
		err = fmt.Errorf("failed to getPodAndContainerID for pid %d: %v ", pid, e)
		return
	}
	s.log.Sugar().Debugf("pod %d got: podUuid=%s, containerId=%s", pid, podId, containerId)

	// first, check lookup history
	value := PodID{
		PodUuid:     podId,
		ContainerId: containerId,
	}
	if k, ok := s.podInfo.GetKeyByValue(value); ok {
		return k.Podname, k.Namespace, false, nil
	}

	err = fmt.Errorf("no data of PodName for pid %d", pid)
	return
}
