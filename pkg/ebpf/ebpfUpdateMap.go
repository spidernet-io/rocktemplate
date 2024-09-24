package ebpf

import (
	"fmt"
	"github.com/cilium/ebpf"
)

func (s *EbpfProgramStruct) UpdateMapService(keyList []bpf_cgroupMapkeyService, valueList []bpf_cgroupMapvalueService) error {
	if keyList == nil || valueList == nil || len(keyList) == 0 || len(valueList) == 0 {
		return fmt.Errorf("empty parameter")
	}
	if len(keyList) != len(valueList) {
		return fmt.Errorf("invalid parameter")
	}

	c, e := s.BpfObjCgroup.MapService.BatchUpdate(keyList, valueList, &ebpf.BatchOptions{})
	if e != nil {
		return fmt.Errorf("failed to BatchUpdate: %+v", e)
	}

	if len(keyList) != c {
		return fmt.Errorf("update account %v , different from expected account %v ", c, len(keyList))
	}
	return nil
}

func (s *EbpfProgramStruct) UpdateMapBackend(keyList []bpf_cgroupMapkeyBackend, valueList []bpf_cgroupMapvalueBackend) error {
	if keyList == nil || valueList == nil || len(keyList) == 0 || len(valueList) == 0 {
		return fmt.Errorf("empty parameter")
	}
	if len(keyList) != len(valueList) {
		return fmt.Errorf("invalid parameter")
	}

	c, e := s.BpfObjCgroup.MapBackend.BatchUpdate(keyList, valueList, &ebpf.BatchOptions{})
	if e != nil {
		return fmt.Errorf("failed to BatchUpdate: %+v", e)
	}

	if len(keyList) != c {
		return fmt.Errorf("update account %v , different from expected account %v ", c, len(keyList))
	}
	return nil
}

func (s *EbpfProgramStruct) UpdateMapNode(keyList []bpf_cgroupMapkeyNode, valueList []uint32) error {
	if keyList == nil || valueList == nil || len(keyList) == 0 || len(valueList) == 0 {
		return fmt.Errorf("empty parameter")
	}
	if len(keyList) != len(valueList) {
		return fmt.Errorf("invalid parameter")
	}

	c, e := s.BpfObjCgroup.MapNode.BatchUpdate(keyList, valueList, &ebpf.BatchOptions{})
	if e != nil {
		return fmt.Errorf("failed to BatchUpdate: %+v", e)
	}

	if len(keyList) != c {
		return fmt.Errorf("update account %v , different from expected account %v ", c, len(keyList))
	}
	return nil
}
