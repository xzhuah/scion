package flowmonitor

import (
	"github.com/scionproto/scion/go/lib/sibra"
	"sync"
	"time"
)

// Stateful flow monitor implementation
type leakyBucket struct {
	createdTime time.Time
	nextSendTime time.Time
}

type StatefulFlowMonitor struct {
	stateLock sync.Mutex

	flowStates map[EphemeralId]*leakyBucket
	burstAmount time.Duration
}

func NewStetefulMonitor(maxBurstAmount time.Duration)*StatefulFlowMonitor{
	fm := StatefulFlowMonitor {
		flowStates:make(map[EphemeralId]*leakyBucket),
		burstAmount:maxBurstAmount,
	}

	//TODO: Create cleanup goroutine
	return &fm
}

func (fm *StatefulFlowMonitor)IsFlowRateExceeded(info *FlowInfo, addUnknownFlow bool) FlowMonitoringResult {
	var flowId EphemeralId
	copy(flowId[:], info.ReservationId[:sibra.EphemIDLen])
	var state *leakyBucket = nil

	fm.stateLock.Lock()
	defer fm.stateLock.Unlock()

	if fs, ok := fm.flowStates[flowId]; ok {
		state=fs
	} else if addUnknownFlow {
		state = &leakyBucket{
			createdTime:time.Now(),
			nextSendTime:time.Now(),
		}
		fm.flowStates[flowId]=state
	}else{
		// Flow hasn't been registered, so we won't check for bandwidth
		return FLOW_NOT_TRACKED
	}

	now:=time.Now()
	rightMargin:=now.Add(fm.burstAmount)
	if rightMargin.Before(state.nextSendTime) {
		// Bandwidth exceeded!
		return BANDWIDTH_EXCEEDED
	} else {
		sendTime := ((info.PacketSize) * int(time.Second)) / int(info.BwCls.Bps())
		if now.Before(state.nextSendTime){
			state.nextSendTime=state.nextSendTime.Add(time.Duration(sendTime))
		} else {
			state.nextSendTime=now.Add(time.Duration(sendTime))
		}
		return BANDWIDTH_OK
	}
}

func (fm *StatefulFlowMonitor)ClearFlow(flow sibra.ID){
	var flowId EphemeralId
	copy(flowId[:], flow[:sibra.EphemIDLen])

	fm.stateLock.Lock()
	defer fm.stateLock.Unlock()

	delete(fm.flowStates, flowId)
}

func (fm *StatefulFlowMonitor)RegisterFlow(flowId EphemeralId){
	fm.stateLock.Lock()
	defer fm.stateLock.Unlock()
	if _, ok := fm.flowStates[flowId]; !ok {
		state := &leakyBucket{
			nextSendTime:time.Now(),
			createdTime:time.Now(),
		}
		fm.flowStates[flowId]=state
	}
}

func (fm *StatefulFlowMonitor)RemoveOldFlows(maxAge time.Duration){
	now := time.Now()

	fm.stateLock.Lock()
	defer fm.stateLock.Unlock()

	for fId, state := range fm.flowStates {
		if state.createdTime.Add(maxAge).Before(now){
			delete(fm.flowStates, fId)
		}
	}
}