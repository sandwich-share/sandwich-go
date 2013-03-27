package main

import(
	"os"
	"log"
	"io/ioutil"
	"sandwich-go/addresslist"
	"net"
	"time"
	"sandwich-go/fileindex"
)

var AddressList *addresslist.SafeIPList //Thread safe
var FileIndex *fileindex.SafeFileList

func InitializeAddressList() {
	path := ConfPath("peerlist")
	file, err := os.Open(path)

	pathErr, ok := err.(*os.PathError)
	if err != nil && ok && pathErr.Err.Error() == "no such file or directory" { //Yeah, this is pretty bad but the library 
		// did not expose a constant to represent this

		log.Println(err)
		BootStrap() //This bootstraps us into the network
		return
	} else if err != nil {
		log.Fatal(err)
	}

	data, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}
	ipList := addresslist.Unmarshal(data)
	AddressList = addresslist.New(ipList)
	log.Println("Loaded AddressList from file")
}

//TODO: Make an initializer that does something useful
func InitializeFileIndex() {
	fileList := &fileindex.FileList{}
	fileList.List = append(fileList.List, &fileindex.FileItem{"foo", 123, 123}, &fileindex.FileItem{"bar", 456, 456},
		&fileindex.FileItem{"something", 789, 789}, &fileindex.FileItem{"XXX", 69, 69}, &fileindex.FileItem{"The Devil", 666, 666})
	FileIndex = fileindex.New(fileList)
	FileIndex.RemoveAt(1, 4)
	FileIndex.UpdateHash()
}

//TODO: Make a BootStrap that does something reasonable
func BootStrap() {
	iplist := make(addresslist.PeerList, 1)
	iplist[0] = &addresslist.PeerItem{net.IPv4(127, 0, 0, 1), 0, time.Now()}
	AddressList = addresslist.New(iplist)
	log.Println("Created new peerlist")
}

