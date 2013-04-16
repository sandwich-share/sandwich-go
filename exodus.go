package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"sandwich-go/addresslist"
	"sandwich-go/directory"
	"sandwich-go/fileindex"
	"sandwich-go/settings"
	"strings"
	"sync"
)

var AddressList *addresslist.SafeIPList //Thread safe
var AddressSet *addresslist.AddressSet	//Thread safe
var FileIndex *fileindex.SafeFileList	//Thread safe
var BlackWhiteList *addresslist.BlackWhiteList //Thread safe
var FileManifest fileindex.FileManifest //NOT THREAD SAFE
var ManifestLock = new(sync.Mutex)
var IsCleanManifest int32
var LocalIP net.IP
var Settings *settings.Settings
var Whitelist = []*addresslist.IPRange{&addresslist.IPRange{net.ParseIP("129.22.0.0"), net.ParseIP("129.22.255.255")},
	&addresslist.IPRange{net.ParseIP("173.241.224.0"), net.ParseIP("173.241.239.255")},
	&addresslist.IPRange{net.ParseIP("127.0.0.0"), net.ParseIP("127.255.255.255")}}

func InitializeAddressList() error {
	err := GetLocalIP()
	if err != nil {
		log.Fatal(err)
		return err
	}

	path := ConfPath("peerlist")
	file, err := os.Open(path)

	if err != nil && os.IsNotExist(err) {
		if !Settings.DoNotBootStrap {
			BootStrap() //This bootstraps us into the network
		} else {
			var ipList addresslist.PeerList
			AddressList = addresslist.New(ipList)
			log.Println("Created empty AddressList")
		}

		return nil
	} else if err != nil {
		log.Fatal(err)
	} else {
		data, err := ioutil.ReadAll(file)
		if err != nil {
			log.Fatal(err)
			return err
		}
		ipList := addresslist.Unmarshal(data)
		AddressList = addresslist.New(ipList)
		log.Println("Loaded AddressList from file")
	}

	return err
}

func GetLocalIP() error {
	resp, err := http.Get("http://curlmyip.com")
	if err != nil {
		log.Println(err)
		conn, err := net.Dial("tcp", "google.com:80")
		if err != nil {
			log.Fatal(err)
			return err
		}
		LocalIP = net.ParseIP(strings.Split(conn.LocalAddr().String(), ":")[0])
		err = conn.Close()
		if err != nil {
			log.Fatal(err)
		}
	} else {
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println("Failed to read body")
			return err
		} else {
			LocalIP = net.ParseIP(strings.TrimSpace(string(b)))
		}
	}
	log.Println("Local IP is: " + LocalIP.String())
	return nil
}

func InitializePaths() error {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
		return err
	}
	HomePath = usr.HomeDir
	SandwichPath = filepath.Join(HomePath, SandwichDirName)
	ConfigPath = ConfigDirName
	_, err = os.Stat(SandwichPath)
	if err != nil && os.IsNotExist(err) {
		err = os.MkdirAll(SandwichPath, os.ModePerm)
		if err != nil {
			log.Fatal(err)
			return err
		}
		log.Println("Created: " + SandwichPath)
		return nil
	} else if err != nil {
		log.Fatal(err)
		return err
	}
	_, err = os.Stat(ConfigPath)
	if err != nil && os.IsNotExist(err) {
		err = os.MkdirAll(ConfigPath, os.ModePerm)
		if err != nil {
			log.Fatal(err)
			return err
		}
		log.Println("Created: " + ConfigPath)
		return nil
	} else if err != nil {
		log.Fatal(err)
		return err
	}
	return nil
}

func InitializeSettings() error {
	var err error
	settings.SettingsPath = ConfPath("settings.xml")
	Settings, err = settings.Load()
	if err != nil {
		Settings = &settings.Settings{}
	}
	if Settings.LocalServerPort == "" {
		Settings.LocalServerPort = "9001"
	}
	Settings.Save()
	if Settings.SandwichDirName != "" {
		SandwichPath = Settings.SandwichDirName
		_, err = os.Stat(SandwichPath)
		if err != nil && os.IsNotExist(err) {
			err = os.MkdirAll(SandwichPath, os.ModePerm)
			if err != nil {
				log.Fatal(err)
				return err
			}
			log.Println("Created: " + SandwichPath)
		} else if err != nil {
			log.Fatal(err)
			return err
		}
	}
	return nil
}

func InitializeFileIndex() error {
	FileIndex = fileindex.New(nil)
	directory.CheckSumMaxSize = Settings.CheckSumMaxSize
	directory.StartWatch(SandwichPath, FileIndex)
	return nil
}

//TODO: Make a BootStrap that does something reasonable
func BootStrap() error {
	iplist := make(addresslist.PeerList, 1)
	var rawIP string
	fmt.Print("Please enter an IP address for bootstrap\n=>")
	_, err := fmt.Scanln(&rawIP)
	if err != nil {
		log.Println(err)
		return BootStrap()
	}
	addrs := net.ParseIP(rawIP)
	iplist[0] = &addresslist.PeerItem{addrs, FileIndex.IndexHash(), FileIndex.TimeStamp()}
	AddressList = addresslist.New(iplist)
	log.Println("Created new peerlist")
	return nil
}

func Shutdown() {
	ioutil.WriteFile(ConfPath("peerlist"), AddressList.Contents().Marshal(), os.ModePerm)
	ioutil.WriteFile(ConfPath("blackwhitelist.xml"), BlackWhiteList.Marshal(), os.ModePerm)
	err := ioutil.WriteFile(ConfPath("manifest-cache.json"), FileManifest.Marshal(), os.ModePerm)
	if err != nil {
		log.Println(err)
	}
	Settings.Save()
	os.Exit(0)
}

func main() {

	runtime.GOMAXPROCS(2)

	AddressSet = addresslist.NewAddressSet()
	FileManifest = fileindex.NewFileManifest()
	file, err := os.Open(ConfPath("blackwhitelist.xml"))
	if err != nil {
		BlackWhiteList = addresslist.NewBWList(Whitelist)
	} else if data, err := ioutil.ReadAll(file); err != nil {
		file.Close()
		BlackWhiteList = addresslist.NewBWList(Whitelist)
	} else if BlackWhiteList, err = addresslist.UnmarshalBWList(data, Whitelist); err != nil {
		file.Close()
		BlackWhiteList = addresslist.NewBWList(Whitelist)
	}

	err = InitializePaths()
	if err != nil {
		return
	}
	err = InitializeSettings()
	if err != nil {
		return
	}
	err = InitializeFileIndex()
	if err != nil {
		return
	}
	err = InitializeAddressList()
	if err != nil {
		return
	}
	go InitializeKeepAliveLoop()
	InitializeUserThread()
	if !Settings.WriteLogToScreen {
		logWriter, err := os.Create("log")
		if err != nil {
			log.Fatal(err)
			return
		}
		log.SetOutput(logWriter)
	}
	err = InitializeServer()
	if err != nil {
		return
	}
}
