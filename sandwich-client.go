package main

import(
	"net"
	"log"
	"io"
	"io/ioutil"
	"math/rand"
	"sort"
	"time"
	"sandwich-go/addresslist"
	"net/http"
	"bufio"
	"os"
	"path"
	"path/filepath"
	"sandwich-go/fileindex"
	"compress/gzip"
	"net/url"
	"strings"
)

var RemoveSet map[string]time.Time

func Get(address net.IP, extension string) ([]byte, error) {
	conn, err := net.DialTimeout("tcp", address.String() + GetPort(address), 2 * time.Second)
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequest("GET", extension, nil)
	if err != nil {
		return nil, err
	}
	request.Header = map[string][]string{
		"Accept-Encoding": {"gzip, deflate"},
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
	err = conn.Close()
	return data, err
}

func DownloadFile(address net.IP, filePath string) error {
	conn, err := net.DialTimeout("tcp", address.String() + GetPort(address), 10 * time.Second)
	if err != nil {
		return err
	}

	url := url.URL{}
	url.Path = filePath
	request, err := http.NewRequest("GET", "/file?path=" + url.String(), nil)
	if err != nil {
		conn.Close()
		return err
	}

	err = request.Write(conn)
	if err != nil {
		conn.Close()
		return err
	}
	buffer := bufio.NewReader(conn)
	response, err := http.ReadResponse(buffer, request)
	if err != nil {
		conn.Close()
		return err
	}
	buffer = bufio.NewReader(response.Body)
	dirPath, _ := filepath.Split(filePath)
	err = os.MkdirAll(path.Join(SandwichPath, dirPath), os.ModePerm)
	if err != nil {
		conn.Close()
		return err
	}
	file, err := os.Create(path.Join(SandwichPath, filePath))
	if err != nil {
		conn.Close()
		file.Close()
		return err
	}
	byteBuf := make([]byte, 1024)
	for done := false; !done; {
		numRead, err := buffer.Read(byteBuf)
		if err == io.EOF {
			done = true
		} else if err != nil {
			conn.Close()
			file.Close()
			return err
		}
		_, err = file.Write(byteBuf[:numRead])
		if err != nil {
			conn.Close()
			file.Close()
			return err
		}
	}
	err = file.Close()
	if err != nil {
		conn.Close()
		return err
	}
	err = conn.Close()
	if err != nil {
		return err
	}
	return err
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

//TODO: Should make this multithreaded. Take advantage of the fact that we are pinging addresses
func BuildFileManifest() {
	peerList := AddressList.Contents()
	for _, item := range peerList {
		fileList, err := GetFileIndex(item.IP)
		if err != nil {
			continue
		}
		FileManifest.Put(item.IP, fileList)
		log.Println("Got index: " + item.IP.String())
	}
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

func removeDeadEntries(newList addresslist.PeerList) addresslist.PeerList {
	resultList := make(addresslist.PeerList, 0, len(newList))
	for _, elem := range newList {
		lastSeen, ok := RemoveSet[elem.IP.String()]
		if ok && lastSeen.After(elem.LastSeen) {
			continue
		} else if ok {
			delete(RemoveSet, elem.IP.String())
		} else {
			resultList = append(resultList, elem)
		}
	}
	return resultList
}

func flush() {
	now := time.Now()
	for ip, lastSeen := range RemoveSet {
		if now.Sub(lastSeen) > time.Hour {
			delete(RemoveSet, ip)
		}
	}
}

func UpdateAddressList(newList addresslist.PeerList) {
	oldList := AddressList.Contents()
	var resultList addresslist.PeerList
	reduceMap := make(map[string]*addresslist.PeerItem)
	sort.Sort(oldList)
	sort.Sort(newList)
	newList = removeDeadEntries(newList)
	if newList == nil {
		return
	}
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
	if AddressList.Len() == 0 && !Settings.LoopOnEmpty {
		log.Fatal("AddressList ran out of peers")
	}
	if Settings.PingUntilFoundOnStart {
		for !Ping(AddressList.At(0).IP) {}
	}
	KeepAliveLoop()
}

func KeepAliveLoop() {
	log.Println("KeepAliveLoop has been started")
	RemoveSet = make(map[string]time.Time)
	lastFlush := time.After(time.Hour)
	for {
		var peerList addresslist.PeerList
		var err error
		if AddressSet.Len() > 0 {
			peerList, err = GetPeerList(AddressSet.Pop())
			if err != nil { //The peer gets deleted from the list if error
				log.Println(err)
				continue //shit happens but we do not want a defunct list
			}
		} else if AddressList.Len() == 0 {
			if Settings.LoopOnEmpty {
				time.Sleep(5 * time.Second)
				continue
			}
			log.Fatal("AddressList ran out of peers")
		} else {
			index := rand.Intn(AddressList.Len())
			entry := AddressList.At(index)
			peerList, err = GetPeerList(entry.IP)
			AddressList.RemoveAt(index)
			if err != nil { //The peer gets deleted from the list if error
				RemoveSet[entry.IP.String()] = time.Now() //Remember to delete them when merging list
				log.Println(err)
				continue //shit happens but we do not want a defunct list
			}
			AddressList.Add(&addresslist.PeerItem{entry.IP, entry.IndexHash, time.Now()})
		}
		UpdateAddressList(peerList)
		select {
		case <-lastFlush:
			lastFlush = time.After(time.Hour)
			flush()
		default:
			time.Sleep(2 * time.Second)
		}
	}
}

