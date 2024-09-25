// Copyright 2024 Authors of spidernet-io
// SPDX-License-Identifier: Apache-2.0

package ebpfWriter

import (
	"fmt"
	"github.com/spidernet-io/rocktemplate/pkg/k8s"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	discovery "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

func shallowCopy(t map[string]*discovery.EndpointSlice) map[string]*discovery.EndpointSlice {
	m := make(map[string]*discovery.EndpointSlice)
	for k, v := range t {
		m[k] = v
	}
	return m
}

func (s *ebpfWriter) UpdateService(l *zap.Logger, svc *corev1.Service, onlyUpdateTime bool) error {

	if svc == nil {
		return fmt.Errorf("empty service")
	}

	// use it to record last update time
	svc.ObjectMeta.CreationTimestamp = metav1.Time{
		time.Now(),
	}

	index := svc.Namespace + "/" + svc.Name
	l.Sugar().Debugf("update the service %s", index)

	s.l.Lock()
	defer s.l.Unlock()
	if d, ok := s.endpointData[index]; ok {
		if d.EpsliceList != nil && len(d.EpsliceList) > 0 {
			if !onlyUpdateTime {
				l.Sugar().Infof("apply new data to ebpf map for service %v", index)
				s.ebpfhandler.UpdateEbpfMapForService(l, d.Svc, svc, d.EpsliceList, d.EpsliceList)
				d.Svc = svc
			} else {
				l.Sugar().Debugf("just update lastUpdateTime")
				d.Svc = svc
			}
		} else {
			l.Sugar().Debugf("no need to apply new data to ebpf map, cause miss endpointslice")
			d.Svc = svc
		}
	} else {
		l.Sugar().Debugf("no need to apply new data to ebpf map, cause miss endpointslice")
		s.endpointData[index] = &EndpointData{
			Svc:         svc,
			EpsliceList: make(map[string]*discovery.EndpointSlice),
		}
	}

	return nil
}

func (s *ebpfWriter) DeleteService(l *zap.Logger, svc *corev1.Service) error {
	if svc == nil {
		return fmt.Errorf("empty service")
	}

	index := svc.Namespace + "/" + svc.Name
	l.Sugar().Debugf("delete service %s", index)

	s.l.Lock()
	defer s.l.Unlock()
	if d, ok := s.endpointData[index]; ok {
		// todo : generate a ebpf map data and apply it
		l.Sugar().Infof("delete data from ebpf map for service: %v", index)
		s.ebpfhandler.DeleteEbpfMapForService(l, d.Svc, d.EpsliceList)
		delete(s.endpointData, index)
	} else {
		l.Sugar().Debugf("no need to delete data from ebpf map, cause already removed")
	}

	return nil
}

func (s *ebpfWriter) UpdateEndpointSlice(l *zap.Logger, epSlice *discovery.EndpointSlice, onlyUpdateTime bool) error {

	if epSlice == nil {
		return fmt.Errorf("empty EndpointSlice")
	}
	epSlice.ObjectMeta.CreationTimestamp = metav1.Time{
		time.Now(),
	}

	// for default/kubernetes ，there is no owner
	index := k8s.GetEndpointSliceOwnerName(epSlice)
	epindex := epSlice.Namespace + "/" + epSlice.Name
	l.Sugar().Debugf("update EndpointSlice %s for the service %s", epindex, index)

	s.l.Lock()
	defer s.l.Unlock()
	if d, ok := s.endpointData[index]; ok {
		if d.Svc != nil {
			if !onlyUpdateTime {
				l.Sugar().Infof("apply new data to ebpf map for the service %v", index)
				oldEps := shallowCopy(d.EpsliceList)
				d.EpsliceList[epindex] = epSlice
				s.ebpfhandler.UpdateEbpfMapForService(l, d.Svc, d.Svc, oldEps, d.EpsliceList)
			} else {
				l.Sugar().Debugf("just update lastUpdateTime")
				d.EpsliceList[epindex] = epSlice
			}
		} else {
			d.EpsliceList[epindex] = epSlice
			l.Sugar().Debugf("no need to apply new data to ebpf map, cause miss service")
		}
	} else {
		l.Sugar().Debugf("no need to apply new data to ebpf map, cause miss service")
		s.endpointData[index] = &EndpointData{
			Svc: nil,
			EpsliceList: map[string]*discovery.EndpointSlice{
				epindex: epSlice,
			},
		}
	}

	return nil
}

func (s *ebpfWriter) DeleteEndpointSlice(l *zap.Logger, epSlice *discovery.EndpointSlice) error {

	if epSlice == nil {
		return fmt.Errorf("empty service")
	}

	index := k8s.GetEndpointSliceOwnerName(epSlice)
	epindex := epSlice.Namespace + "/" + epSlice.Name
	l.Sugar().Debugf("delete EndpointSlice %s for the service %s", epindex, index)

	s.l.Lock()
	defer s.l.Unlock()
	if d, ok := s.endpointData[index]; ok {
		if d.Svc == nil {
			// when the service event happens, the data has been removed
			delete(d.EpsliceList, epindex)
		} else {
			if _, ok := d.EpsliceList[epindex]; ok {
				l.Sugar().Infof("apply new data to ebpf map for the service %v", index)
				oldEps := shallowCopy(d.EpsliceList)
				delete(d.EpsliceList, epindex)
				s.ebpfhandler.UpdateEbpfMapForService(l, d.Svc, d.Svc, oldEps, d.EpsliceList)

				goto finish
			}
		}
	}
	l.Sugar().Debugf("no need to apply data for ebpf map, cause the data has been already removed")

finish:
	return nil
}