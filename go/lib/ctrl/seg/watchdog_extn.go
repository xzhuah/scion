package seg

import (
	"fmt"

	"github.com/scionproto/scion/go/proto"
)

var _ proto.Cerealizable = (*WatchDogMetricExtn)(nil)

type WatchDogMetricExtn struct {
	Set           bool
	bandwidthinfo Bandwidthinfo
}

type Bandwidthinfo struct {
	bandwidthClusters []BwCluster
	egressBW          uint32
	inToOutBW         uint32
}

type BwCluster struct {
	clusterBW  uint32
	interfaces []uint16
}

func NewWatchDogMetricExtn() *WatchDogMetricExtn {
	return &WatchDogMetricExtn{Set: true, bandwidthinfo: Bandwidthinfo{bandwidthClusters: nil, egressBW: 1, inToOutBW: 2}}
}

func (wdExt *WatchDogMetricExtn) ProtoId() proto.ProtoIdType {
	return proto.WatchDogMetricExt_TypeID
}

func (wdExt *WatchDogMetricExtn) String() string {
	if wdExt == nil {
		return fmt.Sprintf("%v", false)
	}
	return fmt.Sprintf("%v%v%v", wdExt.Set, wdExt.bandwidthinfo.egressBW, wdExt.bandwidthinfo.inToOutBW)
}
