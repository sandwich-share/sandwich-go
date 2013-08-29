package addresslist

import (
	"encoding/xml"
	"log"
	"net"
	"sync"
)

func Inc(address net.IP) net.IP {
	addressCopy := make(net.IP, len(address))
	copy(addressCopy, address)
	return inc(addressCopy)
}

//Destroys the array passed in
func inc(address net.IP) net.IP {
	address[len(address)-1]++
	if address[len(address)-1] == byte(0) && len(address) > 1 {
		return append(inc(address[:len(address)-1]), address[len(address)-1])
	}
	return address
}

type IPRange struct {
	Start net.IP
	End   net.IP
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

//NOTE: YOU SHOULD NEVER EVER ACCESS BLACKLIST DIRECTLY UNLESS YOU KNOW WHAT YOU ARE DOING
// THE ONLY REASON THAT THIS VALUE IS EXPORTED IS FOR SERIALIZATION
type BlackWhiteList struct {
	whitelist, Blacklist []*IPRange //Blacklist is exported so that it may be serialized properly
	m                    sync.RWMutex
}

//Destroys the array passed in
func remove(list []*IPRange, index int) []*IPRange {
	for i := index + 1; i < len(list); i++ {
		list[i-1] = list[i]
	}
	return list[:len(list)-1]
}

func NewBWList(whitelist []*IPRange) *BlackWhiteList {
	retVal := new(BlackWhiteList)
	retVal.whitelist = whitelist
	return retVal
}

func UnmarshalBWList(data []byte, list []*IPRange) (*BlackWhiteList, error) {
	retVal := new(BlackWhiteList)
	err := xml.Unmarshal(data, retVal)
	if err != nil {
		return nil, err
	}
	retVal.whitelist = list
	return retVal, nil
}

//This is very useful for testing
func (list *BlackWhiteList) Equal(newList *BlackWhiteList) bool {
	list.m.RLock()
	if len(list.whitelist) != len(newList.whitelist) || len(list.Blacklist) != len(newList.Blacklist) {
		list.m.RUnlock()
		return false
	}
	for i, iprange := range list.whitelist {
		if !iprange.Equal(newList.whitelist[i]) {
			list.m.RUnlock()
			return false
		}
	}
	for i, iprange := range list.Blacklist {
		if !iprange.Equal(newList.Blacklist[i]) {
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
	for _, elem := range list.Blacklist {
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

func (list *BlackWhiteList) ok(address net.IP) bool {
	keep := false
	for _, elem := range list.whitelist {
		if elem.Has(address) {
			keep = true
			break
		}
	}
	if !keep {
		return false
	}
	for _, elem := range list.Blacklist {
		if elem.Has(address) {
			keep = false
			break
		}
	}
	return keep
}

func (list *BlackWhiteList) OK(address net.IP) bool {
	list.m.RLock()
	ok := list.ok(address)
	list.m.RUnlock()
	return ok
}

func (list *BlackWhiteList) FilterList(peerlist PeerList) PeerList {
	list.m.RLock()
	resultlist := make(PeerList, 0, len(peerlist))
	for _, peeritem := range peerlist {
		if list.ok(peeritem.IP) {
			resultlist = append(resultlist, peeritem)
		}
	}
	list.m.RUnlock()
	return resultlist
}

func (list *BlackWhiteList) BlacklistRange(newRange *IPRange) {
	list.m.Lock()
	for i, iprange := range list.Blacklist {
		if iprange.shouldCombine(newRange) {
			if IPGreater(newRange.End, iprange.End) {
				iprange.End = newRange.End
				for i++; i < len(list.Blacklist); {
					if iprange.shouldCombine(list.Blacklist[i]) {
						if IPLess(iprange.End, list.Blacklist[i].End) {
							iprange.End = list.Blacklist[i].End
							list.Blacklist = remove(list.Blacklist, i)
							break
						}
						list.Blacklist = remove(list.Blacklist, i)
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
				for i++; i < len(list.Blacklist); {
					if iprange.shouldCombine(list.Blacklist[i]) {
						if IPLess(iprange.End, list.Blacklist[i].End) {
							iprange.End = list.Blacklist[i].End
							list.Blacklist = remove(list.Blacklist, i)
							break
						}
						list.Blacklist = remove(list.Blacklist, i)
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
	for i, iprange := range list.Blacklist {
		if IPLess(newRange.End, iprange.Start) {
			temp := make([]*IPRange, len(list.Blacklist[i:]))
			copy(temp, list.Blacklist[i:])
			list.Blacklist = append(append(list.Blacklist[:i], newRange), temp...)
			list.m.Unlock()
			return
		}
	}
	list.m.Unlock()
	list.Blacklist = append(list.Blacklist, newRange)
}
