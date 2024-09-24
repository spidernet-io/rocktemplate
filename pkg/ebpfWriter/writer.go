// Copyright 2024 Authors of spidernet-io
// SPDX-License-Identifier: Apache-2.0

package ebpfWriter

import (
	"fmt"
	"github.com/spidernet-io/rocktemplate/pkg/k8s"
	"github.com/spidernet-io/rocktemplate/pkg/lock"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	discovery "k8s.io/api/discovery/v1"
	"time"
)

type EbpfWriter interface {
	UpdateService(*zap.Logger, *corev1.Service, bool) error
	UpdateEndpointSlice(*zap.Logger, *discovery.EndpointSlice, bool) error
	DeleteService(*zap.Logger, *corev1.Service) error
	DeleteEndpointSlice(*zap.Logger, *discovery.EndpointSlice) error
}

type EndpointData struct {
	Svc *corev1.Service
	// one endpointslice store 100 endpoints by default
	// index: namesapce/name
	EpsliceList map[string]*discovery.EndpointSlice
}

type ebpfWriter struct {
	l *lock.Mutex
	// index: namesapce/name
	endpointData map[string]*EndpointData
	// use the creationTimestamp to record the last update time, and calculate the validityTime
	validityTime time.Duration
	log          *zap.Logger
}

var _ EbpfWriter = (*ebpfWriter)(nil)

func NewEbpfWriter(validityTime time.Duration, l *zap.Logger) EbpfWriter {
	t := ebpfWriter{
		l:            &lock.Mutex{},
		endpointData: make(map[string]*EndpointData),
		validityTime: validityTime,
		log:          l,
	}
	go t.DeamonGC()
	return &t
}

func (s *ebpfWriter) DeamonGC() {
	// todo: delete ebpf map data according the metadata.CreationTimestamp by the validityTime
	logger := s.log
	logger("ebpfWriter DeamonGC begin to retrieve ebpf data with validityTime %v", s.validityTime)
	for {
		time.Sleep(time.Hour)
	}
}

func (s *ebpfWriter) UpdateService(l *zap.Logger, svc *corev1.Service, onlyUpdateTime bool) error {

	if svc == nil {
		return fmt.Errorf("empty service")
	}

	// use it to record last update time
	svc.ObjectMeta.CreationTimestamp = time.Now()

	index := svc.Namespace + "/" + svc.Name
	l.Sugar().Debugf("update the service %s", index)

	s.l.Lock()
	defer s.l.Unlock()
	if d, ok := s.endpointData[index]; ok {
		if d.EpsliceList != nil && len(d.EpsliceList) > 0 {
			if !onlyUpdateTime {
				l.Sugar().Infof("apply new data to ebpf map for service %v", index)
				// todo: use the old data to generate ebpf data
				buildMapDataForService(d.Svc, d.EpsliceList)
				// todo: use the new data to generate ebpf data
				d.Svc = svc
				buildMapDataForService(d.Svc, d.EpsliceList)
				// write to ebpf map
				updateEbpfMapForService()
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
		buildMapDataForService(d.Svc, d.EpsliceList)
		deleteEbpfMapForService()
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
	epSlice.ObjectMeta.CreationTimestamp = time.Now()

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
				// todo: use the old data to generate ebpf data
				buildMapDataForService(s.endpointData[index].Svc, s.endpointData[index].EpsliceList)
				// todo: use the new data to generate ebpf data
				d.EpsliceList[epindex] = epSlice
				buildMapDataForService(s.endpointData[index].Svc, s.endpointData[index].EpsliceList)
				// write to ebpf
				updateEbpfMapForService()
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
				// todo: use the old data to generate ebpf data
				buildMapDataForService(d.Svc, d.EpsliceList)
				// todo: use the new data to generate ebpf data
				delete(d.EpsliceList, epindex)
				buildMapDataForService(d.Svc, d.EpsliceList)
				// write to ebpf
				updateEbpfMapForService()
				goto finish
			}
		}
	}
	l.Sugar().Debugf("no need to apply data for ebpf map, cause the data has been already removed")

finish:
	return nil
}
