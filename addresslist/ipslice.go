package addresslist

import(
	"net"
	"log"
	"encoding/json"
)

type IPSlice []IP

func (ipList IPSlice) String() string {
	json, err := json.Marshal(ipList)
	if err != nil {
		log.Println(err)
	}
	return json
}

func FromString(json string) IPSlice {
	ipList := make(IPSlice, 10)
	err := json.Unmarshal(json, ipList)
	if err != nil {
		log.Println(err)
	}
	return ipList
}

func (ipList IPSlice) Add(address IP) {
	ipList = append(ipList, address)
}

func (ipList IPSlice) Concat(newList IPSlice) {
	ipList = append(ipList, newList...)
}

