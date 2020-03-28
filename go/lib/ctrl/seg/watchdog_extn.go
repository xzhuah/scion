package seg

import (
	"fmt"

	"github.com/scionproto/scion/go/proto"
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

func (bc *BwCluster) String() string {
	if bc == nil {
		return fmt.Sprintf("%v", false)
	}
	return fmt.Sprintf("%+v", *bc)
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
		return fmt.Sprintf("%v", false)
	}
	return fmt.Sprintf("[EgressBW=%v | InToOutBW=%v]", bi.EgressBW, bi.InToOutBW)
}

type WatchDogMetricExtn struct {
	Set    bool
	BwInfo *BwInfo
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
	}
}

func (wdExt *WatchDogMetricExtn) ProtoId() proto.ProtoIdType {
	return proto.WatchDogMetricExt_TypeID
}

func (wdExt *WatchDogMetricExtn) String() string {
	if wdExt == nil {
		return fmt.Sprintf("%v", false)
	}
	return fmt.Sprintf("[Set=%v | BwInfo=%v]", wdExt.Set, *wdExt.BwInfo)
}
