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

func (list *SafeIPList) Len() int {
	list.m.Lock()
	retVal := len(list.list)
	list.m.Unlock()
	return retVal
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

func (list *SafeIPList) RemoveAt(indexList ...int) {
	list.m.Lock()
	subtract := 0
	i := 0
	for j, elem := range list.list {
		list.list[j - subtract] = elem
		if i < len(indexList) && indexList[i] == j {
			i++
			subtract++
		}
	}
	list.list = list.list[:len(list.list) - subtract]
	list.m.Unlock()
}

