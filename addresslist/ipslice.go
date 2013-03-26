package addresslist

import(
	"net"
	"log"
	. "encoding/json"
)

type IPSlice []net.IP

// Makes json Marshalling easier: more human readable and json.Marshal does not play nice with 
// arrays unless you wrap them in something else
type ListWrapper struct {
	List []string
}

func (ipList IPSlice) String() string {
	listWrapper := ListWrapper{make([]string, len(ipList))}
	for i, elem := range ipList {
		listWrapper.List[i] = elem.String()
	}
	json, err := Marshal(listWrapper)
	if err != nil {
		log.Println(err)
	}
	return string(json)
}

func FromString(json string) IPSlice {
	listWrapper := ListWrapper{}
	err := Unmarshal([]byte(json), &listWrapper) //Unmarshal expects a pointer (not immediately clear in spec)
	ipList := make(IPSlice, len(listWrapper.List))
	for i, elem := range listWrapper.List {
		ipList[i] = net.ParseIP(elem)
	}
	if err != nil {
		log.Println(err)
	}
	return ipList
}

func (ipList IPSlice) Add(address net.IP) {
	ipList = append(ipList, address)
}

func (ipList IPSlice) Concat(newList IPSlice) {
	ipList = append(ipList, newList...)
}

