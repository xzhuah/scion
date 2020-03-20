package seg

import (
	"fmt"

	"github.com/scionproto/scion/go/proto"
)

var _ proto.Cerealizable = (*WatchDog)(nil)

type WatchDog struct {
	Set   bool
	Value uint32
}

func NewWatchDog() *WatchDog {
	return &WatchDog{Set: true, Value: 32}
}

func (wdExt *WatchDog) ProtoId() proto.ProtoIdType {
	return proto.WatchDogMetricExt_TypeID
}

func (wdExt *WatchDog) String() string {
	if wdExt == nil {
		return fmt.Sprintf("%v", false)
	}
	return fmt.Sprintf("%v%d", wdExt.Set, wdExt.Value)
}
