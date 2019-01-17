// Copyright 2016 ETH Zurich
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package metrics defines and exports sibra service metrics to be
// scraped by prometheus.
package metrics

import (
	"flag"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/ringbuf"
	"net"
	"net/http"

	"github.com/scionproto/scion/go/lib/prom"
)

const (
	NAMESPACE = "colibri_srv"
	EPHEMERAL_RES_USAGE = "eph_path_bandwidth_res"
	STEADY_RES_USAGE = "steady_path_bandwidth"
)

var PrometheusAddr = flag.String("prom", "127.0.0.1:1283", "Address to export prometheus metrics on")

var (
	// Represents cumulative bandwidth for steady paths created to certain destination AS
	SteadyPathsBandwidth	*prometheus.GaugeVec
	// Represents total bandwidth reserved for eph paths for a given steady path
	EphBandwidthRsrvd		*prometheus.GaugeVec
)

// Ensure all metrics are registered.
func Init(elem string) {
	log.Debug("Initializing metrics", "service_id", elem)
	constLabels := prometheus.Labels{"elem": elem}

	newGVec := func(name, help string, lNames []string) *prometheus.GaugeVec {
		v := prom.NewGaugeVec(NAMESPACE, "", name, help, constLabels, lNames)
		prometheus.MustRegister(v)
		return v
	}

	SteadyPathsBandwidth = newGVec(
		STEADY_RES_USAGE,
		"Total amount of steady path bandwidth between ASes", []string{"dstAs", "type"})

	EphBandwidthRsrvd = newGVec(
		EPHEMERAL_RES_USAGE,
		"Total reserved bandwidth between given ASes", []string{"steadyRes"})

	// Initialize ringbuf metrics.
	ringbuf.InitMetrics(NAMESPACE, constLabels, []string{"ringId"})

	http.Handle("/metrics", promhttp.Handler())
}

// Start handles exposing prometheus metrics.
func Start() error {
	ln, err := net.Listen("tcp", *PrometheusAddr)
	if err != nil {
		return common.NewBasicError("Unable to bind prometheus metrics port", err)
	}
	log.Info("Exporting prometheus metrics", "addr", *PrometheusAddr)
	go http.Serve(ln, nil)
	return nil
}