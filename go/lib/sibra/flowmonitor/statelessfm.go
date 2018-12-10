package flowmonitor

import (
	"github.com/aviddiviner/go-murmur"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/sibra"
	"math"
	"math/rand"
	"sort"
	"time"
)

const (
	MINOR_CYCLES_IN_MAJOR_CYCLE int = 64
	MINOR_CYCLE_COUNTERS int = 64
	FLOW_SAMPLES_PER_MAJOR_CYCLE = 100
	DISTRIBUTION_RATE float64 = 1.0 / FLOW_SAMPLES_PER_MAJOR_CYCLE
	AGING_FACTOR = 0.8
	MAX_AGE = 6

	SCALE_FACTOR uint64 = 1000000

	SUSPICIOUS_FLOWS_PER_CYCLE = 2
)

type minorCycle struct {
	counters [MINOR_CYCLE_COUNTERS] uint64
}

func (c *minorCycle)clearCounters(){
	for i:=0; i<MINOR_CYCLE_COUNTERS; i++{
		c.counters[i]=0
	}
}

type slowPathCommand struct {
	bufferIndex int
	majorCycle int
	startOfMajorCycle time.Time
}

type flowTrack struct {
	A uint64
	C uint64
	lastCycle int

	lastUpdated time.Time
}

type StatelessFlowMonitor struct {
	minorCycleTime time.Duration
	minorCycles [2][MINOR_CYCLES_IN_MAJOR_CYCLE] minorCycle

	majorCycleBegin time.Time
	majorCycle int
	majorCycleTime time.Duration

	// For sampling of active flows
	sampledFlows [2][FLOW_SAMPLES_PER_MAJOR_CYCLE] EphemeralId
	sampledFlowsLength [2] int
	nextSamplePeriod time.Time

	currentBufferIndex int

	// Used for sending data from fast path to slow path
	slowPathCommands chan slowPathCommand

	// Structures for slow path
	flowTable map[EphemeralId] flowTrack

	// Keep track of blacklisted flows
	blacklist *Blacklist

	// Used for detailed monitoring of suspicious flows
	perFlowMonitoring *StatefulFlowMonitor
}

func NewStatelessFlowMonitor(minCycleTime time.Duration) *StatelessFlowMonitor {
	res:= &StatelessFlowMonitor {
		minorCycleTime:minCycleTime,
		majorCycleBegin:time.Now(),
		majorCycle:0,
		majorCycleTime:time.Duration(MINOR_CYCLES_IN_MAJOR_CYCLE)*minCycleTime,
		currentBufferIndex:0,
		slowPathCommands:make(chan slowPathCommand),
		flowTable:make(map[EphemeralId] flowTrack),
		blacklist:NewBlacklist(),
		perFlowMonitoring:NewStetefulMonitor(500*time.Millisecond),
	}

	go res.slowPath()

	return res
}

func (fm *StatelessFlowMonitor)IsFlowRateExceeded(info *FlowInfo) FlowMonitoringResult {
	// Step 1: Check if flow has already been blacklisted
	var flowId EphemeralId
	copy(flowId[:], info.ReservationId[:sibra.EphemIDLen])
	if fm.blacklist.IsFlowBlacklisted(flowId){
		return BANDWIDTH_EXCEEDED
	}

	// Step 2: Check if flow is monitored and if it exceeded the limit
	if perFlowResult := fm.perFlowMonitoring.IsFlowRateExceeded(info, false); perFlowResult == BANDWIDTH_OK {
		return BANDWIDTH_OK
	}else if perFlowResult==BANDWIDTH_EXCEEDED {
		log.Debug("Wooohoo, detected overuse flow!")
		fm.blacklist.AddFlow(flowId)
		fm.perFlowMonitoring.ClearFlow(info.ReservationId)
		return BANDWIDTH_BLACKLIST
	}

	// Step 3: Increment bucket counters
	now := time.Now()
	var minorCycleIndex int
	bufferIndex := fm.currentBufferIndex

	if fm.majorCycleBegin.Add(fm.majorCycleTime).Before(now){
		// Send notification to slow path to process data
		fm.slowPathCommands <- slowPathCommand{
			bufferIndex:bufferIndex,
			majorCycle:fm.majorCycle,
			startOfMajorCycle:fm.majorCycleBegin,
		}

		// Reset state
		fm.majorCycle++
		fm.majorCycleBegin = now
		minorCycleIndex = 0
		// Switch to other buffer ID
		if bufferIndex == 1 {
			bufferIndex = 0
		} else {
			bufferIndex = 1
		}
		fm.currentBufferIndex=bufferIndex

		// For sampling active flow count
		fm.sampledFlowsLength[bufferIndex]=0
		fm.nextSamplePeriod=now
	} else {
		elapsedTime := now.Sub(fm.majorCycleBegin)
		minorCycleIndex = int(elapsedTime/fm.minorCycleTime)
	}

	bucketInidex := murmur.MurmurHash2(info.ReservationId, uint32(fm.majorCycle*MINOR_CYCLES_IN_MAJOR_CYCLE+minorCycleIndex)) % uint32(MINOR_CYCLE_COUNTERS)
	normalizedPacketSize := uint64(info.PacketSize)*SCALE_FACTOR/uint64(info.BwCls.Bps())

	fm.minorCycles[bufferIndex][minorCycleIndex].counters[bucketInidex]+=normalizedPacketSize

	if now.After(fm.nextSamplePeriod) && fm.sampledFlowsLength[bufferIndex] < FLOW_SAMPLES_PER_MAJOR_CYCLE {
		fm.sampledFlows[bufferIndex][fm.sampledFlowsLength[bufferIndex]]=flowId
		//copy(fm.sampledFlows[bufferIndex][fm.sampledFlowsLength[bufferIndex]][:], info.ReservationId[:sibra.EphemIDLen])
		fm.sampledFlowsLength[bufferIndex]++
		fm.nextSamplePeriod = fm.nextSamplePeriod.Add(time.Duration(rand.ExpFloat64()/DISTRIBUTION_RATE)*time.Millisecond)
	}

	return BANDWIDTH_OK
}

func (fm *StatelessFlowMonitor)ClearFlow(flow sibra.ID){
	// TODO: Implement
}

type flowUsage struct {
	data uint64
	countedFlows uint64
}

type flowRank struct {
	flowId EphemeralId
	ratio uint64
}

type SuspiciousFlows []*flowRank

func (sf SuspiciousFlows) Len() int { return len(sf) }

func (sf SuspiciousFlows) Less(i, j int) bool {
	return sf[i].ratio > sf[j].ratio
}

func (sf SuspiciousFlows) Swap(i, j int) {
	sf[i], sf[j] = sf[j], sf[i]
}

func (fm *StatelessFlowMonitor)slowPath(){
	log.Debug("Starting slow path processing...")
	for processCommand := range fm.slowPathCommands {
		bufferIndex := processCommand.bufferIndex
		sampledFlowsCount := fm.sampledFlowsLength[bufferIndex]

		//log.Debug("Active flows:")
		activeFlows := make(map[EphemeralId] flowUsage)
		for i:=0; i<sampledFlowsCount; i++{
			//log.Debug("  ", "flow_id", fm.sampledFlows[bufferIndex][i].Hash())
			activeFlows[fm.sampledFlows[bufferIndex][i]] = flowUsage{
				data:0,
				countedFlows:0,
			}
		}

		var vcnt uint64 = 0

		for i:=0; i<MINOR_CYCLES_IN_MAJOR_CYCLE; i++{
			var flowCount minorCycle

			for k, _ := range activeFlows {
				bucketIndex := murmur.MurmurHash2(k[:], uint32((processCommand.majorCycle*MINOR_CYCLES_IN_MAJOR_CYCLE)+i)) % uint32(MINOR_CYCLE_COUNTERS)
				flowCount.counters[bucketIndex]++
			}

			for k, v := range activeFlows {
				bucketIndex := murmur.MurmurHash2(k[:], uint32((processCommand.majorCycle*MINOR_CYCLES_IN_MAJOR_CYCLE)+i)) % uint32(MINOR_CYCLE_COUNTERS)

				v.data+=fm.minorCycles[bufferIndex][i].counters[bucketIndex]
				v.countedFlows+=flowCount.counters[bucketIndex]
				activeFlows[k]=v
			}

			for c:=0; c<MINOR_CYCLE_COUNTERS; c++{
				vcnt+=fm.minorCycles[bufferIndex][i].counters[c]
			}

		}

		activeFlowCount := len(activeFlows)
		vcnt = vcnt/uint64(MINOR_CYCLE_COUNTERS)
		vsize := (activeFlowCount/MINOR_CYCLE_COUNTERS+ 1)*MINOR_CYCLES_IN_MAJOR_CYCLE

		for k,v := range activeFlows {
			var ad float64 = 1
			var age float64 = .0

			flow := flowTrack {
				A:0,
				C:0,
				lastCycle:processCommand.majorCycle,
			}

			if f, ok := fm.flowTable[k]; ok {
				flow=f
			}
			flow.lastUpdated=processCommand.startOfMajorCycle

			d := processCommand.majorCycle - flow.lastCycle
			ad = math.Pow(AGING_FACTOR, float64(d))

			// Calculate age
			for t := 1; t<d; t++ {
				age += math.Pow(AGING_FACTOR, float64(t))
			}

			// Update age
			flow.A=uint64(ad*float64(flow.A)+age*float64(vcnt))+v.data
			flow.C=uint64(ad*float64(flow.C)+age*float64(vsize))+v.countedFlows
			flow.lastCycle=processCommand.majorCycle

			fm.flowTable[k]=flow
		}

		now := time.Now()
		// Clear the old counters
		flows := make(SuspiciousFlows, 0, 0)
		//TODO: This heap doesn't seem to be working. Need investigante further, for now using sort...
		//heap.Init(&flows)
		for fId, flow := range fm.flowTable {
			if now.Sub(flow.lastUpdated) > MAX_AGE*fm.majorCycleTime {
				delete(fm.flowTable, fId)
			}else{
				flows=append(flows, &flowRank{flowId:fId, ratio:flow.A/flow.C})
				//heap.Push(&flows, &flowRank{flowId:fId, ratio:flow.A/flow.C})
			}
		}
		sort.Sort(flows)

		fm.perFlowMonitoring.RemoveOldFlows(sibra.TickInterval * time.Second * sibra.MaxEphemTicks)
		for i:=0; i<len(flows) && i<SUSPICIOUS_FLOWS_PER_CYCLE; i++ {
			//flowIP := fm.knownFlows[flows[i].flowId]
			//log.Debug("Registring for detailed monitoring", "flow_hash", flows[i].flowId.Hash(), "IP", flowIP)
			fm.perFlowMonitoring.RegisterFlow(flows[i].flowId)
		}

		// Clear all the bucket values so they can be used in the next major cycle

		fm.sampledFlowsLength[bufferIndex]=0
		for i:=0; i<MINOR_CYCLES_IN_MAJOR_CYCLE; i++{
			fm.minorCycles[bufferIndex][i].clearCounters()
		}
	}
}
