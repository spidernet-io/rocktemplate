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
	endpointData map[string]EndpointData
	logger       *zap.Logger
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
	s.logger.Sugar().Debugf("delete the service %s", index)

	s.l.Lock()
	defer s.l.Unlock()
	if d, ok := s.endpointData[index]; ok {
		if d.epsliceList != nil && len(d.epsliceList) > 0 {

			// todo : generate a ebpf map data and apply it
			s.logger.Sugar().Infof("apply new data to ebpf map for service %v", index)

			// todo: use the old data to generate ebpf data
			buildMapDataForService(s.endpointData[index].svc, d.epsliceList)

			// todo: use the new data to generate ebpf data
			s.endpointData[index].svc = svc
			buildMapDataForService(s.endpointData[index].svc, d.epsliceList)

			updateEbpfMapForService()
		}
	} else {
		s.logger.Sugar().Debugf("no need to apply new data to ebpf map for service %v", index)
	}

	return nil
}

func (s *ebpfWriter) DeleteService(svc *corev1.Service) error {
	if svc == nil {
		return fmt.Errorf("empty service")
	}

	index := svc.Namespace + "/" + svc.Name
	s.logger.Sugar().Debugf("delete service %s", index)

	s.l.Lock()
	defer s.l.Unlock()
	if d, ok := s.endpointData[index]; ok {
		// todo : generate a ebpf map data and apply it
		s.logger.Sugar().Infof("delete data from ebpf map for service: %v", index)
		buildMapDataForService(s.endpointData[index].svc, s.endpointData[index].epsliceList)
		deleteEbpfMapForService()
		delete(s.endpointData, index)
	} else {
		s.logger.Sugar().Debugf("no need to delete data from ebpf map for service %v", index)
	}

	return nil
}

func (s *ebpfWriter) UpdateUpdateEndpointSlice(epSlice *discovery.EndpointSlice) error {

	if epSlice == nil {
		return fmt.Errorf("empty EndpointSlice")
	}

	index := epSlice.Namespace + "/" + epSlice.OwnerReferences[0].Name
	epindex := epSlice.Namespace + "/" + epSlice.Name
	s.logger.Sugar().Debugf("update EndpointSlice %s for the service %s", epindex, index)

	s.l.Lock()
	defer s.l.Unlock()
	if d, ok := s.endpointData[index]; ok {
		if d.svc != nil {
			// todo : generate a ebpf map data and apply it
			s.logger.Sugar().Infof("apply new data to ebpf map for the service %v", index)

			// todo: use the old data to generate ebpf data
			buildMapDataForService(s.endpointData[index].svc, s.endpointData[index].epsliceList)

			// todo: use the new data to generate ebpf data
			s.endpointData[index].epsliceList[epindex] = epSlice
			buildMapDataForService(s.endpointData[index].svc, s.endpointData[index].epsliceList)

			updateEbpfMapForService()
		} else {
			s.endpointData[index].epsliceList[epindex] = epSlice
		}
	} else {
		s.logger.Sugar().Debugf("no need to apply new data to ebpf map for the service %v", index)
	}

	return nil
}

func (s *ebpfWriter) DeleteEndpointSlice(epSlice *discovery.EndpointSlice) error {

	if epSlice == nil {
		return fmt.Errorf("empty service")
	}

	index := epSlice.Namespace + "/" + epSlice.OwnerReferences[0].Name
	epindex := epSlice.Namespace + "/" + epSlice.Name
	s.logger.Sugar().Debugf("delete EndpointSlice %s for the service %s", epindex, index)

	s.l.Lock()
	defer s.l.Unlock()
	if d, ok := s.endpointData[index]; ok {
		if d.svc == nil {
			// when the service event happens, the data has been removed
			delete(s.endpointData[index].epsliceList, epindex)
		} else {
			if t, ok := d[epindex]; ok {
				s.logger.Sugar().Infof("apply new data to ebpf map for the service %v", index)

				// todo: use the old data to generate ebpf data
				buildMapDataForService(d.svc, s.endpointData[index].epsliceList)

				// todo: use the new data to generate ebpf data
				delete(s.endpointData[index].epsliceList, epindex)
				buildMapDataForService(d.svc, s.endpointData[index].epsliceList)

				updateEbpfMapForService()

				goto finish
			}
		}
	}
	s.logger.Sugar().Debugf("no need to apply data for ebpf map  for the service %v", index)

finish:
	return nil
}
