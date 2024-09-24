package ebpf

import (
	"encoding/binary"
	"fmt"
	"net"
)

func GetProtocolStr(p uint8) string {
	proto := "unknown"
	switch p {
	case PROTOCOL_TCP:
		proto = "tcp"
	case PROTOCOL_UDP:
		proto = "udp"
	default:
	}
	return proto
}

func GetIpStr(p uint32) string {
	tmp := make([]byte, 4)
	binary.LittleEndian.PutUint32(tmp, p)
	return net.IP(tmp).String()
}

func GetNatTypeStr(p uint8) string {
	natType := "unknown"
	switch p {
	case NatTypeService:
		natType = "service"
	case NatTypeLocalRedirect:
		natType = "localRedirect"
	case NatTypeFloatIP:
		natType = "floatIP"
	default:
	}
	return natType
}

func (t bpf_cgroupMapkeyService) String() string {
	return fmt.Sprintf(`{ DestIp:"%s", DestPort:%d, protocol:"%s", NatType:"%s", Scope:%d }`,
		GetIpStr(t.Address), t.Dport, GetProtocolStr(t.Proto), GetNatTypeStr(t.NatType), t.Scope)
}

// ------------------------------------------------

func (t bpf_cgroupMapkeyNode) String() string {
	return fmt.Sprintf(`{ NodeIp:"%s"}`, GetIpStr(t.Address))
}

// ------------------------------------------------

func (t bpf_cgroupMapvalueBackend) String() string {
	return fmt.Sprintf(`{ PodIp:"%s" , PodPort:%d, NodeIp:"%s", NodePort:%d, protocol:"%s", Flags:%d }`,
		GetIpStr(t.PodAddress), t.PodPort, GetIpStr(t.NodeAddress), t.NodePort, GetProtocolStr(t.Proto), t.Flags)
}

// ------------------------------------------------

func (t bpf_cgroupMapkeyNatRecord) String() string {
	return fmt.Sprintf(`{ SocketCookie:%d, NatIp:"%s", NatPort:%d, protocol:"%s" }`,
		t.SocketCookie, GetIpStr(t.NatIp), t.NatPort, GetProtocolStr(t.Proto))
}

func (t bpf_cgroupMapvalueNatRecord) String() string {
	return fmt.Sprintf(`{ OriginalDstIp:"%s" , OriginalDstPort:%d }`,
		GetIpStr(t.OriginalDestIp), t.OriginalDestPort)
}

// -------------------------------------------------

// struct for ebpf map : event
type MapEventValue struct {
	NatIp            uint32
	OriginalDestIp   uint32
	NatPort          uint16
	OriginalDestPort uint16
	Tgid             uint32
	IsIpv4           uint8 /* 0 for ipv6 data, 1 for ipv4 data */
	IsSuccess        uint8 /* 1 for success , 0 for failure */
	NatType          uint8 /* 1 for NAT_TYPE_FLOATIP , 2 for NAT_TYPE_SVC, 3 for NAT_TYPE_REDIRECT  */
	Pad              uint8
}

func (t MapEventValue) String() string {
	return fmt.Sprintf(`{ IsIpv4:%d, IsSuccess:%d, NatType:"%s", OriginalDestIp:"%s", OriginalDestPort:%d, NatIp:"%s", NatPort:%d , Tgid:%d }`,
		t.IsIpv4, t.IsSuccess, GetNatTypeStr(t.NatType),
		GetIpStr(t.OriginalDestIp), t.OriginalDestPort, GetIpStr(t.NatIp), t.NatPort, t.Tgid)
}
