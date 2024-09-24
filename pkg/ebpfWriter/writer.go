// Copyright 2024 Authors of spidernet-io
// SPDX-License-Identifier: Apache-2.0

package ebpfWriter

import (
	"fmt"
	"github.com/spidernet-io/rocktemplate/pkg/lock"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	discovery "k8s.io/api/discovery/v1"
)

type EbpfWriter interface {
	UpdateService(svc *corev1.Service)
	UpdateEndpointSlice(*discovery.EndpointSlice)
	DeleteService(svc *corev1.Service)
	DeleteEndpointSlice(*discovery.EndpointSlice)
}

type EndpointData struct {
	svc *corev1.Service
	// one endpointslice store 100 endpoints by default
	// index: namesapce/name
	epsliceList map[string]*discovery.EndpointSlice
}

type ebpfWriter struct {
	l *lock.Mutex
	// index: namesapce/name
	data   map[string]EndpointData
	logger *zap.Logger
}

var _ EbpfWriter = (*ebpfWriter)(nil)

func NewEbpfWriter(logger *zap.Logger) {
	return &ebpfWriter{
		l:      &lock.Mutex{},
		logger: logger,
	}
}

func (s *ebpfWriter) UpdateService(svc *corev1.Service) error {

	if svc == nil {
		return fmt.Errorf("empty service")
	}

	index := svc.Namespace + "/" + svc.Name

	var oldSvc *corev1.Service
	var oldEdSliceList []*discovery.EndpointSlice
	shouldUpdateEbpf := false

	s.l.Lock()
	defer s.l.Unlock()
	if d, ok := s.data[index]; ok {
		s.logger.Sugar().Debugf("update service info to %s", index)
		if d.epsliceList != nil && len(d.epsliceList) > 0 {
			oldEdSliceList = d.epsliceList
			shouldUpdateEbpf = true
		}
		oldSvc = d
	} else {
		s.logger.Sugar().Debugf("add new service info to %s", index)
	}

	if shouldUpdateEbpf {
		// todo : generate a ebpf map data and apply it
		s.logger.Sugar().Infof("apply new data to ebpf map: %v", index)

		// todo: use the old data to generate ebpf data
		buildMapDataForService(oldSvc, oldEdSliceList)

		// todo: use the new data to generate ebpf data
		s.data[index] = svc
		buildMapDataForService(svc, oldEdSliceList)

		updateEbpfMapForService()
	}

	s.data[index] = svc
	return nil
}

func (s *ebpfWriter) DeleteService(svc *corev1.Service) error {

	if svc == nil {
		return fmt.Errorf("empty service")
	}

	index := svc.Namespace + "/" + svc.Name

	shouldUpdateEbpf := false
	s.l.Lock()
	defer s.l.Unlock()
	if d, ok := s.data[index]; ok {
		s.logger.Sugar().Debugf("delete service info to %s", index)
		shouldUpdateEbpf = true
	} else {
		s.logger.Sugar().Debugf("delete service info to %s", index)
	}

	if shouldUpdateEbpf {
		// todo : generate a ebpf map data and apply it
		s.logger.Sugar().Infof("apply new data to ebpf map: %v", index)
		buildMapDataForService(s.data[index].svc, s.data[index].epsliceList)
		deleteEbpfMapForService()
	}

	delete(s.data, index)
	return nil
}

func (s *ebpfWriter) UpdateUpdateEndpointSlice(epSlice *discovery.EndpointSlice) error {

	if epSlice == nil {
		return fmt.Errorf("empty EndpointSlice")
	}

	index := epSlice.Namespace + "/" + epSlice.OwnerReferences[0].Name
	epindex := epSlice.Namespace + "/" + epSlice.Name
	var oldSvc *corev1.Service

	s.l.Lock()
	defer s.l.Unlock()
	if d, ok := s.data[index]; ok {
		s.logger.Sugar().Debugf("update EndpointSlice info to %s", index)
		if d.svc != nil {
			oldSvc = d.svc
		}
	} else {
		s.logger.Sugar().Debugf("add new EndpointSlice info to %s", index)
	}

	if oldSvc != nil {
		// todo : generate a ebpf map data and apply it
		s.logger.Sugar().Infof("apply new data to ebpf map: %v", index)

		// todo: use the old data to generate ebpf data
		buildMapDataForService(oldSvc, s.data[index].epsliceList)

		// todo: use the new data to generate ebpf data
		s.data[index].epsliceList[epindex] = epSlice
		buildMapDataForService(oldSvc, s.data[index].epsliceList)

		updateEbpfMapForService()

	}

	s.data[index].epsliceList[epindex] = epSlice
	return nil
}

func (s *ebpfWriter) DeleteEndpointSlice(epSlice *discovery.EndpointSlice) error {

	if epSlice == nil {
		return fmt.Errorf("empty service")
	}

	index := epSlice.Namespace + "/" + epSlice.OwnerReferences[0].Name
	epindex := epSlice.Namespace + "/" + epSlice.Name

	shouldUpdateEbpf := false

	s.l.Lock()
	defer s.l.Unlock()

	if d, ok := s.data[index]; ok {
		s.logger.Sugar().Debugf("delete EndpointSlice info to %s", index)
		if d.svc == nil {
			shouldUpdateEbpf = false
		} else {
			shouldUpdateEbpf = false
		}
		delete(s.data[index].epsliceList, epindex)

	} else {
		s.logger.Sugar().Debugf("delete EndpointSlice info to %s", index)
		shouldUpdateEbpf = false
	}

	if shouldUpdateEbpf {
		// todo : generate a ebpf map data and apply it
		s.logger.Sugar().Infof("apply new data to ebpf map: %v", index)
		buildMapDataForService(s.data[index].svc, s.data[index].epsliceList)
		deleteEbpfMapForService()
	}

	return nil
}
