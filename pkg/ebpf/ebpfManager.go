package ebpf

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -no-strip -cc clang -target bpf -cflags "-D__TARGET_ARCH_x86"  bpf_cgroup   bpf/cgroup.c

import (
	"fmt"
	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	// "github.com/cilium/ebpf/ringbuf"
	"github.com/cilium/ebpf/rlimit"
	"os"
)

// -------------------------------------------
const (
	HostMountInfoPath = "/proc/mounts"
	BpfFSPath         = "/sys/fs/bpf"
	MapsPinpath       = BpfFSPath + "/welan"
	CgroupV2Path      = "/run/welan/cgroupv2"

	EventChanLength = 1000
)

const (
	NatTypeService = iota
	NatTypeLocalRedirect
	NatTypeFloatIP
)

const (
	PROTOCOL_TCP = 6
	PROTOCOL_UDP = 17
)

// -----------------------------------

type EbpfProgramStruct struct {
	BpfObjCgroup bpf_cgroupObjects
	CgroupLink   link.Link
	Event        chan MapEventValue
}

type EbpfProgram interface {
	LoadProgramp() error
	UnloadProgramp() error

	GetMapDataEvent() <-chan MapEventValue

	PrintMapService() error
	PrintMapNode() error
	PrintMapBackend() error
	PrintMapAffinity() error
	PrintMapNatRecord() error

	UpdateMapService([]bpf_cgroupMapkeyService, []bpf_cgroupMapvalueService) error
	UpdateMapBackend([]bpf_cgroupMapkeyBackend, []bpf_cgroupMapvalueBackend) error
	UpdateMapNode([]bpf_cgroupMapkeyNode, []uint32) error

	// DeleteMapDataBackendV4([]uint32) error
	// UpdateMapDataBackendV4([]uint32, []bpf_cgroupMapvalueBackendV4) error
	// PrintMapDataBackendV4()
}

var _ EbpfProgram = &EbpfProgramStruct{}

// ------------------------------------

func NewEbpfProgramMananger() EbpfProgram {
	return &EbpfProgramStruct{
		Event: make(chan MapEventValue, EventChanLength),
	}
}

func (s *EbpfProgramStruct) LoadProgramp() error {

	if err := checkOrMountBpfFs(BpfFSPath); err != nil {
		return fmt.Errorf("failed to mount bpf fs: %v", err)
	}

	// Allow the current process to lock memory for eBPF resources.
	if err := rlimit.RemoveMemlock(); err != nil {
		return fmt.Errorf("failed to RemoveMemlock:", err)
	}

	// attach to cgroup
	// sync.Once.Do(func() {
	if err := checkOrMountCgroupV2(CgroupV2Path); err != nil {
		return fmt.Errorf("failed to mount cgroup v2: %s", err)
	}
	// })

	// create the directory for map pin path
	if stat, err := os.Stat(MapsPinpath); err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(MapsPinpath, os.ModePerm)
			if err != nil {
				return fmt.Errorf("failed to mkdir %s: %v", MapsPinpath, err)
			}
		} else {
			return fmt.Errorf("Failed to stat the path %s: %w", MapsPinpath, err)
		}
	} else {
		if !stat.IsDir() {
			return fmt.Errorf("%s is a file which is not a directory", MapsPinpath)
		}
	}

	// 这个函数( loadBpf_xxxObjects )是 bpf2go 生成的 go 文件中 加载 ebpf 程序到内核
	err := loadBpf_cgroupObjects(&s.BpfObjCgroup, &ebpf.CollectionOptions{
		Maps: ebpf.MapOptions{
			PinPath: MapsPinpath,
		}})
	if err != nil {
		return fmt.Errorf("failed to load ebpf obj: %v", err)
	}

	// 把 ebpf 程序再挂载到 cgroup
	// https://github.com/cilium/ebpf/blob/main/link/cgroup.go#L43
	s.CgroupLink, err = link.AttachCgroup(link.CgroupOptions{
		Path:    CgroupV2Path,
		Attach:  ebpf.AttachCGroupInet4Connect,
		Program: s.BpfObjCgroup.bpf_cgroupPrograms.Sock4Connect,
	})
	if err != nil {
		return fmt.Errorf("Error attaching Sock4Connect to cgroup: %v", err)
	}
	s.CgroupLink, err = link.AttachCgroup(link.CgroupOptions{
		Path:    CgroupV2Path,
		Attach:  ebpf.AttachCGroupUDP4Sendmsg,
		Program: s.BpfObjCgroup.bpf_cgroupPrograms.Sock4Sendmsg,
	})
	if err != nil {
		return fmt.Errorf("Error attaching Sock4Sendmsg to cgroup: %v", err)
	}
	s.CgroupLink, err = link.AttachCgroup(link.CgroupOptions{
		Path:    CgroupV2Path,
		Attach:  ebpf.AttachCGroupUDP4Recvmsg,
		Program: s.BpfObjCgroup.bpf_cgroupPrograms.Sock4Recvmsg,
	})
	if err != nil {
		return fmt.Errorf("Error attaching Sock4Recvmsg to cgroup: %v", err)
	}
	s.CgroupLink, err = link.AttachCgroup(link.CgroupOptions{
		Path:    CgroupV2Path,
		Attach:  ebpf.AttachCgroupInet4GetPeername,
		Program: s.BpfObjCgroup.bpf_cgroupPrograms.Sock4Getpeername,
	})
	if err != nil {
		return fmt.Errorf("Error attaching Sock4Getpeername to cgroup: %v", err)
	}

	go s.daemonGetEvent()

	return nil
}

func (s *EbpfProgramStruct) UnloadProgramp() error {

	if s.CgroupLink != nil {
		fmt.Printf("Closing  cgroup v2 ...\n")
		s.CgroupLink.Close()
	}

	// unping and close ebpf map
	if s.BpfObjCgroup.bpf_cgroupMaps.MapBackend != nil {
		s.BpfObjCgroup.bpf_cgroupMaps.MapBackend.Unpin()
		s.BpfObjCgroup.bpf_cgroupMaps.MapBackend.Close()
	}
	if s.BpfObjCgroup.bpf_cgroupMaps.MapService != nil {
		s.BpfObjCgroup.bpf_cgroupMaps.MapService.Unpin()
		s.BpfObjCgroup.bpf_cgroupMaps.MapService.Close()
	}
	if s.BpfObjCgroup.bpf_cgroupMaps.MapAffinity != nil {
		s.BpfObjCgroup.bpf_cgroupMaps.MapAffinity.Unpin()
		s.BpfObjCgroup.bpf_cgroupMaps.MapAffinity.Close()
	}
	if s.BpfObjCgroup.bpf_cgroupMaps.MapNode != nil {
		s.BpfObjCgroup.bpf_cgroupMaps.MapNode.Unpin()
		s.BpfObjCgroup.bpf_cgroupMaps.MapNode.Close()
	}
	if s.BpfObjCgroup.bpf_cgroupMaps.MapNatRecord != nil {
		s.BpfObjCgroup.bpf_cgroupMaps.MapNatRecord.Unpin()
		s.BpfObjCgroup.bpf_cgroupMaps.MapNatRecord.Close()
	}
	if s.BpfObjCgroup.bpf_cgroupMaps.MapEvent != nil {
		s.BpfObjCgroup.bpf_cgroupMaps.MapEvent.Unpin()
		s.BpfObjCgroup.bpf_cgroupMaps.MapEvent.Close()
	}

	fmt.Printf("Closing progs ...\n")
	s.BpfObjCgroup.bpf_cgroupPrograms.Close()
	s.BpfObjCgroup.bpf_cgroupMaps.Close()

	s.BpfObjCgroup.Close()

	return nil
}

// ------------------------------------------- map

// get data from map
// func (s *EbpfProgramStruct) daemonGetEvent() {
//
// 	rd, err := ringbuf.NewReader(s.BpfObjCgroup.bpf_cgroupMaps.MapEvent)
// 	if err != nil {
// 		fmt.Printf("failed to read ebpf map : %v ", err)
// 		os.Exit(1)
// 	}
// 	defer rd.Close()
//
// 	for {
// 		record, err := rd.Read()
// 		if err != nil {
// 			if errors.Is(err, ringbuf.ErrClosed) {
// 				fmt.Printf("received signal, exiting reading events..\n")
// 				break
// 			}
// 			fmt.Printf("failed to read event: %v", err)
// 			continue
// 		}
//
// 		t := MapEventValue{}
// 		if err := binary.Read(bytes.NewBuffer(record.RawSample), binary.NativeEndian, &t); err != nil {
// 			fmt.Printf("parsing ringbuf event: %s", err)
// 			continue
// 		}
// 		// fmt.Printf("get event data: %v \n", t)
//
// 		select {
// 		case s.Event <- t:
// 		default:
// 			fmt.Printf("error, failed to write data to event chan, miss data: %v \n", t)
// 		}
// 	}
//
// }

// -----------------------

// func (s *EbpfProgramStruct) DeleteMapDataFloatip(keyList []bpf_cgroupMapkeyFloatipV4) error {
// 	if s.BpfObjCgroup.MapFloatipV4 == nil {
// 		return fmt.Errorf("ebpf map does not exist")
// 	}
// 	if keyList != nil && len(keyList) == 0 {
// 		return fmt.Errorf("empty key")
// 	}
//
// 	if keyList == nil {
// 		// delete all
// 		keys := make([]bpf_cgroupMapkeyFloatipV4, 100)
// 		vals := make([]bpf_cgroupMapvlaueFloatipV4, 100)
//
// 		var cursor ebpf.MapBatchCursor
// 		count := 0
// 		for {
// 			c, batchErr := s.BpfObjCgroup.MapFloatipV4.BatchLookup(&cursor, keys, vals, nil)
// 			count += c
// 			if batchErr != nil {
// 				if errors.Is(batchErr, ebpf.ErrKeyNotExist) {
// 					// end
// 					break
// 				}
// 				return fmt.Errorf("failed to batchlookup , reason: %v ", batchErr)
// 			}
// 			if _, batchErr = s.BpfObjCgroup.MapFloatipV4.BatchDelete(keys, &ebpf.BatchOptions{}); batchErr != nil {
// 				return fmt.Errorf("failed to BatchDelete , reason: %v ", batchErr)
// 			}
// 		}
// 		fmt.Printf("delted item account: %v \n", count)
// 	} else {
// 		if _, batchErr := s.BpfObjCgroup.MapFloatipV4.BatchDelete(keyList, &ebpf.BatchOptions{}); batchErr != nil {
// 			return fmt.Errorf("failed to BatchDelete , reason: %v ", batchErr)
// 		}
// 	}
// 	return nil
// }
