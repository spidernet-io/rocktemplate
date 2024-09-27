package podBank

import (
	"context"
	"fmt"
	"github.com/spidernet-io/rocktemplate/pkg/lock"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"net"
)

type PidBankManager interface {
	Update(*corev1.Pod, *corev1.Pod)
	LookupPodByPid(uint32, uint32) (*PodInfo, error)
}

type PodInfo struct {
	PodName   string
	NameSpace string
}

type podBankManager struct {
	client       *kubernetes.Clientset
	log          *zap.Logger
	NodeName     string
	LookupCacher *LimitedStore
	dataLock     *lock.RWMutex
	// store the pod and its ip .
	// key the ip
	podIpDatabase map[string]PodInfo
}

var _ PidBankManager = (*podBankManager)(nil)

var PodBankHander PidBankManager

// save all ip of local non-hostwork pod on this node
func InitPodBankManager(c *kubernetes.Clientset, log *zap.Logger, nodeName string) {
	if _, ok := PodBankHander.(*podBankManager); !ok {
		t := &podBankManager{
			client:        c,
			podIpDatabase: make(map[string]PodInfo),
			dataLock:      &lock.RWMutex{},
			log:           log,
			NodeName:      nodeName,
			LookupCacher:  NewLimitedStore(1000),
		}
		t.initPodBank()
		PodBankHander = t
		log.Sugar().Info("finish initialize PodBankHander")
	} else {
		log.Sugar().Errorf("secondary calling for PodBankHander")
	}
}

// -----------------------------------

func (s *podBankManager) updatePodInfo(pod *corev1.Pod) {
	s.dataLock.Lock()
	defer s.dataLock.Unlock()
	for _, v := range pod.Status.PodIPs {
		t := net.ParseIP(v.IP)
		if t == nil {
			continue
		}
		// unify the ip string in case of ipv6
		s.podIpDatabase[t.String()] = PodInfo{
			PodName:   pod.Name,
			NameSpace: pod.Namespace,
		}
	}
}

func (s *podBankManager) deletePodInfo(pod *corev1.Pod) {
	s.dataLock.Lock()
	defer s.dataLock.Unlock()
	for _, v := range pod.Status.PodIPs {
		t := net.ParseIP(v.IP)
		if t == nil {
			continue
		}
		delete(s.podIpDatabase, t.String())
	}
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
		pods, err := s.client.CoreV1().Pods(ns.Name).List(context.TODO(), metav1.ListOptions{
			FieldSelector: fmt.Sprintf("spec.nodeName=%s", s.NodeName),
		})
		if err != nil {
			s.log.Sugar().Fatalf("Error listing pods in namespace %s: %v\n", ns.Name, err)
		}
		for _, pod := range pods.Items {
			s.log.Sugar().Debugf("save ip information of pod %s/%s", pod.Namespace, pod.Name)
			s.updatePodInfo(&pod)
		}
	}

	s.log.Sugar().Infof("succeeded to get all pod IP information: %+v", s.podIpDatabase)
	return
}

// statefulset pod always has same pod name
func (s *podBankManager) Update(oldPod, newPod *corev1.Pod) {
	if oldPod == nil && newPod == nil {
		return
	} else if newPod == nil && oldPod != nil {
		// delete
		if len(oldPod.Status.PodIPs) > 0 && !oldPod.Spec.HostNetwork {
			s.log.Sugar().Debugf("delete podInfor for pod %s/%s", oldPod.Namespace, oldPod.Name)
			s.deletePodInfo(oldPod)
		} else {
			s.log.Sugar().Debugf("ignore deleting podInfor for pod %s/%s", oldPod.Namespace, oldPod.Name)
		}
	} else if newPod != nil && oldPod == nil {
		// add
		if len(newPod.Status.PodIPs) > 0 && !newPod.Spec.HostNetwork {
			s.log.Sugar().Debugf("add podInfor for pod %s/%s", newPod.Namespace, newPod.Name)
			s.updatePodInfo(newPod)
		} else {
			s.log.Sugar().Debugf("ignore adding podInfor for pod %s/%s", newPod.Namespace, newPod.Name)
		}
	} else {
		// update
		// only for statefulset, they will use same podname with possibly different ip address
		// delete the old one first
		if len(oldPod.Status.PodIPs) > 0 && !oldPod.Spec.HostNetwork {
			s.log.Sugar().Debugf("update old podInfor for pod %s/%s", newPod.Namespace, newPod.Name)
			s.deletePodInfo(oldPod)
		} else {
			s.log.Sugar().Debugf("ignore updating old podInfor for pod %s/%s", newPod.Namespace, newPod.Name)
		}
		// add the new one
		if len(newPod.Status.PodIPs) > 0 && !newPod.Spec.HostNetwork {
			s.log.Sugar().Debugf("update new podInfor for pod %s/%s", newPod.Namespace, newPod.Name)
			s.updatePodInfo(newPod)
		} else {
			s.log.Sugar().Debugf("ignore updating new podInfor for pod %s/%s", newPod.Namespace, newPod.Name)
		}
	}
	return
}

// pid 用于查询 ip ， 关联 pod
// 返回为空，可能（1）没记录   ； 肯（2）是 hostnetwork pod 或者 宿主机上的进程
// cgrouid 只是为了多一个变量，形成 pid 之间的绑定，避免 pod 销毁后，其它 pod 的 进程 使用了 相同的 pid，在查询 LookupCacher 时得出了错误的记录
func (s *podBankManager) LookupPodByPid(cgrouid, pid uint32) (*PodInfo, error) {
	if pid == 0 || cgrouid == 0 {
		return nil, fmt.Errorf("empty input")
	}

	// check whether the pid belongs to host or host-network pod
	if host, err := CheckHostNetNamespaceByPid(int(pid)); err != nil {
		s.log.Sugar().Errorf("failed to CheckHostNetNamespaceByPid: %v", err)
	} else {
		if host {
			s.log.Sugar().Warnf("failed to find podName for a pid %d running the hostnetowrk", pid)
			// the pid runs in the hostnetwork
			return nil, nil
		}
	}

	// first, check lookup history
	key := Key{
		Pid:    pid,
		Cgroup: cgrouid,
	}
	if value, ok := s.LookupCacher.Get(key); ok {
		return &PodInfo{
			PodName:   value.Podname,
			NameSpace: value.Namespace,
		}, nil
	}

	// then, check step by step
	IPList, err := GetContainerIP(int(pid), []string{"eth0"})
	if err != nil {
		t := fmt.Sprintf("failed to get container ip for pid %d: %v", pid, err)
		s.log.Sugar().Errorf("%s", t)
		return nil, fmt.Errorf("%s", s)
	}
	if len(IPList.IPv4) == 0 && len(IPList.IPv4) == 0 {
		t := fmt.Sprintf("no container ip for pid %d", pid)
		s.log.Sugar().Errorf("%s", t)
		return nil, fmt.Errorf("%s", s)
	}

	s.dataLock.RLock()
	defer s.dataLock.RUnlock()
	for _, addr := range IPList.IPv4 {
		ipStr := addr.String()
		if podInfo, ok := s.podIpDatabase[ipStr]; ok {
			// cache the feat
			key := Key{
				Pid:    pid,
				Cgroup: cgrouid,
			}
			value := Value{
				Podname:   podInfo.PodName,
				Namespace: podInfo.NameSpace,
			}
			s.LookupCacher.Set(key, value)
			return &podInfo, nil
		}
	}
	for _, addr := range IPList.IPv6 {
		ipStr := addr.String()
		if podInfo, ok := s.podIpDatabase[ipStr]; ok {
			// cache the feat
			key := Key{
				Pid:    pid,
				Cgroup: cgrouid,
			}
			value := Value{
				Podname:   podInfo.PodName,
				Namespace: podInfo.NameSpace,
			}
			s.LookupCacher.Set(key, value)
			return &podInfo, nil
		}
	}

	s.log.Sugar().Errorf("no data for PodName for pid %d ", pid)
	s.log.Sugar().Errorf("podIpDatabase: %+v", s.podIpDatabase)
	s.log.Sugar().Errorf("local ip of the pid: %+v", IPList)

	return nil, fmt.Errorf("no data for to find PodName")
}
