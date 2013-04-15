package addresslist

import(
	"net"
	"encoding/xml"
	"log"
	"sync"
)

func Inc(address net.IP) net.IP {
	addressCopy := make(net.IP, len(address))
	copy(addressCopy, address)
	return inc(addressCopy)
}

//Destroys the array passed in
func inc(address net.IP) net.IP {
	address[len(address) - 1]++
	if address[len(address) - 1] == byte(0) && len(address) > 1 {
		return append(inc(address[:len(address) - 1]), address[len(address) - 1])
	}
	return address
}

type IPRange struct {
	Start net.IP
	End net.IP
}

func (pair *IPRange) String() string {
	return pair.Start.String() + " to " + pair.End.String()
}

func (pair *IPRange) Equal(newRange *IPRange) bool {
	return pair.Start.Equal(newRange.Start) && pair.End.Equal(newRange.End)
}

func (pair *IPRange) Has(address net.IP) bool {
	return !IPLess(address, pair.Start) && !IPLess(pair.End, address)
}

func (pair *IPRange) shouldCombine(newRange *IPRange) bool {
	return pair.Has(newRange.Start) || newRange.Start.Equal(Inc(pair.End))
}

type BlackWhiteList struct {
	whitelist, blacklist []*IPRange
	m sync.RWMutex
}

//Destroys the array passed in
func remove(list []*IPRange, index int) []*IPRange {
	for i := index + 1; i < len(list); i++ {
		list[i - 1] = list[i]
	}
	return list[:len(list) - 1]
}

func NewBWList(whitelist []*IPRange) *BlackWhiteList {
	retVal := new(BlackWhiteList)
	retVal.whitelist = whitelist
	return retVal
}

func UnmarshalBlackWhite(data []byte) (*BlackWhiteList, error) {
	retVal := new(BlackWhiteList)
	err := xml.Unmarshal(data, retVal)
	if err != nil {
		return nil, err
	}
	return retVal, nil
}

//This is very useful for testing
func (list *BlackWhiteList) Equal(newList *BlackWhiteList) bool {
	list.m.RLock()
	if len(list.whitelist) != len(newList.whitelist) || len(list.blacklist) != len(newList.blacklist) {
		list.m.RUnlock()
		return false
	}
	for i, iprange := range list.whitelist {
		if !iprange.Equal(newList.whitelist[i]) {
			list.m.RUnlock()
			return false
		}
	}
	for i, iprange := range list.blacklist {
		if !iprange.Equal(newList.blacklist[i]) {
			list.m.RUnlock()
			return false
		}
	}
	list.m.RUnlock()
	return true
}

func (list *BlackWhiteList) String() string {
	list.m.RLock()
	retVal := "Whitelist:\n"
	for _, elem := range list.whitelist {
		retVal += elem.String() + "\n"
	}
	retVal += "Blacklist:\n"
	for _, elem := range list.blacklist {
		retVal += elem.String() + "\n"
	}
	list.m.RUnlock()
	return retVal
}

func (list *BlackWhiteList) Marshal() []byte {
	list.m.RLock()
	data, err := xml.Marshal(list)
	if err != nil {
		log.Println(err)
	}
	list.m.RUnlock()
	return data
}

func (list *BlackWhiteList) FilterList(peerlist PeerList) PeerList {
	list.m.RLock()
	resultlist := make(PeerList, 0, len(peerlist))
	for _, peeritem := range peerlist {
		keep := false
		for _, elem := range list.whitelist {
			if elem.Has(peeritem.IP) {
				keep = true
				break
			}
		}
		if !keep {
			continue
		}
		for _, elem := range list.blacklist {
			if elem.Has(peeritem.IP) {
				keep = false
				break
			}
		}
		if !keep {
			continue
		}
		resultlist = append(resultlist, peeritem)
	}
	list.m.RUnlock()
	return resultlist
}

func (list *BlackWhiteList) BlacklistRange(newRange *IPRange) {
	list.m.Lock()
	for i, iprange := range list.blacklist {
		if iprange.shouldCombine(newRange) {
			if IPGreater(newRange.End, iprange.End) {
				iprange.End = newRange.End
				for i++; i < len(list.blacklist); {
					if iprange.shouldCombine(list.blacklist[i]) {
						if IPLess(iprange.End, list.blacklist[i].End) {
							iprange.End = list.blacklist[i].End
							list.blacklist = remove(list.blacklist, i)
							break
						}
						list.blacklist = remove(list.blacklist, i)
					} else {
						break
					}
				}
			}
			list.m.Unlock()
			return
		}
		if newRange.shouldCombine(iprange) {
			iprange.Start = newRange.Start
			if IPGreater(newRange.End, iprange.End) {
				iprange.End = newRange.End
				for i++; i < len(list.blacklist); {
					if iprange.shouldCombine(list.blacklist[i]) {
						if IPLess(iprange.End, list.blacklist[i].End) {
							iprange.End = list.blacklist[i].End
							list.blacklist = remove(list.blacklist, i)
							break
						}
						list.blacklist = remove(list.blacklist, i)
					} else {
						break
					}
				}
			}
			list.m.Unlock()
			return
		}
	}
	//If we get this far we know that range being inserted is disjoint from every other range
	for i, iprange := range list.blacklist {
		if IPLess(newRange.End, iprange.Start) {
			temp := make([]*IPRange, len(list.blacklist[i:]))
			copy(temp, list.blacklist[i:])
			list.blacklist = append(append(list.blacklist[:i], newRange), temp...)
			list.m.Unlock()
			return
		}
	}
	list.m.Unlock()
	list.blacklist = append(list.blacklist, newRange)
}

