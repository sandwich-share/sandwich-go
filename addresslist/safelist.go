package addresslist

import(
	"sync"
	"net"
)

// A thread safe wrapper around an IPSlice
type SafeIPList struct {
	list IPSlice
	m sync.Mutex
}

func New(list IPSlice) *SafeIPList {
	var mutex sync.Mutex
	return &SafeIPList{list, mutex}
}

func (list *SafeIPList) Add(address net.IP) {
	list.m.Lock()
	list.list.Add(address)
	list.m.Unlock()
}

func (list *SafeIPList) Concat(newList IPSlice) {
	list.m.Lock()
	list.list.Concat(newList)
	list.m.Unlock()
}

func (list *SafeIPList) At(index int) net.IP {
	list.m.Lock()
	address := list.list[index]
	list.m.Unlock()
	return address
}

func (list *SafeIPList) Copy(newList IPSlice) {
	list.m.Lock()
	list.list = newList
	list.m.Unlock()
}

// Returns a COPY of the underlying IPSlice in the SafeIPList thus
// it will not change as the SafeIPList is modified
func (list *SafeIPList) Contents() IPSlice {
	list.m.Lock()
	retVal := make(IPSlice, len(list.list))
	copy(retVal, list.list)
	list.m.Unlock()
	return retVal
}

