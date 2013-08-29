package addresslist

import (
	"net"
	"sync"
)

type AddressSet struct {
	hashSet map[string]net.IP
	lock    sync.RWMutex
}

func NewAddressSet() *AddressSet {
	var lock sync.RWMutex
	hashSet := make(map[string]net.IP)
	return &AddressSet{hashSet, lock}
}

func (set *AddressSet) Len() int {
	set.lock.RLock()
	retVal := len(set.hashSet)
	set.lock.RUnlock()
	return retVal
}

func (set *AddressSet) Contains(address net.IP) bool {
	set.lock.RLock()
	_, ok := set.hashSet[address.String()]
	set.lock.RUnlock()
	return ok
}

func (set *AddressSet) Add(address net.IP) {
	set.lock.Lock()
	set.hashSet[address.String()] = address
	set.lock.Unlock()
}

func (set *AddressSet) Get() net.IP {
	set.lock.RLock()
	var retVal net.IP
	for _, value := range set.hashSet {
		retVal = value
		break
	}
	set.lock.RUnlock()
	return retVal
}

func (set *AddressSet) GetList() []net.IP {
	set.lock.RLock()
	netList := make([]net.IP, 1)
	for _, address := range set.hashSet {
		netList = append(netList, address)
	}
	set.lock.RUnlock()
	return netList
}

func (set *AddressSet) Delete(address net.IP) {
	set.lock.Lock()
	delete(set.hashSet, address.String())
	set.lock.Unlock()
}

func (set *AddressSet) Pop() net.IP {
	set.lock.Lock()
	var retVal net.IP
	for _, value := range set.hashSet {
		retVal = value
		delete(set.hashSet, retVal.String())
		break
	}
	set.lock.Unlock()
	return retVal
}
