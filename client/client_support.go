package client

import (
	"bufio"
	"compress/gzip"
	"errors"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	"sandwich-go/addresslist"
	"sandwich-go/fileindex"
	"sandwich-go/settings"
	"sandwich-go/util"
	"strings"
	"time"
)

var addressList *addresslist.SafeIPList        //Thread safe
var addressSet *addresslist.AddressSet         //Thread safe
var blackWhiteList *addresslist.BlackWhiteList //Thread safe
var illegalIPError = errors.New("The requested ip is illegal")
var localIP net.IP
var removeSet map[string]time.Time
var sandwichSettings *settings.Settings

func get(address net.IP, extension string) ([]byte, error) {
	if !blackWhiteList.OK(address) {
		return nil, illegalIPError
	}
	conn, err := net.DialTimeout("tcp", address.String()+util.GetPort(address),
		2*time.Second)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	request, err := http.NewRequest("GET", extension, nil)
	if err != nil {
		return nil, err
	}
	request.Header = map[string][]string{
		"Accept-Encoding": {"gzip"},
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
	var data []byte
	if strings.Contains(response.Header.Get("Content-Encoding"), "gzip") {
		unzipper, err := gzip.NewReader(response.Body)
		if err != nil {
			return nil, err
		}
		data, err = ioutil.ReadAll(unzipper)
	} else {
		data, err = ioutil.ReadAll(response.Body)
	}
	if err != nil {
		return nil, err
	}
	return data, err
}

func getFileIndexLoop(in chan net.IP, out chan fileindex.FileManifest) {
	resultSet := fileindex.NewFileManifest()
	for ip := range in {
		fileList, err := GetFileIndex(ip)
		if err != nil {
			continue
		}
		resultSet.Put(ip, fileList)
	}
	out <- resultSet
}

func aggregate(manifest, outManifest fileindex.FileManifest) fileindex.FileManifest {
	for ip, fileList := range manifest {
		outManifest[ip] = fileList
	}
	return outManifest
}

func removeDeadEntries(newList addresslist.PeerList) addresslist.PeerList {
	resultList := make(addresslist.PeerList, 0, len(newList))
	for _, elem := range newList {
		lastSeen, ok := removeSet[elem.IP.String()]
		if ok && lastSeen.After(elem.LastSeen) {
			continue
		} else if ok {
			delete(removeSet, elem.IP.String())
		} else {
			resultList = append(resultList, elem)
		}
	}
	return resultList
}

func flush() {
	now := time.Now()
	for ip, lastSeen := range removeSet {
		if now.Sub(lastSeen) > time.Hour {
			delete(removeSet, ip)
		}
	}
}

func keepAliveLoop() {
	log.Println("KeepAliveLoop has been started")
	removeSet = make(map[string]time.Time)
	lastFlush := time.After(time.Hour)
	for {
		var peerList addresslist.PeerList
		var err error
		if addressSet.Len() > 0 {
			address := addressSet.Pop()
			peerList, err = GetPeerList(address)
			if err != nil { //The peer gets deleted from the list if error
				log.Println(err)
				continue //shit happens but we do not want a defunct list
			}
		} else if addressList.Len() == 0 {
			if sandwichSettings.LoopOnEmpty {
				time.Sleep(5 * time.Second)
				continue
			}
			log.Fatal("AddressList ran out of peers")
		} else {
			index := rand.Intn(addressList.Len())
			entry := addressList.At(index)
			peerList, err = GetPeerList(entry.IP)
			addressList.RemoveAt(index)
			if err != nil { //The peer gets deleted from the list if error
				//Remember to delete them when merging list
				removeSet[entry.IP.String()] = time.Now()
				log.Println(err)
				continue //shit happens but we do not want a defunct list
			}
			addressList.Add(&addresslist.PeerItem{entry.IP, entry.IndexHash,
				time.Now()})
		}
		updateAddressList(peerList)
		select {
		case <-lastFlush:
			lastFlush = time.After(time.Hour)
			flush()
		default:
			time.Sleep(2 * time.Second)
		}
	}
}
