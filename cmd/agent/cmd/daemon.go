// Copyright 2022 Authors of spidernet-io
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/spidernet-io/rocktemplate/pkg/debug"
	"github.com/spidernet-io/rocktemplate/pkg/types"
)

func SetupUtility() {
	// run gops
	d := debug.New(rootLogger)
	if types.AgentConfig.GopsPort != 0 {
		d.RunGops(int(types.AgentConfig.GopsPort))
	}

	if types.AgentConfig.PyroscopeServerAddress != "" {
		d.RunPyroscope(types.AgentConfig.PyroscopeServerAddress, types.AgentConfig.PodName)
	}
}

func DaemonMain() {
	rootLogger.Sugar().Infof("config: %+v", types.AgentConfig)

	SetupUtility()

	// SetupHttpServer()

	// RunMetricsServer(types.AgentConfig.PodName)

	// SetupController()
	RunReconciles()

}
