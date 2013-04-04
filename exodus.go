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
	"runtime"
)

var AddressList *addresslist.SafeIPList //Thread safe
var AddressSet *addresslist.AddressSet //Thread safe
var FileIndex *fileindex.SafeFileList //Thread safe
var FileManifest fileindex.FileManifest //NOT THREAD SAFE
var LocalIP net.IP
var Settings *settings.Settings

func InitializeAddressList() {
	GetLocalIP()
	path := ConfPath("peerlist")
	file, err := os.Open(path)

	pathErr, ok := err.(*os.PathError)
	if err != nil && ok && pathErr.Err.Error() == "no such file or directory" && !Settings.DoNotBootStrap {
		//Yeah, this is pretty bad but the library 
		// did not expose a constant to represent this

		log.Println(err)
		BootStrap() //This bootstraps us into the network
		return
	} else if err != nil && ok && pathErr.Err.Error() == "no such file or directory" {
		var ipList addresslist.PeerList
		AddressList = addresslist.New(ipList)
		log.Println("Created empty AddressList")
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
	LocalIP = net.ParseIP(strings.Split(conn.LocalAddr().String(), ":")[0])
	err = conn.Close()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Local IP is: " + LocalIP.String())
}

func InitializePaths() {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	HomePath = usr.HomeDir
	SandwichPath = path.Join(HomePath, SandwichDirName)
	ConfigPath = ConfigDirName
	_, err = os.Stat(SandwichPath)
	pathErr, ok := err.(*os.PathError)
	if err != nil && ok && pathErr.Err.Error() == "no such file or directory" {
		err = os.MkdirAll(SandwichPath, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("Created: " + SandwichPath)
	} else if err != nil {
		log.Fatal(err)
	}
	_, err = os.Stat(ConfigPath)
	pathErr, ok = err.(*os.PathError)
	if err != nil && ok && pathErr.Err.Error() == "no such file or directory" {
		err = os.MkdirAll(ConfigPath, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("Created: " + ConfigPath)
	} else if err != nil {
		log.Fatal(err)
	}
}

func InitializeSettings() {
	var err error
	settings.SettingsPath = ConfPath("settings.xml")
	Settings, err = settings.Load()
	if err != nil {
		Settings = &settings.Settings{}
	}
	Settings.Save()
}

func InitializeFileIndex() {
	FileIndex = fileindex.New(nil)
	directory.CheckSumMaxSize = Settings.CheckSumMaxSize
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

	runtime.GOMAXPROCS(2)

	AddressSet = addresslist.NewAddressSet()
	FileManifest = fileindex.NewFileManifest()
	InitializePaths()
	InitializeSettings()
	InitializeFileIndex()
	InitializeAddressList()
	go InitializeKeepAliveLoop()
	if !Settings.DisableInterface {
		InitializeUserThread()
	}
	if !Settings.WriteLogToScreen {
		logWriter, err := os.Create("log")
		if err != nil {
			log.Fatal(err)
		}
		log.SetOutput(logWriter)
	}
	InitializeServer()
}

