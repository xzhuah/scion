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
	"io/ioutil"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/scionproto/scion/go/lib/common"
)

var testReservations ResvsMap

func loadFile(filename string, t *testing.T) common.RawBytes {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		t.Fatalf("Error loading config from '%s': %v", filename, err)
	}
	return b
}

func loadReservations(filename string, t *testing.T) {
	m, err := ReservationsFromFile(filename)
	if err != nil {
		t.Fatalf("Error loading config from '%s': %v", filename, err)
	}
	testReservations = m
}

func Test_ReservationFromRaw(t *testing.T) {
	fn := "testdata/reservations.json"
	Convey("Matrix is loaded correctly", t, func() {
		raw := loadFile(fn, t)
		m, err := ReservationsFromRaw(raw)
		SoMsg("Err", err, ShouldBeNil)
		SoMsg("Predicate", m["Up-1-ff00:0:110"].PathPredicate.String(),
			ShouldEqual, "1-ff00:0:110#0")
	})
}
