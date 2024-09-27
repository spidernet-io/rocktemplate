package ebpf

import (
	"encoding/binary"
	"fmt"
	"github.com/spidernet-io/rocktemplate/pkg/nodeId"
	"github.com/spidernet-io/rocktemplate/pkg/types"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"
	"net"
	"reflect"
)

// map the key and value of the nodeIp map
type nodeIpMapData struct {
	key *bpf_cgroupMapkeyNodeIp
	val *uint32
}

func (s *EbpfProgramStruct) applyEpfMapDataNodeIpV4(l *zap.Logger, oldNode *corev1.Node, newNode *corev1.Node) error {
	buildDataFunc := func(node *corev1.Node) []*nodeIpMapData {
		if node == nil {
			return []*nodeIpMapData{}
		}
		nodeIpMapList := []*nodeIpMapData{}
		for _, v := range node.Status.Addresses {
			t := net.ParseIP(v.Address)
			if t.To4() != nil {
				nodeIpMapList = append(nodeIpMapList, &nodeIpMapData{
					key: &bpf_cgroupMapkeyNodeIp{
						IpAddr: binary.LittleEndian.Uint32(t.To4()),
					},
					val: ptr.To(uint32(0)),
				})
			}
		}
		return nodeIpMapList
	}
	// ---- build key and value first
	oldList := buildDataFunc(oldNode)
	newList := buildDataFunc(newNode)

	// ----- diff
	delKeyList := []bpf_cgroupMapkeyNodeIp{}
	addKeyList := []bpf_cgroupMapkeyNodeIp{}
	addValList := []uint32{}

	l.Sugar().Debugf("nodeIp map %d items in oldList: ", len(oldList))
	for k, v := range oldList {
		l.Sugar().Debugf("nodeIp map oldList[%d]: key=%s, value=%d ", k, *v.key, *v.val)
	}
	l.Sugar().Debugf("nodeIp map %d items in newList: ", len(newList))
	for k, v := range newList {
		l.Sugar().Debugf("nodeIp map newList[%d]: key=%s, value=%d ", k, *v.key, *v.val)
	}

OUTER_OLD:
	for _, oldKey := range oldList {
		for _, newKey := range newList {
			if reflect.DeepEqual(oldKey.key, newKey.key) {
				if !reflect.DeepEqual(oldKey.val, newKey.val) {
					addKeyList = append(addKeyList, *newKey.key)
					addValList = append(addValList, *newKey.val)
					l.Sugar().Infof("ebpf map of the nodeIp updates: key=%s , value=%d ", newKey.key, newKey.val)
				}
				continue OUTER_OLD
			}
		}
		l.Sugar().Infof("ebpf map of the nodeIp deletes: key=%s , value=%d ", oldKey.key, oldKey.val)
		delKeyList = append(delKeyList, *oldKey.key)
	}

OUTER_NEW:
	for _, newKey := range newList {
		for _, oldKey := range oldList {
			if reflect.DeepEqual(oldKey.key, newKey.key) {
				continue OUTER_NEW
			}
		}
		addKeyList = append(addKeyList, *newKey.key)
		addValList = append(addValList, *newKey.val)
		l.Sugar().Infof("ebpf map of the nodeIp updates: key=%s , value=%d ", newKey.key, newKey.val)
	}

	// -------- apply
	// must deletion first, then apply new .
	if len(delKeyList) > 0 {
		if e := s.DeleteMapNodeIp(delKeyList); e != nil {
			l.Sugar().Errorf("failed to delete nodeIp map: %v", e)
			return fmt.Errorf("failed to delete nodeIp map: %v", e)
		}
		l.Sugar().Infof("succeeded to delete %d items in nodeIp data ", len(delKeyList))
	}
	if len(addKeyList) > 0 {
		if e := s.UpdateMapNodeIp(addKeyList, addValList); e != nil {
			l.Sugar().Errorf("failed to update nodeIp map: %v", e)
			return fmt.Errorf("failed to update nodeIp map: %v", e)
		}
		l.Sugar().Infof("succeeded to update %d items in nodeIp map: ", len(addKeyList))
	}

	return nil
}

// map the key and value of the nodeEntryIp map
type nodeEntryIpMapData struct {
	key *uint32
	val *bpf_cgroupMapvalueNodeEntryIp
}

func (s *EbpfProgramStruct) applyEpfMapDataNodeEntryIpV4(l *zap.Logger, oldNode *corev1.Node, newNode *corev1.Node) error {

	l.Sugar().Debugf("applyEpfMapDataNodeEntryIpV4 1 ")

	if newNode == nil && oldNode == nil {
		return fmt.Errorf("empty node obj")
	}

	l.Sugar().Debugf("applyEpfMapDataNodeEntryIpV4 2 ")

	// each node just has only one key
	if newNode == nil && oldNode != nil {
		// delete node
		nodeId, err := nodeId.NodeIdManagerHander.GetNodeId(oldNode.Name)
		if err != nil {
			l.Sugar().Errorf("failed to find the nodeIP for node %s when deleting ebpf data: %v", oldNode.Name, err)
			return fmt.Errorf("failed to find the nodeIP for node %s when deleting ebpf data: %v", oldNode.Name, err)
		}

		l.Sugar().Infof("ebpf map of the nodeEntryIP deletes: key=%d ", nodeId)
		err = s.DeleteMapNodeEntryIp([]uint32{nodeId})
		if err != nil {
			l.Sugar().Errorf("failed to update nodeEntryIP map: %v", err)
			return fmt.Errorf("failed to update nodeEntryIP map: %v", err)
		}
		l.Sugar().Infof("succeeded to update 1 items in nodeEntryIP map: ")
		return nil
	}

	l.Sugar().Debugf("applyEpfMapDataNodeEntryIpV4 3 ")

	// update or create
	entryIp, _ := newNode.ObjectMeta.Annotations[types.NodeAnnotaitonNodeEntryIPv4]
	if len(entryIp) != 0 && net.ParseIP(entryIp).To4() == nil {
		l.Sugar().Errorf("the v4 entryIp %s defined by the use is invalid, use the internal ip of the node %s ", entryIp, newNode.Name)
		entryIp = ""
	}
	if len(entryIp) == 0 {
		// for the internal ip
		for _, v := range newNode.Status.Addresses {
			t := net.ParseIP(v.Address)
			if t.To4() != nil {
				entryIp = v.Address
				break
			}
		}
		if len(entryIp) == 0 {
			l.Sugar().Errorf("did not find ipv4 internal ip for node %s", newNode.Name)
			return fmt.Errorf("did not find ipv4 internal ip for node %s", newNode.Name)
		}
	}

	l.Sugar().Debugf("applyEpfMapDataNodeEntryIpV4 4 %s", entryIp)

	nodeId, err := nodeId.NodeIdManagerHander.GetNodeId(newNode.Name)
	l.Sugar().Debugf("applyEpfMapDataNodeEntryIpV4 5 ")

	if err != nil {
		l.Sugar().Errorf("failed to find the nodeIP for node %s when updating ebpf data: %v", oldNode.Name, err)
		return fmt.Errorf("failed to find the nodeIP for node %s when updating ebpf data: %v", oldNode.Name, err)
	}
	l.Sugar().Debugf("applyEpfMapDataNodeEntryIpV4 6 ")
	r := bpf_cgroupMapvalueNodeEntryIp{
		IpAddr: binary.LittleEndian.Uint32(net.ParseIP(entryIp).To4()),
	}
	l.Sugar().Debugf("applyEpfMapDataNodeEntryIpV4 7 ")

	l.Sugar().Infof("ebpf map of the nodeEntryIP updates: key=%d , value=%s ", nodeId, r.String())

	err = s.UpdateMapNodeEntryIp([]uint32{nodeId}, []bpf_cgroupMapvalueNodeEntryIp{r})
	if err != nil {
		l.Sugar().Errorf("failed to update nodeEntryIP map: %v", err)
		return fmt.Errorf("failed to update nodeEntryIP map: %v", err)
	}

	l.Sugar().Debugf("applyEpfMapDataNodeEntryIpV4 8 ")

	l.Sugar().Infof("succeeded to update 1 items in nodeEntryIP map ")
	return nil
}

func (s *EbpfProgramStruct) UpdateEbpfMapForNode(l *zap.Logger, oldNode *corev1.Node, newNode *corev1.Node) error {

	// for ipv4
	if true {
		l.Sugar().Infof("UpdateEbpfMapForNode for ipv4 ")

		if e := s.applyEpfMapDataNodeIpV4(l, oldNode, newNode); e != nil {
			return fmt.Errorf("failed to applyEpfMapDataNodeIpV4: %v", e)
		}
		if e := s.applyEpfMapDataNodeEntryIpV4(l, oldNode, newNode); e != nil {
			return fmt.Errorf("failed to applyEpfMapDataNodeEntryIpV4: %v", e)
		}
	}

	// for ipv6
	if false {
		l.Sugar().Infof("does not suppport ipv6, abandon applying ")
	}

	return nil
}

func (s *EbpfProgramStruct) DeleteEbpfMapForNode(l *zap.Logger, node *corev1.Node) error {

	// for ipv4
	if true {
		l.Sugar().Infof("DeleteEbpfMapForNode for ipv4 ")

		if e := s.applyEpfMapDataNodeIpV4(l, node, nil); e != nil {
			return fmt.Errorf("failed to applyEpfMapDataNodeIpV4: %v", e)
		}

		if e := s.applyEpfMapDataNodeEntryIpV4(l, node, nil); e != nil {
			return fmt.Errorf("failed to applyEpfMapDataNodeEntryIpV4: %v", e)
		}
	}

	// for ipv6
	if false {
		l.Sugar().Infof("does not suppport ipv6, abandon applying ")
	}
	return nil
}
