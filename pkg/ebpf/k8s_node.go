package ebpf

import (
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
)

// map the key and value of the node map
type nodeMapData struct {
	key *bpf_cgroupMapkeyService
	val *bpf_cgroupMapvalueService
}

func (s *EbpfProgramStruct) UpdateEbpfMapForNode(l *zap.Logger, oldNode *corev1.Node, newNode *corev1.Node) error {

}

func (s *EbpfProgramStruct) DeleteEbpfMapForNode(l *zap.Logger, node *corev1.Node) error {

}
