package main

import(
	"net"
	"log"
	"io/ioutil"
	"math/rand"
	"sort"
	"time"
	"sandwich-go/addresslist"
	"fmt"
)

func Get(address net.IP, extension string) ([]byte, error) {
	conn, err := net.Dial("tcp", address.String() + GetPort(address))
	if err != nil {
		return nil, err
	}
	fmt.Fprintf(conn, "GET /" + extension + " HTTP/1.1\r\n\r\n")
	data, err := ioutil.ReadAll(conn)
	if err != nil {
		log.Println(err)
	}
	fmt.Println(string(data))
	err = conn.Close()
	return data, err
}

func GetPeerList(address net.IP) (addresslist.PeerList, error) {
	resp, err := Get(address ,"peerlist")
	/*if err != nil {
		log.Println(err)
		return nil, err
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return nil, err
	}*/
	peerlist := addresslist.Unmarshal(resp)
	return peerlist, err
}

func UpdateAddressList(newList addresslist.PeerList) {
	oldList := AddressList.Contents()
	var resultList addresslist.PeerList
	sort.Sort(oldList)
	sort.Sort(newList)
	i := 0
	j := 0
	for i < len(oldList) && j < len(newList) {
		if oldList[i].IP.Equal(LocalIP) {
			i++
		} else if newList[j].IP.Equal(LocalIP) {
			j++
		} else if addresslist.IPLess(oldList[i].IP, newList[j].IP) {
			resultList = append(resultList, oldList[i])
			i++
		} else if addresslist.IPLess(newList[j].IP, oldList[i].IP) {
			resultList = append(resultList, newList[j])
			j++
		} else if oldList[i].LastSeen.Before(newList[j].LastSeen) {
			resultList = append(resultList, newList[j]) //if the IPs match, keep the most recent
			i++
			j++
		} else {
			resultList = append(resultList, oldList[i])
			i++
			j++
		}
	}
	for ;i < len(oldList); i++ {
		resultList = append(resultList, oldList[i])
	}
	for ;j < len(newList); j++ {
		resultList = append(resultList, newList[j])
	}
	Save(resultList)
	AddressList.Copy(resultList)
}

func Ping(address net.IP) bool {
	resp, err := Get(address, "ping")
	if err != nil {
		log.Println(err)
		return false
	}
	return string(resp) == "pong\n"
}

func InitializeKeepAliveLoop() {
	if AddressList.Len() == 0 {
		log.Fatal("AddressList ran out of peers")
	}
	if Settings.PingUntilFoundOnStart {
		for !Ping(AddressList.At(0).IP) {}
	}
	KeepAliveLoop()
}

func KeepAliveLoop() {
	log.Println("KeepAliveLoop has been started")
	for {
		if AddressList.Len() == 0 {
			log.Fatal("AddressList ran out of peers")
		}
		index := rand.Intn(AddressList.Len())
		entry := AddressList.At(index)
		peerList, err := GetPeerList(entry.IP)
		AddressList.RemoveAt(index)
		if err != nil { //The peer gets deleted from the list if error
			log.Println(err)
			continue //shit happens but we do not want a defunct list
		}
		AddressList.Add(&addresslist.PeerItem{entry.IP, entry.IndexHash, time.Now()})
		UpdateAddressList(peerList)
		time.Sleep(2 * time.Second)
	}
}

