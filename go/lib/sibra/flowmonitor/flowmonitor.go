package flowmonitor

import (
	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/sibra"
	"hash/fnv"
	"net"
)

type EphemeralId [sibra.EphemIDLen]byte

type FlowMonitoringResult int

const (
	BANDWIDTH_EXCEEDED FlowMonitoringResult = iota	// Packet should be dropped
	BANDWIDTH_BLACKLIST	// Packet should be dropped and flow should be blacklisted
	BANDWIDTH_OK	// Packet should be forwarded
	FLOW_NOT_TRACKED	// Flow hasn't been registered for monitoring. Only makes sense when `addUnknownFlow` is set to false
)

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
	SourceIA addr.IA
	// For debugging purposes
	HostIP net.IP
}
