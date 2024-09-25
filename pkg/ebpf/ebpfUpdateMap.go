package ebpf

import (
	"fmt"
	"github.com/cilium/ebpf"
)

// ------------------------- update ----------------------------------

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

func (s *EbpfProgramStruct) UpdateMapNatRecord(keyList []bpf_cgroupMapkeyNatRecord, valueList []bpf_cgroupMapvalueNatRecord) error {
	if keyList == nil || valueList == nil || len(keyList) == 0 || len(valueList) == 0 {
		return fmt.Errorf("empty parameter")
	}
	if len(keyList) != len(valueList) {
		return fmt.Errorf("invalid parameter")
	}

	c, e := s.BpfObjCgroup.MapNatRecord.BatchUpdate(keyList, valueList, &ebpf.BatchOptions{})
	if e != nil {
		return fmt.Errorf("failed to BatchUpdate: %+v", e)
	}

	if len(keyList) != c {
		return fmt.Errorf("update account %v , different from expected account %v ", c, len(keyList))
	}
	return nil
}

func (s *EbpfProgramStruct) UpdateMapAffinity(keyList []bpf_cgroupMapkeyAffinity, valueList []bpf_cgroupMapvalueAffinity) error {
	if keyList == nil || valueList == nil || len(keyList) == 0 || len(valueList) == 0 {
		return fmt.Errorf("empty parameter")
	}
	if len(keyList) != len(valueList) {
		return fmt.Errorf("invalid parameter")
	}

	c, e := s.BpfObjCgroup.MapAffinity.BatchUpdate(keyList, valueList, &ebpf.BatchOptions{})
	if e != nil {
		return fmt.Errorf("failed to BatchUpdate: %+v", e)
	}

	if len(keyList) != c {
		return fmt.Errorf("update account %v , different from expected account %v ", c, len(keyList))
	}
	return nil
}

// ------------------------- delete ----------------------------------

func (s *EbpfProgramStruct) DeleteMapService(keyList []bpf_cgroupMapkeyService) error {
	if keyList == nil || len(keyList) == 0 {
		return nil
	}
	c, e := s.BpfObjCgroup.MapService.BatchDelete(keyList, &ebpf.BatchOptions{})
	if e != nil {
		return fmt.Errorf("failed to BatchDelete: %+v", e)
	}
	if len(keyList) != c {
		return fmt.Errorf("deleted account %v , different from expected account %v ", c, len(keyList))
	}
	return nil
}

func (s *EbpfProgramStruct) DeleteMapBackend(keyList []bpf_cgroupMapkeyBackend) error {
	if keyList == nil || len(keyList) == 0 {
		return nil
	}
	c, e := s.BpfObjCgroup.MapBackend.BatchDelete(keyList, &ebpf.BatchOptions{})
	if e != nil {
		return fmt.Errorf("failed to BatchDelete: %+v", e)
	}
	if len(keyList) != c {
		return fmt.Errorf("deleted account %v , different from expected account %v ", c, len(keyList))
	}
	return nil
}

func (s *EbpfProgramStruct) DeleteMapNode(keyList []bpf_cgroupMapkeyNode) error {
	if keyList == nil || len(keyList) == 0 {
		return nil
	}
	c, e := s.BpfObjCgroup.MapNode.BatchDelete(keyList, &ebpf.BatchOptions{})
	if e != nil {
		return fmt.Errorf("failed to BatchDelete: %+v", e)
	}
	if len(keyList) != c {
		return fmt.Errorf("deleted account %v , different from expected account %v ", c, len(keyList))
	}
	return nil
}

func (s *EbpfProgramStruct) DeleteMapAffinity(keyList []bpf_cgroupMapkeyAffinity) error {
	if keyList == nil || len(keyList) == 0 {
		return nil
	}
	c, e := s.BpfObjCgroup.MapAffinity.BatchDelete(keyList, &ebpf.BatchOptions{})
	if e != nil {
		return fmt.Errorf("failed to BatchDelete: %+v", e)
	}
	if len(keyList) != c {
		return fmt.Errorf("deleted account %v , different from expected account %v ", c, len(keyList))
	}
	return nil
}

func (s *EbpfProgramStruct) DeleteMapNatRecord(keyList []bpf_cgroupMapkeyNatRecord) error {
	if keyList == nil || len(keyList) == 0 {
		return nil
	}
	c, e := s.BpfObjCgroup.MapNatRecord.BatchDelete(keyList, &ebpf.BatchOptions{})
	if e != nil {
		return fmt.Errorf("failed to BatchDelete: %+v", e)
	}
	if len(keyList) != c {
		return fmt.Errorf("deleted account %v , different from expected account %v ", c, len(keyList))
	}
	return nil
}
