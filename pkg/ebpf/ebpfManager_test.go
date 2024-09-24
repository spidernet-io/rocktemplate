package main

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -no-strip -cc clang -target bpf -cflags "-D__TARGET_ARCH_x86"  bpf_cgroup   bpf/cgroup.c

import (
	"bytes"
	"encoding/binary"
	"fmt"
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
	bpfManager := NewEbpfProgramMananger()
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
		var keyList []bpf_cgroupMapkeyService
		var valueList []bpf_cgroupMapvalueService
		keyList = append(keyList, bpf_cgroupMapkeyService{
			Address: floatipLittle,
			Dport:   floatPortLittle,
			Proto:   PROTOCOL_TCP,
			NatType: NatTypeFloatIP,
			Scope:   0,
		})
		valueList = append(valueList, bpf_cgroupMapvalueService{
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
		var keyList []bpf_cgroupMapkeyBackend
		keyList = append(keyList, bpf_cgroupMapkeyBackend{
			BackendId: backendId,
			Order:     0,
		})
		keyList = append(keyList, bpf_cgroupMapkeyBackend{
			BackendId: backendId,
			Order:     1,
		})
		var valueList []bpf_cgroupMapvalueBackend
		valueList = append(valueList, bpf_cgroupMapvalueBackend{
			PodAddress:  pod1Little,
			PodPort:     podPortLittle,
			NodeAddress: node1Little,
			NodePort:    nodePortLittle,
			Proto:       PROTOCOL_TCP,
			Flags:       0,
		})
		valueList = append(valueList, bpf_cgroupMapvalueBackend{
			PodAddress:  pod2Little,
			PodPort:     podPortLittle,
			NodeAddress: node2Little,
			NodePort:    nodePortLittle,
			Proto:       PROTOCOL_TCP,
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
		var keyList []bpf_cgroupMapkeyNode
		keyList = append(keyList, bpf_cgroupMapkeyNode{
			Address: node1Little,
		})
		keyList = append(keyList, bpf_cgroupMapkeyNode{
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

func main() {
	fmt.Printf("\n")
	fmt.Printf("------------------------------\n")
	fmt.Printf("begin to work \n")

	bpfManager := NewEbpfProgramMananger()
	if err := bpfManager.LoadProgramp(); err != nil {
		fmt.Printf("failed to Load ebpf Programp: %v \n", err)
		os.Exit(1)
	}
	defer bpfManager.UnloadProgramp()

	fmt.Println("Run...")
	defer fmt.Println("Exiting...")

	Run()

}
