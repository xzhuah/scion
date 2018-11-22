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

package sibra_mgmt

import (
	"fmt"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/sibra"
	"github.com/scionproto/scion/go/proto"
)

var _ proto.Cerealizable = (*BandwidthExceeded)(nil)

type BandwidthExceeded struct {
	Id            	sibra.ID	`capnp:"id"`
	RawOriginIA 	addr.IAInt 	`capnp:"originIA"`
}

func (be *BandwidthExceeded) OriginIA() addr.IA {
	return be.RawOriginIA.IA()
}

func (be *BandwidthExceeded) ProtoId() proto.ProtoIdType {
	return proto.SibraBWExceeded_TypeID
}

func (be *BandwidthExceeded) String() string {
	return fmt.Sprintf("flow_id %s from AS %s",
		be.Id, be.OriginIA())
}
