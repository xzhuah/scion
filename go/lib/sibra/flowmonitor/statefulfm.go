package flowmonitor

import (
	"github.com/scionproto/scion/go/lib/sibra"
	"time"
)

// Stateful flow monitor implementation
type leakyBucket struct {
	nextSendTime time.Time
}

type StatefulFlowMonitor struct {
	flowStates map[EphemeralId]*leakyBucket
	burstAmount time.Duration
}

func NewStetefulMonitor(maxBurstAmount time.Duration)FlowMonitor{
	fm := StatefulFlowMonitor {
		flowStates:make(map[EphemeralId]*leakyBucket),
		burstAmount:maxBurstAmount,
	}

	//TODO: Create cleanup goroutine
	return &fm
}

func (fm *StatefulFlowMonitor)IsFlowRateExceeded(info *FlowInfo) bool {
	var flowId EphemeralId
	copy(flowId[:], info.ReservationId[:sibra.EphemIDLen])
	var state *leakyBucket = nil

	if fs, ok := fm.flowStates[flowId]; ok{
		state=fs
	}else{
		state = &leakyBucket{
			nextSendTime:time.Now(),
		}
		fm.flowStates[flowId]=state
	}

	now:=time.Now()
	rightMargin:=now.Add(fm.burstAmount)
	if rightMargin.Before(state.nextSendTime) {
		// Bandwidth exceeded!
		return true
	} else {
		sendTime := ((info.PacketSize) * int(time.Second)) / int(info.BwCls.Bps())
		if now.Before(state.nextSendTime){
			state.nextSendTime=state.nextSendTime.Add(time.Duration(sendTime))
		}else{
			state.nextSendTime=now.Add(time.Duration(sendTime))
		}
		return false
	}
}

func (fm *StatefulFlowMonitor)ClearFlow(flow sibra.ID){
	var flowId EphemeralId
	copy(flowId[:], flow[:sibra.EphemIDLen])
	//TODO: Add lock. This map should be protected by a mutex
	delete(fm.flowStates, flowId)
}
