package flowmonitor

import (
	"github.com/scionproto/scion/go/lib/sibra"
	"hash/fnv"
)

type EphemeralId [sibra.EphemIDLen]byte

func (eid EphemeralId)Hash() uint32 {
	h:=fnv.New32()
	h.Write(eid[:])
	return h.Sum32()
}

type FlowInfo struct {
	BwCls sibra.BwCls
	PacketSize int
	ReservationId sibra.ID
	ReservationIndex sibra.Index
}

type FlowMonitor interface {
	IsFlowRateExceeded(info *FlowInfo) bool
	ClearFlow(flow sibra.ID)
}