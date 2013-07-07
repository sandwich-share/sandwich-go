package addresslist

import(
	"net"
	"time"
	"log"
	"encoding/json"
)

// A very convenient representation of the elements that identify a peer
type PeerItem struct {
	IP net.IP
	IndexHash uint32
	LastSeen time.Time
}

// Need to convert PeerItems to this since json.Marshal chokes on []byte to IP
type SerialItem struct {
	IP string
	IndexHash uint32
	LastSeen time.Time
}

type PeerList []*PeerItem

// For json.Marshal to override default serialization
func (item *PeerItem) MarshalJSON() ([]byte, error) {
	serialItem := &SerialItem{item.IP.String(), item.IndexHash, item.LastSeen}
	jsonItem, err := json.Marshal(serialItem)
	return jsonItem, err
}

// For json.Unmarshal to override default serialization
func (item *PeerItem) UnmarshalJSON(jsonItem []byte) error {
	serialItem := &SerialItem{}
	err := json.Unmarshal(jsonItem, serialItem)
	item.IP = net.ParseIP(serialItem.IP)
	item.IndexHash = serialItem.IndexHash
	item.LastSeen = serialItem.LastSeen
	return err
}

//This is useful for testing
func (item *PeerItem) Equal(newItem *PeerItem) bool {
    return item.IP.Equal(newItem.IP) && item.IndexHash == item.IndexHash && item.LastSeen.Equal(newItem.LastSeen)
}

func Unmarshal(jsonList []byte) PeerList {
	var list PeerList
	err := json.Unmarshal(jsonList, &list)
	if err != nil {
		log.Println(err)
	}
	return list
}

func (list PeerList) Marshal() []byte {
	jsonList, err := json.Marshal(list)
	if err != nil {
		log.Println(err)
	}
	return jsonList
}

func (list PeerList) Len() int {
	return len(list)
}

func (list PeerList) Equal(newList PeerList) bool {
    if list.Len() != newList.Len() {
        return false
    }
    for i := range list {
        if list[i] != newList[i] {
            return false
        }
    }
    return true
}

func IPLess(listA, listB net.IP) bool {
	first := []byte(listA)
	second := []byte(listB)
	if len(first) < len(second) {
		return true
	} else if len(second) < len(first) {
		return false
	} else {
		for i, elem := range first {
			if elem < second[i] {
				return true
			} else if second[i] < elem {
				return false
			}
		}
	}
	return false
}

func IPLessEqual(listA, listB net.IP) bool {
	return listA.Equal(listB) || IPLess(listA, listB)
}

func IPGreater(listA, listB net.IP) bool {
	return !IPLess(listA, listB) && !listA.Equal(listB)
}

func IPGreaterEqual(listA, listB net.IP) bool {
	return !IPLess(listA, listB)
}

func (list PeerList) Contains(ip net.IP) bool {
	for _, entry := range(list) {
		if entry.IP.Equal(ip) {
			return true
		}
	}
	return false
}

func (list PeerList) Less(i, j int) bool {
	return IPLess(list[i].IP, list[j].IP)
}

func (list PeerList) Swap(i, j int) {
	temp := list[i]
	list[i] = list[j]
	list[j] = temp
}

func (list PeerList) Add(item *PeerItem) {
	list = append(list, item)
}


func (list PeerList) Concat(newList PeerList) {
	list = append(list, newList...)
}

