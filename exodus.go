package main

import(
	"fmt"
	"os"
	"os/user"
	"path"
	"log"
	"io/ioutil"
	"strings"
	"sandwich-go/addresslist"
	"net"
	"sandwich-go/fileindex"
	"sandwich-go/directory"
	"sandwich-go/settings"
)

var AddressList *addresslist.SafeIPList //Thread safe
var FileIndex *fileindex.SafeFileList //Thread safe
var FileManifest fileindex.FileManifest //NOT THREAD SAFE
var LocalIP net.IP
var Settings *settings.Settings

func InitializeAddressList() {
	GetLocalIP()
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
	} else {

		data, err := ioutil.ReadAll(file)
		if err != nil {
			log.Fatal(err)
		}
		ipList := addresslist.Unmarshal(data)
		AddressList = addresslist.New(ipList)
		log.Println("Loaded AddressList from file")
	}
}

func GetLocalIP() {
	conn, err := net.Dial("tcp", "google.com:80")
	if err != nil {
		log.Fatal(err)
	}
	err = conn.Close()
	if err != nil {
		log.Fatal(err)
	}
	LocalIP = net.ParseIP(strings.Split(conn.LocalAddr().String(), ":")[0])
	log.Println("Local IP is: " + LocalIP.String())
}

func InitializePaths() {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	HomePath = usr.HomeDir
	SandwichPath = path.Join(HomePath, SandwichDirName)
}

func InitializeSettings() {
	var err error
	settings.SettingsPath = ConfPath("settings.xml")
	Settings, err = settings.Load()
	if err != nil {
		Settings = &settings.Settings{}
		Settings.Save()
	}
}

func InitializeFileIndex() {
	FileIndex = fileindex.New(directory.BuildFileIndex(SandwichPath))
	directory.StartWatch(SandwichPath, FileIndex)
}

//TODO: Make a BootStrap that does something reasonable
func BootStrap() {
	iplist := make(addresslist.PeerList, 1)
	var rawIP string
	fmt.Print("Please enter an IP address for bootstrap\n=>")
	_, err := fmt.Scanln(&rawIP)
	if err != nil {
		log.Println(err)
		BootStrap()
		return
	}
	addrs := net.ParseIP(rawIP)
	iplist[0] = &addresslist.PeerItem{addrs, FileIndex.IndexHash(), FileIndex.TimeStamp()}
	AddressList = addresslist.New(iplist)
	log.Println("Created new peerlist")
}

func main() {

	InitializePaths()
	InitializeSettings()
	InitializeFileIndex()
	InitializeAddressList()
	go InitializeKeepAliveLoop()
	InitializeServer()
}

