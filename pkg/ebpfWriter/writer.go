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
	UpdateService(svc *corev1.Service) error
	UpdateEndpointSlice(*discovery.EndpointSlice) error
	DeleteService(svc *corev1.Service) error
	DeleteEndpointSlice(*discovery.EndpointSlice) error
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
	endpointData map[string]EndpointData
	logger       *zap.Logger
}

var _ EbpfWriter = (*ebpfWriter)(nil)

func NewEbpfWriter(logger *zap.Logger) EbpfWriter {
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
		if d.EpsliceList != nil && len(d.EpsliceList) > 0 {

			// todo : generate a ebpf map data and apply it
			s.logger.Sugar().Infof("apply new data to ebpf map for service %v", index)

			// todo: use the old data to generate ebpf data
			buildMapDataForService(d.Svc, d.EpsliceList)

			// todo: use the new data to generate ebpf data
			s.endpointData[index].Svc = svc
			buildMapDataForService(d.Svc, s.endpointData[index].EpsliceList)

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
	if _, ok := s.endpointData[index]; ok {
		// todo : generate a ebpf map data and apply it
		s.logger.Sugar().Infof("delete data from ebpf map for service: %v", index)
		buildMapDataForService(s.endpointData[index].Svc, s.endpointData[index].EpsliceList)
		deleteEbpfMapForService()
		delete(s.endpointData, index)
	} else {
		s.logger.Sugar().Debugf("no need to delete data from ebpf map for service %v", index)
	}

	return nil
}

func (s *ebpfWriter) UpdateEndpointSlice(epSlice *discovery.EndpointSlice) error {

	if epSlice == nil {
		return fmt.Errorf("empty EndpointSlice")
	}

	index := epSlice.Namespace + "/" + epSlice.OwnerReferences[0].Name
	epindex := epSlice.Namespace + "/" + epSlice.Name
	s.logger.Sugar().Debugf("update EndpointSlice %s for the service %s", epindex, index)

	s.l.Lock()
	defer s.l.Unlock()
	if d, ok := s.endpointData[index]; ok {
		if d.Svc != nil {
			// todo : generate a ebpf map data and apply it
			s.logger.Sugar().Infof("apply new data to ebpf map for the service %v", index)

			// todo: use the old data to generate ebpf data
			buildMapDataForService(s.endpointData[index].Svc, s.endpointData[index].EpsliceList)

			// todo: use the new data to generate ebpf data
			s.endpointData[index].EpsliceList[epindex] = epSlice
			buildMapDataForService(s.endpointData[index].Svc, s.endpointData[index].EpsliceList)

			updateEbpfMapForService()
		} else {
			s.endpointData[index].EpsliceList[epindex] = epSlice
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
		if d.Svc == nil {
			// when the service event happens, the data has been removed
			delete(s.endpointData[index].EpsliceList, epindex)
		} else {
			if _, ok := d.EpsliceList[epindex]; ok {
				s.logger.Sugar().Infof("apply new data to ebpf map for the service %v", index)

				// todo: use the old data to generate ebpf data
				buildMapDataForService(d.Svc, s.endpointData[index].EpsliceList)

				// todo: use the new data to generate ebpf data
				delete(s.endpointData[index].EpsliceList, epindex)
				buildMapDataForService(d.Svc, s.endpointData[index].EpsliceList)

				updateEbpfMapForService()

				goto finish
			}
		}
	}
	s.logger.Sugar().Debugf("no need to apply data for ebpf map  for the service %v", index)

finish:
	return nil
}
