package seg

import (
	"fmt"
	"github.com/scionproto/scion/go/proto"
	"math/rand"
)

var _ proto.Cerealizable = (*BwCluster)(nil)
var _ proto.Cerealizable = (*BwInfo)(nil)
var _ proto.Cerealizable = (*WatchDogMetricExtn)(nil)

type BwCluster struct {
	ClusterBW  uint32
	Interfaces []uint16
}

func (bc *BwCluster) ProtoId() proto.ProtoIdType {
	return proto.BwCluster_TypeID
}

// bandwidth information
func (bc *BwCluster) String() string {
	if bc == nil {
		return "[nil]"
	}
	return fmt.Sprintf("[ClusterBW=%v | Interfaces=%v]", bc.ClusterBW, bc.Interfaces)
}

type BwInfo struct {
	BwClusters []*BwCluster
	EgressBW   uint32
	InToOutBW  uint32
}

func (bi *BwInfo) ProtoId() proto.ProtoIdType {
	return proto.BwInfo_TypeID
}

func (bi *BwInfo) String() string {
	if bi == nil {
		return "[nil]"
	}
	clusters := "["
	for _, cluster := range bi.BwClusters {
		clusters += fmt.Sprintf("%v ", cluster)
	}
	clusters += "]"
	return fmt.Sprintf("[BwClusters=%v | EgressBW=%v | InToOutBW=%v]", clusters, bi.EgressBW, bi.InToOutBW)
}

type GeoInfo struct {
	Latitude  float64
	Longitude float64
}

func (gi *GeoInfo) ProtoId() proto.ProtoIdType {
	return proto.GeoInfo_TypeID
}

func (gi *GeoInfo) String() string {
	if gi == nil {
		return "[nil]"
	}
	return fmt.Sprintf("[Latitude=%v | Longitude=%v]", gi.Latitude, gi.Longitude)
}

type WatchDogMetricExtn struct {
	Set     bool
	BwInfo  *BwInfo
	GeoInfo *GeoInfo
}

func NewWatchDogMetricExtn() *WatchDogMetricExtn {
	return &WatchDogMetricExtn{
		Set: true,
		BwInfo: &BwInfo{
			BwClusters: []*BwCluster{
				&BwCluster{
					ClusterBW:  100,
					Interfaces: []uint16{1, 3, 5},
				},
				&BwCluster{
					ClusterBW:  200,
					Interfaces: []uint16{2, 4},
				},
			},
			EgressBW:  300,
			InToOutBW: 400,
		},
		GeoInfo: &GeoInfo{
			Latitude:  rand.Float64() * 180,
			Longitude: rand.Float64() * 180,
		},
	}
}

func (wdExt *WatchDogMetricExtn) ProtoId() proto.ProtoIdType {
	return proto.WatchDogMetricExt_TypeID
}

func (wdExt *WatchDogMetricExtn) String() string {
	if wdExt == nil {
		return "[nil]"
	}
	return fmt.Sprintf("[Set=%v | BwInfo=%v | GeoInfo=%v]", wdExt.Set, *wdExt.BwInfo, *wdExt.GeoInfo)
}
