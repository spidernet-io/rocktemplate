// Copyright 2022 Authors of spidernet-io
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"context"
	"github.com/spidernet-io/rocktemplate/pkg/debug"
	"github.com/spidernet-io/rocktemplate/pkg/mybookManager"
	"go.opentelemetry.io/otel/attribute"
	"path/filepath"
	"time"
)

func SetupUtility() {

	// run gops
	d := debug.New(rootLogger)
	if globalConfig.GopsPort != 0 {
		d.RunGops(int(globalConfig.GopsPort))
	}

	if globalConfig.PyroscopeServerAddress != "" {
		d.RunPyroscope(globalConfig.PyroscopeServerAddress, globalConfig.PodName)
	}
}

func DaemonMain() {

	rootLogger.Sugar().Infof("config: %+v", globalConfig)

	SetupUtility()

	SetupHttpServer()

	// ------

	RunMetricsServer("controller")
	if globalConfig.EnableMetric == true {
		MetricGaugeEndpoint.Add(context.Background(), 100)
		MetricGaugeEndpoint.Add(context.Background(), -10)
		MetricGaugeEndpoint.Add(context.Background(), 5)

		attrs := []attribute.KeyValue{
			attribute.Key("pod1").String("value1"),
		}
		MetricCounterRequest.Add(context.Background(), 10, attrs...)
		attrs = []attribute.KeyValue{
			attribute.Key("pod2").String("value1"),
		}
		MetricCounterRequest.Add(context.Background(), 5, attrs...)

		MetricHistogramDuration.Record(context.Background(), 10)
		MetricHistogramDuration.Record(context.Background(), 20)
	}

	// ----------
	s := mybookManager.New(rootLogger.Named("mybook"))
	s.RunInformer("testlease", globalConfig.PodNamespace, globalConfig.PodName)
	s.RunWebhookServer(int(globalConfig.WebhookPort), filepath.Dir(globalConfig.TlsServerCertPath))

	// ------------
	rootLogger.Info("hello world")
	time.Sleep(time.Hour)
}
