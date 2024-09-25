// Copyright 2024 Authors of spidernet-io
// SPDX-License-Identifier: Apache-2.0

package ebpfWriter

import (
	"github.com/spidernet-io/rocktemplate/pkg/ebpf"
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
	ebpfhandler  ebpf.EbpfProgram
}

var _ EbpfWriter = (*ebpfWriter)(nil)

func NewEbpfWriter(ebpfhandler ebpf.EbpfProgram, validityTime time.Duration, l *zap.Logger) EbpfWriter {
	t := ebpfWriter{
		l:            &lock.Mutex{},
		endpointData: make(map[string]*EndpointData),
		validityTime: validityTime,
		log:          l,
		ebpfhandler:  ebpfhandler,
	}
	go t.DeamonGC()
	return &t
}

func (s *ebpfWriter) DeamonGC() {
	// todo: delete ebpf map data according the metadata.CreationTimestamp by the validityTime
	logger := s.log
	logger.Sugar().Infof("ebpfWriter DeamonGC begin to retrieve ebpf data with validityTime %s", s.validityTime.String())
	for {
		time.Sleep(time.Hour)
	}
}
