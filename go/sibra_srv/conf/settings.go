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
	"github.com/scionproto/scion/go/lib/log"
	"io/ioutil"
	"sync"
	"time"

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

type ResvsMap map[string]*Resv

const (
	NewReservation = 0x01
	DeletedReservation = 0x02
)
type ReservationUpdate struct{
	Key string
	status uint8
}

func (ru ReservationUpdate)IsNewReservation() bool{
	return ru.status & NewReservation != 0
}

func (ru ReservationUpdate)IsDeletedReservation() bool{
	return ru.status & DeletedReservation != 0
}

type ResvMonitor struct {
	sync.Mutex
	reservations ResvsMap
	filePath string
	stop chan struct{}
	newReservations chan ReservationUpdate
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
}

func ReservationsFromRaw(b common.RawBytes) (ResvsMap, error) {
	var m ResvsMap
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, common.NewBasicError(ErrorOpenResvs, err)
	}
	return m, nil
}

func ReservationsFromFile(path string) (ResvsMap, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, common.NewBasicError(ErrorParseResvs, err, "path", path)
	}
	return ReservationsFromRaw(b)
}

func MonitorReservationsFile(path string) (*ResvMonitor) {
	rm := &ResvMonitor{
		reservations:make(map[string]*Resv),
		filePath:path,
		stop:make(chan struct{}),
		newReservations:make(chan ReservationUpdate, 1),
	}

	go rm.watchFile()
	return rm
}

func (rm *ResvMonitor)GetNewReservations()<-chan ReservationUpdate {
	return rm.newReservations
}

func (rm *ResvMonitor)watchFile(){
	t := time.NewTicker(3*time.Second)
	for {
		select {
		case <-rm.stop:
			return
		case <-t.C:
			newReservations, err:= ReservationsFromFile(rm.filePath)
			if err!=nil{
				log.Warn("Error loading reservations file", "err", err)
				continue
			}

			newElements := getDifference(newReservations, rm.reservations)
			deletedElements := getDifference(rm.reservations, newReservations)
			rm.Lock()
			rm.reservations=newReservations
			rm.Unlock()
			for _, resKey := range deletedElements {
				rm.newReservations<-ReservationUpdate{Key:resKey, status:DeletedReservation}
			}
			for _, resKey := range newElements{
				rm.newReservations<-ReservationUpdate{Key:resKey, status:NewReservation}
			}

		}
	}
}

func getDifference(first ResvsMap, second ResvsMap)[]string{
	res :=  make([]string,0)
	for k := range first{
		if _, ok := second[k]; !ok{
			res=append(res, k)
		}
	}
	return res
}

func (rm *ResvMonitor)GetReservation(key string)(*Resv, bool){
	rm.Lock()
	defer rm.Unlock()
	resv, exists := rm.reservations[key]
	return resv, exists
}
