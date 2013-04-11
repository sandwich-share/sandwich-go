package main

import (
	"crypto/md5"
	"io/ioutil"
	"log"
	"net"
	"path/filepath"
	"sandwich-go/addresslist"
	"strconv"
	"time"
)

const SandwichDirName = "sandwich"
const ConfigDirName = "conf"

var HomePath string
var SandwichPath string
var ConfigPath string

// Quick way to make a path for a config file
func ConfPath(newPath string) string {
	return filepath.Join(ConfigPath, newPath)
}

func ComputeHash(address net.IP) int {
	var port uint16
	hash := md5.New()
	hash.Write([]byte(address))
	result := hash.Sum(nil)
	port = (uint16(result[0]) + uint16(result[3])) << 8
	port += uint16(result[1]) + uint16(result[2])
	if port < 1024 {
		port += 1024
	}
	return int(port)
}

func GetPort(address net.IP) string {
	if address.Equal(net.IPv4(127, 0, 0, 1)) {
		return ":8000" //It would be silly to do something funny on your own computer
	}
	port := ComputeHash(address)
	return ":" + strconv.Itoa(port)
}

func Save(list addresslist.PeerList) error {
	json := list.Marshal()
	err := ioutil.WriteFile(ConfPath("peerlist"), json, 0777)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func MakeLocalPeerItem() *addresslist.PeerItem {
	return &addresslist.PeerItem{LocalIP, FileIndex.IndexHash(), time.Now()}
}
