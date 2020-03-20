package seg

import (
	"fmt"

	"github.com/scionproto/scion/go/proto"
)

var _ proto.Cerealizable = (*WatchDogMetricExtn)(nil)

type WatchDogMetricExtn struct {
	Set bool
	Val uint32
}

func NewWatchDogMetricExtn() *WatchDogMetricExtn {
	return &WatchDogMetricExtn{Set: true, Val: 32}
}

func (wdExt *WatchDogMetricExtn) ProtoId() proto.ProtoIdType {
	return proto.WatchDogMetricExt_TypeID
}

func (wdExt *WatchDogMetricExtn) String() string {
	if wdExt == nil {
		return fmt.Sprintf("%v", false)
	}
	return fmt.Sprintf("%v%d", wdExt.Set, wdExt.Value)
}
