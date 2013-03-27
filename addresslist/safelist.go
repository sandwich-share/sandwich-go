package addresslist

import(
	"sync"
)

// A thread safe wrapper around an IPSlice
type SafeIPList struct {
	list PeerList
	m sync.Mutex
}

func New(list PeerList) *SafeIPList {
	var mutex sync.Mutex
	return &SafeIPList{list, mutex}
}

func (list *SafeIPList) Add(entry *PeerItem) {
	list.m.Lock()
	list.list.Add(entry)
	list.m.Unlock()
}

func (list *SafeIPList) Concat(newList PeerList) {
	list.m.Lock()
	list.list.Concat(newList)
	list.m.Unlock()
}

func (list *SafeIPList) At(index int) *PeerItem {
	list.m.Lock()
	entry := list.list[index]
	list.m.Unlock()
	return entry
}

func (list *SafeIPList) Copy(newList PeerList) {
	list.m.Lock()
	list.list = newList
	list.m.Unlock()
}

// Returns a COPY of the underlying IPSlice in the SafeIPList thus
// it will not change as the SafeIPList is modified
func (list *SafeIPList) Contents() PeerList {
	list.m.Lock()
	retVal := make(PeerList, len(list.list))
	copy(retVal, list.list)
	list.m.Unlock()
	return retVal
}

