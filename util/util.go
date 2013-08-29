package util

import (
	"crypto/md5"
	"io/ioutil"
	"log"
	"net"
	"path/filepath"
	"sandwich-go/addresslist"
	"sandwich-go/fileindex"
	"strconv"
	"strings"
	"time"
)

const SandwichDirName = "sandwich"
const ConfigDirName = "conf"

var HomePath string
var SandwichPath string
var ConfigPath string

type Version struct {
	Major, Minor, Patch int
	Commit              string
}

func ParseVersion(raw string) *Version {
	retVal := new(Version)
	parsed := strings.Split(raw, ":")
	retVal.Commit = parsed[1]
	parsed = strings.Split(parsed[0], ".")
	major, _ := strconv.ParseInt(parsed[0], 10, 32)
	minor, _ := strconv.ParseInt(parsed[1], 10, 32)
	retVal.Major = int(major)
	retVal.Minor = int(minor)
	if len(parsed) == 3 {
		patch, _ := strconv.ParseInt(parsed[2], 10, 32)
		retVal.Patch = int(patch)
	}
	return retVal
}

func (ver *Version) Equal(com *Version) bool {
	return ver.Major == com.Major && ver.Minor == com.Minor && ver.Patch == com.Patch
}

func (ver *Version) Less(com *Version) bool {
	if ver.Major < com.Major {
		return true
	} else if ver.Major == com.Major {
		if ver.Minor < com.Minor {
			return true
		} else if ver.Minor == com.Minor && ver.Patch < com.Patch {
			return true
		}
	}
	return false
}

func (ver *Version) Greater(com *Version) bool {
	return !ver.Less(com) && !ver.Equal(com)
}

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

func MakePeerItem(ip net.IP, fileIndex *fileindex.SafeFileList) *addresslist.PeerItem {
	return &addresslist.PeerItem{ip, fileIndex.IndexHash(), time.Now()}
}
