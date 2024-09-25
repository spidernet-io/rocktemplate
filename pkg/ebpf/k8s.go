package ebpf

import (
	"encoding/binary"
	"fmt"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	discovery "k8s.io/api/discovery/v1"
	"reflect"
)

type serviceMapData struct {
	key *bpf_cgroupMapkeyService
	val *bpf_cgroupMapvalueService
}

type backendMapData struct {
	key *bpf_cgroupMapkeyBackend
	val *bpf_cgroupMapvalueBackend
}

func buildEbpfMapDataForV4ServiceTypeService(svc *corev1.Service, edsList map[string]*discovery.EndpointSlice) ([]*serviceMapData, []*backendMapData, error) {

	if svc == nil {
		return nil, nil, fmt.Errorf("service is empty")
	}

	resultSvcList := []*serviceMapData{}
	resultBackList := []*backendMapData{}

	svcV4Id := GenerateSvcV4Id(svc)
	affinityTime := GetServiceAffinityTime(svc)
	serviceFlags := GetServiceFlag(svc)

	for _, svcPort := range svc.Spec.Ports {

		protocol := GetPortProtocol(&svcPort)

		// generate data for backend map
		// 随着有多组 service port， 也有着多组的 backend
		localEp, remoteEp := ClassifyV4Endpoint(edsList)
		allEp := []*discovery.Endpoint{}
		allEp = append(allEp, localEp...)
		allEp = append(allEp, remoteEp...)
		for order, edp := range allEp {
			backMapKey := bpf_cgroupMapkeyBackend{
				Order:   uint32(order),
				SvcId:   svcV4Id,
				Dport:   uint16(svcPort.Port),
				Proto:   protocol,
				NatType: NAT_TYPE_SERVICE,
				Scope:   SCOPE_LOCAL_CLUSTER,
			}
			backMapVal := bpf_cgroupMapvalueBackend{
				PodAddress:  GetEndpointIPv4Address(edp),
				NodeAddress: 0,
				PodPort:     uint16(svcPort.TargetPort.IntValue()),
				NodePort:    uint16(svcPort.NodePort),
			}
			resultBackList = append(resultBackList, &backendMapData{
				key: &backMapKey,
				val: &backMapVal,
			})
		}

		// ----------------- generate data of service map  ----------------
		// get clusterIP, loadbalancerIP, externalIP
		// they use the same port, So deal with them together
		allVip := GetServiceV4AllVip(svc)
		for _, vip := range allVip {
			// generate data for service map
			svcMapKey := bpf_cgroupMapkeyService{
				Address: binary.LittleEndian.Uint32(vip.To4()),
				Dport:   uint16(svcPort.Port),
				Proto:   protocol,
				NatType: NAT_TYPE_SERVICE,
				Scope:   SCOPE_LOCAL_CLUSTER,
			}
			svcMapVal := bpf_cgroupMapvalueService{
				SvcId:             svcV4Id,
				TotalBackendCount: uint32(len(allEp)),
				LocalBackendCount: uint32(len(localEp)),
				AffinitySecond:    affinityTime,
				ServiceFlags:      serviceFlags,
				FloatipFlags:      0,
				RedirectFlags:     0,
			}
			resultSvcList = append(resultSvcList, &serviceMapData{
				key: &svcMapKey,
				val: &svcMapVal,
			})
		}

		// handle nodePort alone cause it uses nodePort
		if svcPort.NodePort != 0 {
			// generate data for service map
			svcMapKey := bpf_cgroupMapkeyService{
				Address: binary.LittleEndian.Uint32(NODEPORT_V4_IP),
				Dport:   uint16(svcPort.NodePort),
				Proto:   protocol,
				NatType: NAT_TYPE_SERVICE,
				Scope:   SCOPE_LOCAL_CLUSTER,
			}
			svcMapVal := bpf_cgroupMapvalueService{
				SvcId:             svcV4Id,
				TotalBackendCount: uint32(len(allEp)),
				LocalBackendCount: uint32(len(localEp)),
				AffinitySecond:    affinityTime,
				ServiceFlags:      serviceFlags,
				FloatipFlags:      0,
				RedirectFlags:     0,
			}
			resultSvcList = append(resultSvcList, &serviceMapData{
				key: &svcMapKey,
				val: &svcMapVal,
			})
		}
	}
	return resultSvcList, resultBackList, nil

}

func buildEbpfMapDataForV4Service(natType uint8, svc *corev1.Service, edsList map[string]*discovery.EndpointSlice) ([]*serviceMapData, []*backendMapData, error) {
	if svc == nil {
		return nil, nil, fmt.Errorf("failed to buildEbpfMapDataForV4Service, service is nil")
	}

	if natType == NAT_TYPE_SERVICE {
		return buildEbpfMapDataForV4ServiceTypeService(svc, edsList)
	} else if natType == NAT_TYPE_REDIRECT {
		return nil, nil, fmt.Errorf("buildEbpfMapDataForV4Service: unimplemented NAT_TYPE_REDIRECT")
	} else if natType == NAT_TYPE_BALANCING {
		return nil, nil, fmt.Errorf("buildEbpfMapDataForV4Service: unimplemented NAT_TYPE_BALANCING")
	}
	return nil, nil, fmt.Errorf("buildEbpfMapDataForV4Service: unknowd nat type %d", natType)
}

func (s *EbpfProgramStruct) applyEpfMapDataService(l *zap.Logger, oldList, newList []*serviceMapData) error {

	delKeyList := []bpf_cgroupMapkeyService{}
	addKeyList := []bpf_cgroupMapkeyService{}
	addValList := []bpf_cgroupMapvalueService{}

	l.Sugar().Debugf("service map %d items in oldList: \n", len(oldList))
	for k, v := range oldList {
		l.Sugar().Debugf("service map oldList[%d]: key=%s, value=%s \n", k, *v.key, *v.val)
	}
	l.Sugar().Debugf("service map %d items in newSvcList: \n", len(newList))
	for k, v := range newList {
		l.Sugar().Debugf("service map newSvcList[%d]: key=%s, value=%s \n", k, *v.key, *v.val)
	}

OUTER_OLD:
	for _, oldKey := range oldList {
		shouldDelOld := true
		for _, newKey := range newList {
			if reflect.DeepEqual(oldKey.key, newKey.key) {
				if !reflect.DeepEqual(oldKey.val, newKey.val) {
					addKeyList = append(addKeyList, *newKey.key)
					addValList = append(addValList, *newKey.val)
					l.Sugar().Infof("update new service data: key=%s , value=%s \n", newKey.key, newKey.val)
					continue OUTER_OLD
				}
			}
		}
		if shouldDelOld {
			l.Sugar().Infof("delete service data: key=%s , value=%s \n", oldKey.key, oldKey.val)
			delKeyList = append(delKeyList, *oldKey.key)
		}
	}

OUTER_NEW:
	for _, newKey := range newList {
		shouldAddNew := true
		for _, oldKey := range oldList {
			if reflect.DeepEqual(oldKey.key, newKey.key) {
				continue OUTER_NEW
			}
		}
		if shouldAddNew {
			addKeyList = append(addKeyList, *newKey.key)
			addValList = append(addValList, *newKey.val)
			l.Sugar().Infof("create new service data: key=%s , value=%s \n", newKey.key, newKey.val)
		}
	}

	// -------- apply
	if len(addKeyList) > 0 {
		if e := s.UpdateMapService(addKeyList, addValList); e != nil {
			return fmt.Errorf("failed to update service map: %v", e)
		}
		l.Sugar().Infof("succeeded to update %d items in service map: \n", len(addKeyList))
		for k, _ := range addKeyList {
			l.Sugar().Infof("succeeded to update service map: key=%s, value=%s \n", addKeyList[k], addValList[k])
		}
	}
	if len(delKeyList) > 0 {
		if e := s.DeleteMapService(delKeyList); e != nil {
			return fmt.Errorf("failed to delete service map: %v", e)
		}
		l.Sugar().Infof("succeeded to delete %d items in service data \n", len(delKeyList))
		for k, _ := range delKeyList {
			l.Sugar().Infof("succeeded to delete service map: key=%s \n", delKeyList[k])
		}
	}
	return nil
}

func (s *EbpfProgramStruct) applyEpfMapDataBackend(l *zap.Logger, oldList, newList []*backendMapData) error {

	delKeyList := []bpf_cgroupMapkeyBackend{}
	addKeyList := []bpf_cgroupMapkeyBackend{}
	addValList := []bpf_cgroupMapvalueBackend{}

	l.Sugar().Debugf("backend map %d items in oldList: \n", len(oldList))
	for k, v := range oldList {
		l.Sugar().Debugf("backend map oldList[%d]: key=%s, value=%s ", k, *v.key, *v.val)
	}
	l.Sugar().Debugf("backend map %d items in newSvcList: \n", len(newList))
	for k, v := range newList {
		l.Sugar().Debugf("backend map newSvcList[%d]: key=%s, value=%s ", k, *v.key, *v.val)
	}

OUTER_OLD:
	for _, oldKey := range oldList {
		shouldDelOld := true
		for _, newKey := range newList {
			if reflect.DeepEqual(oldKey.key, newKey.key) {
				if !reflect.DeepEqual(oldKey.val, newKey.val) {
					addKeyList = append(addKeyList, *newKey.key)
					addValList = append(addValList, *newKey.val)
					l.Sugar().Infof("update new backend data: key=%s , value=%s \n", newKey.key, newKey.val)
					continue OUTER_OLD
				}
			}
		}
		if shouldDelOld {
			l.Sugar().Infof("delete backend data: key=%s , value=%s \n", oldKey.key, oldKey.val)
			delKeyList = append(delKeyList, *oldKey.key)
		}
	}

OUTER_NEW:
	for _, newKey := range newList {
		shouldAddNew := true
		for _, oldKey := range oldList {
			if reflect.DeepEqual(oldKey.key, newKey.key) {
				continue OUTER_NEW
			}
		}
		if shouldAddNew {
			addKeyList = append(addKeyList, *newKey.key)
			addValList = append(addValList, *newKey.val)
			l.Sugar().Infof("create new backend data: key=%s , value=%s \n", newKey.key, newKey.val)
		}
	}

	// -------- apply
	if len(addKeyList) > 0 {
		if e := s.UpdateMapBackend(addKeyList, addValList); e != nil {
			return fmt.Errorf("failed to update backend map: %v", e)
		}
		l.Sugar().Infof("succeeded to update %d items in backend map: \n", len(addKeyList))
		for k, _ := range addKeyList {
			l.Sugar().Infof("succeeded to update backend map: key=%s, value=%s \n", addKeyList[k], addValList[k])
		}
	}
	if len(delKeyList) > 0 {
		if e := s.DeleteMapBackend(delKeyList); e != nil {
			return fmt.Errorf("failed to delete backend map: %v", e)
		}
		l.Sugar().Infof("succeeded to delete %d items in backend data \n", len(delKeyList))
		for k, _ := range delKeyList {
			l.Sugar().Infof("succeeded to delete backend map: key=%s \n", delKeyList[k])
		}
	}
	return nil
}

// -------------------------------------------------- for k8s service

func (s *EbpfProgramStruct) UpdateEbpfMapForService(l *zap.Logger, oldSvc, newSvc *corev1.Service, oldEdsList, newEdsList map[string]*discovery.EndpointSlice) error {

	oldSvcList, oldBkList, err1 := buildEbpfMapDataForV4Service(NAT_TYPE_SERVICE, oldSvc, oldEdsList)
	if err1 != nil {
		return fmt.Errorf("failed to buildEbpfMapDataForV4Service: %v", err1)
	}
	newSvcList, newBkList, err2 := buildEbpfMapDataForV4Service(NAT_TYPE_SERVICE, newSvc, newEdsList)
	if err2 != nil {
		return fmt.Errorf("failed to buildEbpfMapDataForV4Service: %v", err2)
	}

	if e := s.applyEpfMapDataService(l, oldSvcList, newSvcList); e != nil {
		return fmt.Errorf("failed to applyEpfMapDataService: %v", e)
	}
	if e := s.applyEpfMapDataBackend(l, oldBkList, newBkList); e != nil {
		return fmt.Errorf("failed to applyEpfMapDataBackend: %v", e)
	}
	return nil
}

func (s *EbpfProgramStruct) DeleteEbpfMapForService(l *zap.Logger, svc *corev1.Service, edsList map[string]*discovery.EndpointSlice) error {
	svcList, bkList, err := buildEbpfMapDataForV4Service(NAT_TYPE_SERVICE, svc, edsList)
	if err != nil {
		return fmt.Errorf("failed to buildEbpfMapDataForV4Service: %v", err)
	}

	if e := s.applyEpfMapDataService(l, svcList, []*serviceMapData{}); e != nil {
		return fmt.Errorf("failed to applyEpfMapDataService: %v", e)
	}
	if e := s.applyEpfMapDataBackend(l, bkList, []*backendMapData{}); e != nil {
		return fmt.Errorf("failed to applyEpfMapDataBackend: %v", e)
	}

	return nil
}

// -------------------------------------------------- for local redirect

// -------------------------------------------------- for custom balancing
