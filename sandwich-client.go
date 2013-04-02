package main

import(
	"net"
	"log"
	"io/ioutil"
	"math/rand"
	"sort"
	"time"
	"sandwich-go/addresslist"
	"net/http"
	"bufio"
	"sandwich-go/fileindex"
)

func Get(address net.IP, extension string) ([]byte, error) {
	conn, err := net.Dial("tcp", address.String() + GetPort(address))
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequest("GET", extension, nil)
	if err != nil {
		return nil, err
	}
	err = request.Write(conn)
	if err != nil {
		return nil, err
	}
	buffer := bufio.NewReader(conn)
	response, err := http.ReadResponse(buffer, request)
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	err = conn.Close()
	return data, err
}

func GetFileIndex(address net.IP) (*fileindex.FileList, error) {
	resp, err := Get(address, "/indexfor/")
	if err != nil {
		log.Println(err)
		return nil, err
	}
	fileList := fileindex.Unmarshal(resp)
	return fileList, err
}

func GetPeerList(address net.IP) (addresslist.PeerList, error) {
	resp, err := Get(address ,"/peerlist/")
	if err != nil {
		log.Println(err)
		return nil, err
	}
	peerlist := addresslist.Unmarshal(resp)
	return peerlist, err
}

func UpdateAddressList(newList addresslist.PeerList) {
	oldList := AddressList.Contents()
	var resultList addresslist.PeerList
	reduceMap := make(map[string]*addresslist.PeerItem)
	sort.Sort(oldList)
	sort.Sort(newList)
	for _, elem := range oldList {
		if elem.IP.Equal(LocalIP) {
			continue
		}
		toReplace, ok := reduceMap[string(elem.IP)]
		if ok {
			if toReplace.LastSeen.Before(elem.LastSeen) {
				reduceMap[string(elem.IP)] = elem
			}
		} else {
			reduceMap[string(elem.IP)] = elem
		}
	}
	for _, elem := range newList {
		if elem.IP.Equal(LocalIP) {
			continue
		}
		toReplace, ok := reduceMap[string(elem.IP)]
		if ok {
			if toReplace.LastSeen.Before(elem.LastSeen) {
				reduceMap[string(elem.IP)] = elem
			}
		} else {
			reduceMap[string(elem.IP)] = elem
		}
	}
	for _, value := range reduceMap {
		resultList = append(resultList, value)
	}
	Save(resultList)
	AddressList.Copy(resultList)
}

func Ping(address net.IP) bool {
	resp, err := Get(address, "/ping/");
	if err != nil {
		log.Println(err)
		return false
	}
	if string(resp) == "pong\n" {
		return true
	}
	return false
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

