package client

import (
	"bufio"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sandwich-go/addresslist"
	"sandwich-go/fileindex"
  "sandwich-go/settings"
  "sandwich-go/util"
	"sort"
	"strings"
	"time"
)

func DownloadFile(address net.IP, filePath string) error {
	if !blackWhiteList.OK(address) {
		return illegalIPError
	}
	conn, err := net.DialTimeout("tcp", address.String() + util.GetPort(address),
    2*time.Minute)
	if err != nil {
		return err
	}
	defer conn.Close()

	url := url.URL{}
	url.Path = filePath
	request, err := http.NewRequest("GET", "/files/" + url.String(), nil)
	if err != nil {
		return err
	}

	err = request.Write(conn)
	if err != nil {
		return err
	}
	buffer := bufio.NewReader(conn)
	response, err := http.ReadResponse(buffer, request)
	if err != nil {
		return err
	}
	buffer = bufio.NewReader(response.Body)
	dirPath, _ := filepath.Split(filePath)
	err = os.MkdirAll(path.Join(util.SandwichPath, dirPath), os.ModePerm)
	if err != nil {
		return err
	}
	file, err := os.Create(path.Join(util.SandwichPath, filePath))
	if err != nil {
		file.Close()
		return err
	}
	byteBuf := make([]byte, 4*1024*1024)
	for done := false; !done; {
		numRead, err := buffer.Read(byteBuf)
		if err == io.EOF {
			done = true
		} else if err != nil {
			file.Close()
			return err
		}
		_, err = file.Write(byteBuf[:numRead])
		if err != nil {
			file.Close()
			return err
		}
	}
	err = file.Close()
	if err != nil {
		return err
	}
	if err != nil {
		return err
	}
	return err
}

func GetFileIndex(address net.IP) (*fileindex.FileList, error) {
	resp, err := get(address, "/fileindex")
	if err != nil {
		log.Println(err)
		return nil, err
	}
	fileList := fileindex.Unmarshal(resp)
	return fileList, err
}

func GetVersion(address net.IP) (string, error) {
	resp, err := get(address, "/version")
	if err != nil {
		log.Println(err)
	}
	return strings.Split(string(resp), "\n")[0], err
}

func BuildFileManifest() (fileManifest fileindex.FileManifest) {
	peerList := addressList.Contents()
	out1 := make(chan net.IP)
	out2 := make(chan net.IP)
	out3 := make(chan net.IP)
	out4 := make(chan net.IP)
	in1 := make(chan fileindex.FileManifest)
	in2 := make(chan fileindex.FileManifest)
	in3 := make(chan fileindex.FileManifest)
	in4 := make(chan fileindex.FileManifest)
	go getFileIndexLoop(out1, in1)
	go getFileIndexLoop(out2, in2)
	go getFileIndexLoop(out3, in3)
	go getFileIndexLoop(out4, in4)
	for _, item := range peerList {
		select {
		case out1 <- item.IP:
		case out2 <- item.IP:
		case out3 <- item.IP:
		case out4 <- item.IP:
		}
	}
	close(out1)
	close(out2)
	close(out3)
	close(out4)
	for i := 0; i < 4; i++ {
		select {
		case manifest := <-in1:
			fileManifest = aggregate(manifest, fileManifest)
		case manifest := <-in2:
			fileManifest = aggregate(manifest, fileManifest)
		case manifest := <-in3:
			fileManifest = aggregate(manifest, fileManifest)
		case manifest := <-in4:
			fileManifest = aggregate(manifest, fileManifest)
		}
	}
	log.Println("File index created")
  return
}

func GetPeerList(address net.IP) (addresslist.PeerList, error) {
	resp, err := get(address, "/peerlist")
	if err != nil {
		log.Println(err)
		return nil, err
	}
	peerlist := addresslist.Unmarshal(resp)
	return peerlist, err
}

func UpdateAddressList(newList addresslist.PeerList) {
	oldList := addressList.Contents()
	var resultList addresslist.PeerList
	reduceMap := make(map[string]*addresslist.PeerItem)
	sort.Sort(oldList)
	sort.Sort(newList)
	newList = removeDeadEntries(newList)
	if newList == nil {
		return
	}
	for _, elem := range oldList {
		if elem.IP.Equal(localIP) {
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
		if elem.IP.Equal(localIP) {
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
	util.Save(resultList)
	addressList.Copy(resultList)
}

func Ping(address net.IP) bool {
	resp, err := get(address, "/ping")
	if err != nil {
		log.Println(err)
		return false
	}
	if string(resp) == "pong\n" {
		return true
	}
	return false
}

func Initialize(newAddressList *addresslist.SafeIPList,
    newAddressSet *addresslist.AddressSet,
    newBlackWhiteList *addresslist.BlackWhiteList,
    newLocalIp net.IP,
    newSettings *settings.Settings) {
  addressList = newAddressList
  addressSet = newAddressSet
  blackWhiteList = newBlackWhiteList
  localIP = newLocalIp
  sandwichSettings = newSettings
	if addressList.Len() == 0 && !sandwichSettings.LoopOnEmpty {
		log.Fatal("AddressList ran out of peers")
	}
	if sandwichSettings.PingUntilFoundOnStart {
		for !Ping(addressList.At(0).IP) {
		}
	}
	keepAliveLoop()
}

func CleanManifest(fileManifest fileindex.FileManifest) fileindex.FileManifest {
	addressList := addressList.Contents()
	for _, entry := range addressList {
		fileIndex, ok := fileManifest[entry.IP.String()]
		if ok && (entry.IndexHash == fileIndex.IndexHash ||
        fileIndex.TimeStamp.After(entry.LastSeen)) {
			continue
		} else {
			log.Println("Updated entry")
			index, err := GetFileIndex(entry.IP)
			if err != nil {
				continue
			}
			fileManifest[entry.IP.String()] = index
		}
	}
  return fileManifest
}
