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

package drkey_mgmt

import (
	"fmt"
	"strings"

	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/proto"
)

type union struct {
	Which   proto.DRKeyMgmt_Which
	Lvl1Req *Lvl1Req `capnp:"drkeyLvl1Req"`
	Lvl1Rep *Lvl1Rep `capnp:"drkeyLvl1Rep"`
	Lvl2Req *Lvl2Req `capnp:"drkeyLvl2Req"`
	Lvl2Rep *Lvl2Rep `capnp:"drkeyLvl2Rep"`
}

func (u *union) set(c proto.Cerealizable) error {
	switch p := c.(type) {
	case *Lvl1Req:
		u.Which = proto.DRKeyMgmt_Which_drkeyLvl1Req
		u.Lvl1Req = p
	case *Lvl1Rep:
		u.Which = proto.DRKeyMgmt_Which_drkeyLvl1Rep
		u.Lvl1Rep = p
	case *Lvl2Req:
		u.Which = proto.DRKeyMgmt_Which_drkeyLvl2Req
		u.Lvl2Req = p
	case *Lvl2Rep:
		u.Which = proto.DRKeyMgmt_Which_drkeyLvl2Rep
		u.Lvl2Rep = p
	default:
		return common.NewBasicError("Unsupported drkey mgmt union type (set)",
			nil, "type", common.TypeOf(c))
	}
	return nil
}

func (u *union) get() (proto.Cerealizable, error) {
	switch u.Which {
	case proto.DRKeyMgmt_Which_drkeyLvl1Req:
		return u.Lvl1Req, nil
	case proto.DRKeyMgmt_Which_drkeyLvl1Rep:
		return u.Lvl1Rep, nil
	case proto.DRKeyMgmt_Which_drkeyLvl2Req:
		return u.Lvl2Req, nil
	case proto.DRKeyMgmt_Which_drkeyLvl2Rep:
		return u.Lvl2Rep, nil
	}
	return nil, common.NewBasicError("Unsupported drkey mgmt union type (get)",
		nil, "type", u.Which)
}

var _ proto.Cerealizable = (*Pld)(nil)

type Pld struct {
	union
	*Data
}

// NewPld creates a new drkey mgmt payload, containing the supplied Cerealizable instance.
func NewPld(u proto.Cerealizable, d *Data) (*Pld, error) {
	p := &Pld{Data: d}
	return p, p.union.set(u)
}

func (p *Pld) Union() (proto.Cerealizable, error) {
	return p.union.get()
}

func (p *Pld) ProtoId() proto.ProtoIdType {
	return proto.DRKeyMgmt_TypeID
}

func (p *Pld) String() string {
	desc := []string{"DRKeyMgmt: Union:"}
	u, err := p.Union()
	if err != nil {
		desc = append(desc, err.Error())
	} else {
		desc = append(desc, fmt.Sprintf("%+v", u))
	}
	return strings.Join(desc, " ")
}

type Data struct {
	// For passing any future non-union data.
}
