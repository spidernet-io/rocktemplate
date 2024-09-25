package ebpf

import (
	"errors"
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

// ------------------------- clean ----------------------------------

func (s *EbpfProgramStruct) CleanMapService() (int, error) {
	keys := make([]bpf_cgroupMapkeyService, 100)
	vals := make([]bpf_cgroupMapvalueService, 100)

	var mapPtr *ebpf.Map
	if s.BpfObjCgroup.MapService != nil {
		mapPtr = s.BpfObjCgroup.MapService
	} else if s.EbpfMaps != nil && s.EbpfMaps.MapService != nil {
		mapPtr = s.EbpfMaps.MapService
	} else {
		return 0, fmt.Errorf("failed to get ebpf map")
	}

	var cursor ebpf.MapBatchCursor
	count := 0
	for {
		c, batchErr := mapPtr.BatchLookup(&cursor, keys, vals, nil)
		count += c
		finished := false
		if batchErr != nil {
			if errors.Is(batchErr, ebpf.ErrKeyNotExist) {
				// end
				finished = true
			} else {
				return 0, fmt.Errorf("failed to batchlookup for %v\n", mapPtr.String())
			}
		}
		c, e := s.BpfObjCgroup.MapNatRecord.BatchDelete(keys, &ebpf.BatchOptions{})
		if e != nil {
			return 0, fmt.Errorf("failed to BatchDelete: %+v", e)
		}
		if len(keys) != c {
			return 0, fmt.Errorf("deleted account %v , different from expected account %v ", c, len(keys))
		}
		if finished {
			break
		}
	}
	return count, nil
}

func (s *EbpfProgramStruct) CleanMapBackend() (int, error) {
	keys := make([]bpf_cgroupMapkeyBackend, 100)
	vals := make([]bpf_cgroupMapvalueBackend, 100)

	var mapPtr *ebpf.Map
	if s.BpfObjCgroup.MapBackend != nil {
		mapPtr = s.BpfObjCgroup.MapBackend
	} else if s.EbpfMaps != nil && s.EbpfMaps.MapBackend != nil {
		mapPtr = s.EbpfMaps.MapBackend
	} else {
		return 0, fmt.Errorf("failed to get ebpf map")
	}

	var cursor ebpf.MapBatchCursor
	count := 0
	for {
		c, batchErr := mapPtr.BatchLookup(&cursor, keys, vals, nil)
		count += c
		finished := false
		if batchErr != nil {
			if errors.Is(batchErr, ebpf.ErrKeyNotExist) {
				// end
				finished = true
			} else {
				return 0, fmt.Errorf("failed to batchlookup for %v\n", mapPtr.String())
			}
		}
		c, e := s.BpfObjCgroup.MapNatRecord.BatchDelete(keys, &ebpf.BatchOptions{})
		if e != nil {
			return 0, fmt.Errorf("failed to BatchDelete: %+v", e)
		}
		if len(keys) != c {
			return 0, fmt.Errorf("deleted account %v , different from expected account %v ", c, len(keys))
		}
		if finished {
			break
		}
	}
	return count, nil
}

func (s *EbpfProgramStruct) CleanMapNode() (int, error) {
	keys := make([]bpf_cgroupMapkeyNode, 100)
	vals := make([]uint32, 100)

	var mapPtr *ebpf.Map
	if s.BpfObjCgroup.MapNode != nil {
		mapPtr = s.BpfObjCgroup.MapNode
	} else if s.EbpfMaps != nil && s.EbpfMaps.MapNode != nil {
		mapPtr = s.EbpfMaps.MapNode
	} else {
		return 0, fmt.Errorf("failed to get ebpf map")
	}

	var cursor ebpf.MapBatchCursor
	count := 0
	for {
		c, batchErr := mapPtr.BatchLookup(&cursor, keys, vals, nil)
		count += c
		finished := false
		if batchErr != nil {
			if errors.Is(batchErr, ebpf.ErrKeyNotExist) {
				// end
				finished = true
			} else {
				return 0, fmt.Errorf("failed to batchlookup for %v\n", mapPtr.String())
			}
		}
		c, e := s.BpfObjCgroup.MapNatRecord.BatchDelete(keys, &ebpf.BatchOptions{})
		if e != nil {
			return 0, fmt.Errorf("failed to BatchDelete: %+v", e)
		}
		if len(keys) != c {
			return 0, fmt.Errorf("deleted account %v , different from expected account %v ", c, len(keys))
		}
		if finished {
			break
		}
	}
	return count, nil
}

func (s *EbpfProgramStruct) CleanMapAffinity() (int, error) {
	keys := make([]bpf_cgroupMapkeyAffinity, 100)
	vals := make([]bpf_cgroupMapvalueAffinity, 100)

	var mapPtr *ebpf.Map
	if s.BpfObjCgroup.MapAffinity != nil {
		mapPtr = s.BpfObjCgroup.MapAffinity
	} else if s.EbpfMaps != nil && s.EbpfMaps.MapAffinity != nil {
		mapPtr = s.EbpfMaps.MapAffinity
	} else {
		return 0, fmt.Errorf("failed to get ebpf map")
	}

	var cursor ebpf.MapBatchCursor
	count := 0
	for {
		c, batchErr := mapPtr.BatchLookup(&cursor, keys, vals, nil)
		count += c
		finished := false
		if batchErr != nil {
			if errors.Is(batchErr, ebpf.ErrKeyNotExist) {
				// end
				finished = true
			} else {
				return 0, fmt.Errorf("failed to batchlookup for %v\n", mapPtr.String())
			}
		}
		c, e := s.BpfObjCgroup.MapNatRecord.BatchDelete(keys, &ebpf.BatchOptions{})
		if e != nil {
			return 0, fmt.Errorf("failed to BatchDelete: %+v", e)
		}
		if len(keys) != c {
			return 0, fmt.Errorf("deleted account %v , different from expected account %v ", c, len(keys))
		}
		if finished {
			break
		}
	}
	return count, nil
}

func (s *EbpfProgramStruct) CleanMapNatRecord() (int, error) {
	keys := make([]bpf_cgroupMapkeyNatRecord, 100)
	vals := make([]bpf_cgroupMapvalueNatRecord, 100)

	var mapPtr *ebpf.Map
	if s.BpfObjCgroup.MapNatRecord != nil {
		mapPtr = s.BpfObjCgroup.MapNatRecord
	} else if s.EbpfMaps != nil && s.EbpfMaps.MapNatRecord != nil {
		mapPtr = s.EbpfMaps.MapNatRecord
	} else {
		return 0, fmt.Errorf("failed to get ebpf map")
	}

	var cursor ebpf.MapBatchCursor
	count := 0
	for {
		c, batchErr := mapPtr.BatchLookup(&cursor, keys, vals, nil)
		count += c
		finished := false
		if batchErr != nil {
			if errors.Is(batchErr, ebpf.ErrKeyNotExist) {
				// end
				finished = true
			} else {
				return 0, fmt.Errorf("failed to batchlookup for %v\n", mapPtr.String())
			}
		}
		c, e := s.BpfObjCgroup.MapNatRecord.BatchDelete(keys, &ebpf.BatchOptions{})
		if e != nil {
			return 0, fmt.Errorf("failed to BatchDelete: %+v", e)
		}
		if len(keys) != c {
			return 0, fmt.Errorf("deleted account %v , different from expected account %v ", c, len(keys))
		}
		if finished {
			break
		}
	}
	return count, nil
}
