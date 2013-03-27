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

func (list PeerList) Add(item *PeerItem) {
	list = append(list, item)
}


func (list PeerList) Concat(newList PeerList) {
	list = append(list, newList...)
}

