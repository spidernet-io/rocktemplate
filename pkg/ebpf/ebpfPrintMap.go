package ebpf

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/perf"
	"os"
)

// func (s *EbpfProgramStruct) PrintMapDataService() error {
//
// 	keys := make([]bpf_cgroupMapkeyService, 100)
// 	vals := make([]bpf_cgroupMapvalueService, 100)
//
// 	var cursor ebpf.MapBatchCursor
// 	count := 0
// 	for {
// 		c, batchErr := s.BpfObjCgroup.MapService.BatchLookup(&cursor, keys, vals, nil)
// 		count += c
// 		finished := false
// 		if batchErr != nil {
// 			if errors.Is(batchErr, ebpf.ErrKeyNotExist) {
// 				// end
// 				finished = true
// 			} else {
// 				return fmt.Errorf("failed to batchlookup for %v\n", s.BpfObjCgroup.MapService.String())
// 			}
// 		}
// 		for i := 0; i < len(keys) && i < c; i++ {
// 			fmt.Printf("%v : key=%+v\n", i, keys[i])
// 			fmt.Printf("%v : value=%+v\n", i, vals[i])
// 			fmt.Printf("\n")
// 		}
// 		if finished {
// 			break
// 		}
// 	}
// 	fmt.Printf("end map %s: total items: %v \n",name, count)
// 	return nil
// }

type MapName string

const (
	MapNameService   = MapName("mapService")
	MapNameAffinity  = MapName("mapAffinity")
	MapNameBackend   = MapName("mapBackend")
	MapNameNatRecord = MapName("mapNatRecord")
	MapNameNode      = MapName("mapNode")
)

func (s *EbpfProgramStruct) PrintMapService() error {
	keys := make([]bpf_cgroupMapkeyService, 100)
	vals := make([]bpf_cgroupMapvalueService, 100)
	mapPtr := s.BpfObjCgroup.MapService
	name := MapNameService

	fmt.Printf("------------------------------\n")
	fmt.Printf("ebgin map  %s\n", name)
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
				return fmt.Errorf("failed to batchlookup for %v\n", mapPtr.String())
			}
		}
		for i := 0; i < len(keys) && i < c; i++ {
			fmt.Printf("%v : key=%+v\n", i, keys[i])
			fmt.Printf("%v : value=%+v\n", i, vals[i])
		}
		if finished {
			break
		}
	}

	fmt.Printf("end map %s: total items: %v \n", name, count)
	fmt.Printf("------------------------------\n")
	fmt.Printf("\n")
	return nil
}

func (s *EbpfProgramStruct) PrintMapBackend() error {
	keys := make([]bpf_cgroupMapkeyBackend, 100)
	vals := make([]bpf_cgroupMapvalueBackend, 100)
	mapPtr := s.BpfObjCgroup.MapBackend
	name := MapNameBackend

	fmt.Printf("------------------------------\n")
	fmt.Printf("ebgin map  %s\n", name)
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
				return fmt.Errorf("failed to batchlookup for %v\n", mapPtr.String())
			}
		}
		for i := 0; i < len(keys) && i < c; i++ {
			fmt.Printf("%v : key=%+v\n", i, keys[i])
			fmt.Printf("%v : value=%+v\n", i, vals[i])
		}
		if finished {
			break
		}
	}

	fmt.Printf("end map %s: total items: %v \n", name, count)
	fmt.Printf("------------------------------\n")
	fmt.Printf("\n")
	return nil
}

func (s *EbpfProgramStruct) PrintMapNode() error {
	keys := make([]bpf_cgroupMapkeyNode, 100)
	vals := make([]uint32, 100)
	mapPtr := s.BpfObjCgroup.MapNode
	name := MapNameNode

	fmt.Printf("------------------------------\n")
	fmt.Printf("ebgin map  %s\n", name)
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
				return fmt.Errorf("failed to batchlookup for %v\n", mapPtr.String())
			}
		}
		for i := 0; i < len(keys) && i < c; i++ {
			fmt.Printf("%v : key=%+v\n", i, keys[i])
			fmt.Printf("%v : value=%+v\n", i, vals[i])
		}
		if finished {
			break
		}
	}

	fmt.Printf("end map %s: total items: %v \n", name, count)
	fmt.Printf("------------------------------\n")
	fmt.Printf("\n")
	return nil
}

func (s *EbpfProgramStruct) PrintMapAffinity() error {
	keys := make([]bpf_cgroupMapkeyAffinity, 100)
	vals := make([]bpf_cgroupMapvalueAffinity, 100)
	mapPtr := s.BpfObjCgroup.MapAffinity
	name := MapNameAffinity

	fmt.Printf("------------------------------\n")
	fmt.Printf("ebgin map  %s\n", name)
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
				return fmt.Errorf("failed to batchlookup for %v\n", mapPtr.String())
			}
		}
		for i := 0; i < len(keys) && i < c; i++ {
			fmt.Printf("%v : key=%+v\n", i, keys[i])
			fmt.Printf("%v : value=%+v\n", i, vals[i])
		}
		if finished {
			break
		}
	}

	fmt.Printf("end map %s: total items: %v \n", name, count)
	fmt.Printf("------------------------------\n")
	fmt.Printf("\n")
	return nil
}

func (s *EbpfProgramStruct) PrintMapNatRecord() error {
	keys := make([]bpf_cgroupMapkeyNatRecord, 100)
	vals := make([]bpf_cgroupMapvalueNatRecord, 100)
	mapPtr := s.BpfObjCgroup.MapNatRecord
	name := MapNameNatRecord

	fmt.Printf("------------------------------\n")
	fmt.Printf("ebgin map  %s\n", name)
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
				return fmt.Errorf("failed to batchlookup for %v\n", mapPtr.String())
			}
		}
		for i := 0; i < len(keys) && i < c; i++ {
			fmt.Printf("%v : key=%+v\n", i, keys[i])
			fmt.Printf("%v : value=%+v\n", i, vals[i])
		}
		if finished {
			break
		}
	}

	fmt.Printf("end map %s: total items: %v \n", name, count)
	fmt.Printf("------------------------------\n")
	fmt.Printf("\n")
	return nil
}

// -------------------------- event map

func (s *EbpfProgramStruct) GetMapDataEvent() <-chan MapEventValue {
	return s.Event

}

// get data from map
func (s *EbpfProgramStruct) daemonGetEvent() {

	rd, err := perf.NewReader(s.BpfObjCgroup.bpf_cgroupMaps.MapEvent, os.Getpagesize())

	if err != nil {
		fmt.Printf("failed to read ebpf map : %v ", err)
		os.Exit(1)
	}
	defer rd.Close()

	for {
		record, err := rd.Read()
		if err != nil {
			fmt.Printf("failed to read event: %v", err)
			continue
		}

		t := MapEventValue{}
		if err := binary.Read(bytes.NewBuffer(record.RawSample), binary.NativeEndian, &t); err != nil {
			fmt.Printf("parsing ringbuf event: %s", err)
			continue
		}
		// fmt.Printf("get event data: %v \n", t)

		select {
		case s.Event <- t:
		default:
			fmt.Printf("error, failed to write data to event chan, miss data: %v \n", t)
		}
	}
}
