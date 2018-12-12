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

package dist

import (
	"github.com/scionproto/scion/go/lib/ctrl/sibra_mgmt"
	"github.com/scionproto/scion/go/lib/infra"
	"github.com/scionproto/scion/go/lib/log"
)

var _ infra.Handler = (*BWExceededReqHandler)(nil)

type BWExceededReqHandler struct{}

func (h *BWExceededReqHandler) Handle(r *infra.Request) {
	//conf:=conf.Get()
	pld := r.Message.(*sibra_mgmt.BandwidthExceeded)
	log.Debug("Blacklisted flow", "sibra_id", pld.Id)
	//conf.SibraAlgo.Blacklist(pld.RawOriginIA.IA(),time.Duration(1*time.Hour))
}
