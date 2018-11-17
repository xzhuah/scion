package flowmonitor

import "github.com/scionproto/scion/go/lib/sibra"

type EphemeralId [sibra.EphemIDLen]byte

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