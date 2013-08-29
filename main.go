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
	"sandwich-go/client"
	"sandwich-go/directory"
	"sandwich-go/fileindex"
	"sandwich-go/frontend"
	"sandwich-go/server"
	"sandwich-go/settings"
	"sandwich-go/util"
	"strings"
)

var AddressList *addresslist.SafeIPList        //Thread safe
var AddressSet *addresslist.AddressSet         //Thread safe
var FileIndex *fileindex.SafeFileList          //Thread safe
var BlackWhiteList *addresslist.BlackWhiteList //Thread safe
var FileManifest fileindex.FileManifest        //NOT THREAD SAFE
var LocalIP net.IP
var Settings *settings.Settings

//var Whitelist = []*addresslist.IPRange{&addresslist.IPRange{net.ParseIP("129.22.0.0"), net.ParseIP("129.22.255.255")}, // CWRUNET
//	&addresslist.IPRange{net.ParseIP("173.241.224.0"), net.ParseIP("173.241.239.255")}, // Hessler
//	&addresslist.IPRange{net.ParseIP("127.0.0.0"), net.ParseIP("127.255.255.255")},     // IPv4 Subnet
//	&addresslist.IPRange{net.ParseIP("192.5.109.0"), net.ParseIP("192.5.109.255")},     // CWRUNET-C0
//	&addresslist.IPRange{net.ParseIP("192.5.110.0"), net.ParseIP("192.5.110.255")},     // CWRUNET-C1
//	&addresslist.IPRange{net.ParseIP("192.5.111.0"), net.ParseIP("192.5.111.255")},     // CWRUNET-C2
//	&addresslist.IPRange{net.ParseIP("192.5.112.0"), net.ParseIP("192.5.112.255")},     // CWRUNET-C3
//	&addresslist.IPRange{net.ParseIP("192.5.113.0"), net.ParseIP("192.5.113.255")}}     // CWRUNET-C4
var Whitelist = []*addresslist.IPRange{&addresslist.IPRange{net.ParseIP("0.0.0.0"), net.ParseIP("255.255.255.255")}}

func initializeAddressList() error {
	err := getLocalIP()
	if err != nil {
		log.Fatal(err)
		return err
	}

	path := util.ConfPath("peerlist")
	file, err := os.Open(path)

	if err != nil && os.IsNotExist(err) {
		if !Settings.DoNotBootStrap {
			bootStrap() //This bootstraps us into the network
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

func getLocalIP() error {
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

func initializePaths() error {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
		return err
	}
	util.HomePath = usr.HomeDir
	if Settings.SandwichDirName != "" {
		util.SandwichPath = Settings.SandwichDirName
	} else {
		util.SandwichPath = filepath.Join(util.HomePath, util.SandwichDirName)
	}
	_, err = os.Stat(util.SandwichPath)
	if err != nil && os.IsNotExist(err) {
		err = os.MkdirAll(util.SandwichPath, os.ModePerm)
		if err != nil {
			log.Fatal(err)
			return err
		}
		log.Println("Created: " + util.SandwichPath)
		return nil
	} else if err != nil {
		log.Fatal(err)
		return err
	}
	return nil
}

func initializeSettings() error {
	var err error
	settings.SettingsPath = util.ConfPath("settings.xml")
	Settings, err = settings.Load()
	if err != nil {
		Settings = &settings.Settings{}
	}
	if Settings.LocalServerPort == "" {
		Settings.LocalServerPort = "9001"
	}
	Settings.Save()
	return nil
}

func initializeFileIndex() error {
	FileIndex = fileindex.New(nil)
	directory.CheckSumMaxSize = Settings.CheckSumMaxSize
	directory.StartWatch(util.SandwichPath, FileIndex)
	return nil
}

func bootStrap() error {
	iplist := make(addresslist.PeerList, 1)
	var host string
	fmt.Print("Please enter a host name for bootstrap\n=>")
	_, err := fmt.Scanln(&host)
	if err != nil {
		log.Println(err)
		return bootStrap()
	}
	addrs, err := net.LookupIP(host)
	if err != nil {
		log.Println(err)
		return bootStrap()
	}
	iplist[0] = &addresslist.PeerItem{addrs[0], FileIndex.IndexHash(), FileIndex.TimeStamp()}
	AddressList = addresslist.New(iplist)
	log.Println("Created new peerlist")
	return nil
}

func Shutdown() {
	util.Save(AddressList.Contents())
	ioutil.WriteFile(util.ConfPath("blackwhitelist.xml"), BlackWhiteList.Marshal(), os.ModePerm)
	err := ioutil.WriteFile(util.ConfPath("manifest-cache.json"), FileManifest.Marshal(), os.ModePerm)
	if err != nil {
		log.Println(err)
	}
	os.Exit(0)
}

func main() {

	log.Println(util.VERSION)

	runtime.GOMAXPROCS(runtime.NumCPU())

	util.ConfigPath = util.ConfigDirName //We need our conf directory to do anything else
	_, err := os.Stat(util.ConfigPath)
	if err != nil && os.IsNotExist(err) {
		err = os.MkdirAll(util.ConfigPath, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("Created: " + util.ConfigPath)
	} else if err != nil {
		log.Fatal(err)
	}
	AddressSet = addresslist.NewAddressSet()
	FileManifest = fileindex.NewFileManifest()
	file, err := os.Open(util.ConfPath("blackwhitelist.xml"))
	if err != nil {
		BlackWhiteList = addresslist.NewBWList(Whitelist)
	} else if data, err := ioutil.ReadAll(file); err != nil {
		file.Close()
		BlackWhiteList = addresslist.NewBWList(Whitelist)
	} else if BlackWhiteList, err = addresslist.UnmarshalBWList(data, Whitelist); err != nil {
		file.Close()
		BlackWhiteList = addresslist.NewBWList(Whitelist)
	}

	err = initializeSettings()
	if err != nil {
		return
	}
	err = initializePaths()
	if err != nil {
		return
	}
	err = initializeFileIndex()
	if err != nil {
		return
	}
	err = initializeAddressList()
	if err != nil {
		return
	}
	go client.Initialize(AddressList, AddressSet, BlackWhiteList, LocalIP, Settings)
	frontend.Initialize(AddressList, Settings, Shutdown)
	if !Settings.WriteLogToScreen {
		logWriter, err := os.Create("log")
		if err != nil {
			log.Fatal(err)
			return
		}
		log.SetOutput(logWriter)
	}
	server.Initialize(AddressList, AddressSet, BlackWhiteList, FileIndex, LocalIP)
	if err != nil {
		return
	}
}
