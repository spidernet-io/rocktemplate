package ebpf

import (
	corev1 "k8s.io/api/core/v1"
	discovery "k8s.io/api/discovery/v1"
)

func (s *EbpfProgramStruct) UpdateEbpfMapForService(oldSvc, newSvc *corev1.Service, oldEdsList, newEdsList map[string]*discovery.EndpointSlice) error {

	return nil
}

func (s *EbpfProgramStruct) DeleteEbpfMapForService(svc *corev1.Service) error {

	return nil
}
