package ebpfEvent

import (
	"fmt"
	"github.com/spidernet-io/rocktemplate/pkg/ebpf"
	"github.com/spidernet-io/rocktemplate/pkg/podBank"
	"github.com/spidernet-io/rocktemplate/pkg/types"
	"go.uber.org/zap"
	"time"
)

type ebpfEventStruct struct {
	l           *zap.Logger
	ebpfHandler ebpf.EbpfProgram
}

type EbpfEvent interface {
	WatchEbpfEvent(chan struct{})
}

var _ EbpfEvent = (*ebpfEventStruct)(nil)

func NewEbpfEvent(l *zap.Logger, ebpfHandler ebpf.EbpfProgram) EbpfEvent {
	return &ebpfEventStruct{
		l:           l,
		ebpfHandler: ebpfHandler,
	}
}

func (s *ebpfEventStruct) WatchEbpfEvent(stopWatch chan struct{}) {
	go func() {
		eventCh := s.ebpfHandler.GetMapDataEvent()

		for {
			select {
			case <-stopWatch:
				s.l.Sugar().Infof("stop watch ebpf event")
				break
			case event, ok := <-eventCh:
				if !ok {
					s.l.Sugar().Fatalf("ebpf event channel is closed")
				}

				s.l.Sugar().Debugf("received an ebpf event: %s ", event)
				var eventStr string

				podName, namespace, hostFlag, err := podBank.PodBankHander.LookupPodByPid(event.Pid)
				if err != nil {
					s.l.Sugar().Errorf("failed to get podName for pid %d: %v", event.Pid, err)
					// container application , but miss pod name
					eventStr += fmt.Sprintf("clientPodName=unknown, namespace=unknown, host=false, ")
				} else {
					if hostFlag {
						eventStr += fmt.Sprintf("ClientPodName=, namespace=, host=true, ")
					} else {
						eventStr += fmt.Sprintf("ClientPodName=%s, namespace=%s, host=false, ", podName, namespace)
					}
				}
				eventStr += fmt.Sprintf("NodeName=%d, ", types.AgentConfig.LocalNodeName)
				eventStr += fmt.Sprintf("IsIpv4=%d, IsSuccess=%d, ", event.IsIpv4, event.IsSuccess)
				eventStr += fmt.Sprintf("NatType=%s, NatMode=%s, ", ebpf.GetNatTypeStr(event.NatType), ebpf.GetNatModeStr(event.NatMode))
				if event.IsIpv4 != 0 {
					eventStr += fmt.Sprintf("DestIp=%s, DestPort=%d, NatIp=%s, NatPort=%d, ",
						ebpf.GetIpStr(event.OriginalDestV4Ip), event.OriginalDestPort, ebpf.GetIpStr(event.NatV4Ip), event.NatPort)
				} else {
					eventStr += fmt.Sprintf("DestIp=%s, DestPort=%d, NatIp=%s, NatPort=%d, ",
						ebpf.GetIpv6Str(event.OriginalDestV6ipHigh, event.OriginalDestV6ipLow), event.OriginalDestPort, ebpf.GetIpv6Str(event.NatV6ipHigh, event.NatV6ipLow), event.NatPort)
				}
				eventStr += fmt.Sprintf("Pid=%d, Failure=%s, ", event.Pid, ebpf.GetFailureStr(event.FailureCode))
				stamp := time.Now().UTC()
				eventStr += fmt.Sprintf("TimeStamp=%s ", stamp.Format("2006-01-02T15:04:05Z"))

				s.l.Sugar().Infof("ebpf event: %s", eventStr)
			}
		}
	}()

}
