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

package conf

import (
	"encoding/json"
	"io/ioutil"
	"sync"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/sibra"
	"github.com/scionproto/scion/go/lib/spath/spathmeta"
)

const (
	ResvName = "reservations.json"
	ErrorOpenResvs  = "Unable to open reservation map"
	ErrorParseResvs = "Unable to parse reservation map from JSON"
)

type ResvsMap struct {
	sync.Mutex
	reservations map[string]*Resv
}

type StatusFlag uint8
const (
	Unchanged   StatusFlag = 0x00
	New 		StatusFlag = 0x01
	Deleted 	StatusFlag = 0x02
)

func (sf StatusFlag)IsDeleted() bool {
	return sf & Deleted == Deleted
}

func (sf StatusFlag)IsNew() bool {
	return sf & New == New
}

type Resv struct {
	IA            addr.IA
	PathPredicate *spathmeta.PathPredicate
	PathType      sibra.PathType
	MinSize       sibra.BwCls
	DesiredSize   sibra.BwCls
	MaxSize       sibra.BwCls
	SplitCls      sibra.SplitCls
	EndProps      sibra.EndProps
	Telescoping   string
	status 		  StatusFlag
}

func ReservationsFromRaw(b common.RawBytes) (*ResvsMap, error) {
	var m ResvsMap
	if err := json.Unmarshal(b, &m.reservations); err != nil {
		return nil, common.NewBasicError(ErrorOpenResvs, err)
	}
	for _, v := range m.reservations{
		v.status |= New
	}
	return &m, nil
}

func ReservationsFromFile(path string) (*ResvsMap, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, common.NewBasicError(ErrorParseResvs, err, "path", path)
	}
	return ReservationsFromRaw(b)
}

func (rm *ResvsMap)GetReservation(key string)(*Resv, StatusFlag){
	rm.Lock()
	defer rm.Unlock()

	resv, exists := rm.reservations[key]
	if !exists{
		return nil, Deleted
	}
	status := resv.status

	if status.IsDeleted(){
		delete(rm.reservations, key)
	}

	resv.status=Unchanged
	return resv, status
}
