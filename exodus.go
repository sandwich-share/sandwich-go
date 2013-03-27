package main

import(
	"os"
	"os/user"
	"path"
	"log"
	"io/ioutil"
	"sandwich-go/addresslist"
	"net"
	"sandwich-go/fileindex"
	"sandwich-go/directory"
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

func InitializePaths() {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	HomePath = usr.HomeDir
	SandwichPath = path.Join(HomePath, SandwichDirName)
}

func InitializeFileIndex() {
	FileIndex = fileindex.New(directory.BuildFileIndex(SandwichPath))
}

//TODO: Make a BootStrap that does something reasonable
func BootStrap() {
	iplist := make(addresslist.PeerList, 1)
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		log.Fatal(err)
	}
	iplist[0] = &addresslist.PeerItem{net.ParseIP(addrs[3].String()), FileIndex.IndexHash(), FileIndex.TimeStamp()}
	AddressList = addresslist.New(iplist)
	log.Println("Created new peerlist")
}

