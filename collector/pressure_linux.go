// Copyright 2018 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// +build !nopressure

package collector

import (
	"fmt"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"github.com/prometheus/procfs"
)

type pressureCollector struct {
	cpu     *prometheus.Desc
	io      *prometheus.Desc
	ioFull  *prometheus.Desc
	mem     *prometheus.Desc
	memFull *prometheus.Desc
}

func init() {
	registerCollector("pressure", defaultEnabled, NewPressureCollector)
}

func NewPressureCollector() (Collector, error) {
	return &pressureCollector{
		cpu: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "pressure", "wait_for_cpu_seconds_total"),
			"Total time in seconds that processes have waited for CPU time",
			nil, nil,
		),
		io: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "pressure", "wait_for_io_seconds_total"),
			"Total time in seconds that processes have waited due to IO congestion",
			nil, nil,
		),
		ioFull: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "pressure", "pause_for_io_seconds_total"),
			"Total time in seconds no process could make progress due to IO congestion",
			nil, nil,
		),
		mem: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "pressure", "wait_for_memory_seconds_total"),
			"Total time in seconds that processes have waited for memory",
			nil, nil,
		),
		memFull: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "pressure", "pause_for_memory_seconds_total"),
			"Total time in seconds no process could make progess due to memory congestion",
			nil, nil,
		),
	}, nil
}

func (c *pressureCollector) Update(ch chan<- prometheus.Metric) error {
	fs, err := procfs.NewFS(*procPath)
	if err != nil {
		return fmt.Errorf("failed to open procfs: %v", err)
	}

	mempressure, err := fs.NewResourcePressure("memory")
	if err != nil {
		if os.IsNotExist(err) {
			log.Debugf("could not find memory pressure file: %v", err)
			return nil
		}
		return err
	}
	iopressure, err := fs.NewResourcePressure("io")
	if err != nil {
		if os.IsNotExist(err) {
			log.Debugf("could not find io pressure file: %v", err)
			return nil
		}
		return err
	}
	cpupressure, err := fs.NewResourcePressure("cpu")
	if err != nil {
		if os.IsNotExist(err) {
			log.Debugf("could not find cpu pressure file: %v", err)
			return nil
		}
		return err
	}

	if cpupressure.Some != nil {
		ch <- prometheus.MustNewConstMetric(c.cpu, prometheus.CounterValue, float64(cpupressure.Some.TotalMicroseconds)/1000.0/1000.0)
	}
	if mempressure.Some != nil {
		ch <- prometheus.MustNewConstMetric(c.mem, prometheus.CounterValue, float64(mempressure.Some.TotalMicroseconds)/1000.0/1000.0)
	}
	if iopressure.Some != nil {
		ch <- prometheus.MustNewConstMetric(c.io, prometheus.CounterValue, float64(iopressure.Some.TotalMicroseconds)/1000.0/1000.0)
	}
	if mempressure.Full != nil {
		ch <- prometheus.MustNewConstMetric(c.memFull, prometheus.CounterValue, float64(mempressure.Full.TotalMicroseconds)/1000.0/1000.0)
	}
	if iopressure.Full != nil {
		ch <- prometheus.MustNewConstMetric(c.ioFull, prometheus.CounterValue, float64(iopressure.Full.TotalMicroseconds)/1000.0/1000.0)
	}

	return nil
}
