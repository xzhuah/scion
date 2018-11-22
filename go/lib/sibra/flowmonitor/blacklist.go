package flowmonitor

import (
	"github.com/scionproto/scion/go/lib/log"
	"github.com/willf/bloom"
	"sync"
	"time"
)

const (
	EXPECTED_FLOW_COUNT = 5000
	HASH_FUNCTINOS = 5
)

type blacklistedFlowsStore struct {
	probabilisticStore *bloom.BloomFilter
	blacklistedFlows map[EphemeralId]bool
}

func newBLStore() *blacklistedFlowsStore{
	res := &blacklistedFlowsStore{
		probabilisticStore:bloom.New(EXPECTED_FLOW_COUNT,HASH_FUNCTINOS),
		blacklistedFlows:make(map[EphemeralId]bool),
	}

	return res
}

func (bls *blacklistedFlowsStore )BlacklistFlow(id EphemeralId){
	bls.blacklistedFlows[id]=true
	bls.probabilisticStore.Add(id[:])
}

func (bls *blacklistedFlowsStore )IsBlacklisted(id EphemeralId) bool{
	if bls.probabilisticStore.Test(id[:]){
		_, ok := bls.blacklistedFlows[id]
		return ok
	}else{
		return false
	}
}

func (bls *blacklistedFlowsStore )ClearAllFlows(){
	bls.probabilisticStore.ClearAll()
	bls.blacklistedFlows=make(map[EphemeralId]bool)
}

type Blacklist struct {
	rwLock *sync.RWMutex

	currentStore int
	stores [2]*blacklistedFlowsStore
}

func NewBlacklist() *Blacklist{
	dataStores := [2]*blacklistedFlowsStore{
		newBLStore(),
		newBLStore(),
	}

	bl:=&Blacklist{
		rwLock:&sync.RWMutex{},
		currentStore:0,
		stores:dataStores,
	}

	go bl.changeStores()

	return bl
}

func (bl *Blacklist)switchStore() {
	var nextStoreId int
	if bl.currentStore==0 {
		nextStoreId=1
	}else{
		nextStoreId=0
	}

	bl.rwLock.Lock()
	defer bl.rwLock.Unlock()

	bl.stores[nextStoreId].ClearAllFlows()
	bl.currentStore=nextStoreId
}

func (bl *Blacklist)AddFlow(id EphemeralId){
	bl.rwLock.Lock()
	defer bl.rwLock.Unlock()

	bl.stores[bl.currentStore].BlacklistFlow(id)
}

func (bl *Blacklist)IsFlowBlacklisted(id EphemeralId) bool {
	bl.rwLock.RLock()
	defer bl.rwLock.RUnlock()

	return bl.stores[0].IsBlacklisted(id) || bl.stores[1].IsBlacklisted(id)
}

func (bl *Blacklist)changeStores() {
	log.Debug("Starting blacklist filter switcher")
	tick := time.Tick(16 * time.Second)
	for {
		select {
			case <- tick:
				bl.switchStore()
		}
	}
}