// Copyright 2018 ETH Zurich
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

package main

import (
	"github.com/scionproto/scion/go/border/rpkt"
	"github.com/scionproto/scion/go/lib/ctrl/sibra_mgmt"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/sibra/flowmonitor"
)

// Used to verify if flow is exceeding reserved bandwidth
// Returns true if flow has exceeded bandwidth and packet should be dropped
func (r *Router) FlowMonitoringCallback(info flowmonitor.FlowInfo, isLocalFlow bool) bool {
	var result flowmonitor.FlowMonitoringResult = flowmonitor.BANDWIDTH_OK
	if isLocalFlow {
		result = r.localFlowMonitor.IsFlowRateExceeded(&info, true)
	}else{
		result = r.transitFlowMonitor.IsFlowRateExceeded(&info)
	}

	if result==flowmonitor.BANDWIDTH_BLACKLIST {
		log.Info("Blacklisting flow", "sibra_id", info.ReservationId)
		r.SIBRACallback(rpkt.SIBRAInternalPacket{
			Payload:&sibra_mgmt.BandwidthExceeded{
				Id:info.ReservationId,
				RawOriginIA:info.SourceIA.IAInt(),
			},
		})
		return true
	}else if result == flowmonitor.BANDWIDTH_EXCEEDED {
		return true
	}else{
		return false
	}
}
