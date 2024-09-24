package ebpf_test

import (
	"bytes"
	"encoding/binary"
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	"github.com/spidernet-io/rocktemplate/pkg/ebpf"
	"net"
	"os"
	"os/signal"
	"syscall"
)

func GenerateEndpointID(address net.IP, port int) (ID uint32) {
	err := binary.Read(bytes.NewBuffer(address), binary.BigEndian, &ID)
	if err != nil {
		fmt.Printf("failed to GenerateEndpointID : %v \n", err)
		return
	}
	ID = ID + uint32(port)

	return
}

func Run() {
	bpfManager := ebpf.NewEbpfProgramMananger()
	if err := bpfManager.LoadProgramp(); err != nil {
		fmt.Printf("failed to Load ebpf Programp: %v \n", err)
		return
	}
	defer bpfManager.UnloadProgramp()

	// ----------- update
	floatip := net.ParseIP("169.254.0.101")
	floatport := 8080

	pod1 := net.ParseIP("172.25.161.21")
	pod2 := net.ParseIP("172.25.132.16")
	podport := 80

	node1 := net.ParseIP("172.16.1.11")
	node2 := net.ParseIP("172.16.1.14")
	nodeport := 31931

	floatipLittle := binary.LittleEndian.Uint32(floatip.To4())
	pod1Little := binary.LittleEndian.Uint32(pod1.To4())
	pod2Little := binary.LittleEndian.Uint32(pod2.To4())
	node1Little := binary.LittleEndian.Uint32(node1.To4())
	node2Little := binary.LittleEndian.Uint32(node2.To4())

	var t [2]byte
	binary.LittleEndian.PutUint16(t[:], uint16(floatport))
	floatPortLittle := binary.LittleEndian.Uint16(t[:])

	binary.LittleEndian.PutUint16(t[:], uint16(podport))
	podPortLittle := binary.LittleEndian.Uint16(t[:])

	binary.LittleEndian.PutUint16(t[:], uint16(nodeport))
	nodePortLittle := binary.LittleEndian.Uint16(t[:])

	// ======================= set floatIP test
	var backendId uint32
	if err := binary.Read(bytes.NewBuffer(floatip.To4()), binary.BigEndian, &backendId); err != nil {
		fmt.Printf("failed to generate backend id")
		return
	}
	backendId += uint32(floatport)

	// ----------- set service map
	if true {
		var keyList []ebpf.bpf_cgroupMapkeyService
		var valueList []ebpf.bpf_cgroupMapvalueService
		keyList = append(keyList, ebpf.bpf_cgroupMapkeyService{
			Address: floatipLittle,
			Dport:   floatPortLittle,
			Proto:   ebpf.PROTOCOL_TCP,
			NatType: ebpf.NatTypeFloatIP,
			Scope:   0,
		})
		valueList = append(valueList, ebpf.bpf_cgroupMapvalueService{
			BackendId:         backendId,
			TotalBackendCount: 2,
			LocalBackendCount: 0,
			AffinityTimeout:   0,
			ServiceFlags:      0,
			FloatipFlags:      0,
			RedirectFlags:     0,
		})
		if e := bpfManager.UpdateMapService(keyList, valueList); e != nil {
			fmt.Printf("failed to set service map: %v\n", e)
			return
		}
	}
	// ----------- set backend map
	if true {
		var keyList []ebpf.bpf_cgroupMapkeyBackend
		keyList = append(keyList, ebpf.bpf_cgroupMapkeyBackend{
			BackendId: backendId,
			Order:     0,
		})
		keyList = append(keyList, ebpf.bpf_cgroupMapkeyBackend{
			BackendId: backendId,
			Order:     1,
		})
		var valueList []ebpf.bpf_cgroupMapvalueBackend
		valueList = append(valueList, ebpf.bpf_cgroupMapvalueBackend{
			PodAddress:  pod1Little,
			PodPort:     podPortLittle,
			NodeAddress: node1Little,
			NodePort:    nodePortLittle,
			Proto:       ebpf.PROTOCOL_TCP,
			Flags:       0,
		})
		valueList = append(valueList, ebpf.bpf_cgroupMapvalueBackend{
			PodAddress:  pod2Little,
			PodPort:     podPortLittle,
			NodeAddress: node2Little,
			NodePort:    nodePortLittle,
			Proto:       ebpf.PROTOCOL_TCP,
			Flags:       0,
		})
		if e := bpfManager.UpdateMapBackend(keyList, valueList); e != nil {
			fmt.Printf("failed to set backend map: %v\n", e)
			return
		}
	}

	// ======================================= 设置 节点
	// ----------- set node node
	if true {
		var keyList []ebpf.bpf_cgroupMapkeyNode
		keyList = append(keyList, ebpf.bpf_cgroupMapkeyNode{
			Address: node1Little,
		})
		keyList = append(keyList, ebpf.bpf_cgroupMapkeyNode{
			Address: node2Little,
		})
		valueList := make([]uint32, 2)
		if e := bpfManager.UpdateMapNode(keyList, valueList); e != nil {
			fmt.Printf("failed to set node map: %v\n", e)
			return
		}
	}

	// ---------------
	bpfManager.PrintMapService()
	bpfManager.PrintMapNode()
	bpfManager.PrintMapBackend()
	bpfManager.PrintMapAffinity()
	bpfManager.PrintMapNatRecord()

	// -----------
	go func() {
		for e := range bpfManager.GetMapDataEvent() {
			fmt.Printf("\n")
			fmt.Printf("==============================================================================\n")
			fmt.Printf("ebpf event: %v \n\n", e)

			bpfManager.PrintMapService()
			bpfManager.PrintMapNode()
			bpfManager.PrintMapBackend()
			bpfManager.PrintMapAffinity()
			bpfManager.PrintMapNatRecord()
		}
		fmt.Printf("ebpf event is closed : %v \n")
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)
	<-c
}

var _ = Describe("ebpf", func() {
	It("ebpf sanity test", func() {
		GinkgoWriter.Printf("\n------------------------------\n")
		GinkgoWriter.Printf("begin to work \n")
		Run()

	})

})

