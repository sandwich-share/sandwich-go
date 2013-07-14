package main

import (
	"encoding/json"
	"github.com/toqueteos/webbrowser"
	"io/ioutil"
	"log"
	"net"
	"os"
	"regexp"
	"sandwich-go/fileindex"
  "sandwich-go/client"
  "sandwich-go/util"
	"sort"
	"strings"
	"sync/atomic"
	"time"
)

type NetIP net.IP

type IPFilePair struct {
	IP       NetIP
	Port     string
	FileName string
}

type IPFilePairs []*IPFilePair

type FileOrDir struct {
	Type int
	Name string
}

type FileOrDirs []FileOrDir

const (
	DIR  = 0
	FILE = 1
)

func (ip NetIP) MarshalJSON() ([]byte, error) {
	return json.Marshal(net.IP(ip).String())
}

func makeFolders(fileList []*fileindex.FileItem) map[string]FileOrDirs {
	r := regexp.MustCompile("/[^/]+$")
	folders := make(map[string]FileOrDirs, 100)
	folders[""] = make(FileOrDirs, 0, 10)
	var last FileOrDir
	for _, file := range fileList {
		last = FileOrDir{FILE, file.FileName}
		for true {
			next_string := r.ReplaceAllString(last.Name, "")
			if next_string == last.Name {
				if !folders[""].Contains(last) {
					folders[""] = append(folders[""], last)
				}
				break
			}
			next := FileOrDir{DIR, next_string}
			if folders[next.Name] == nil {
				folders[next.Name] = make(FileOrDirs, 0, 5)
			}
			if !folders[next.Name].Contains(last) {
				folders[next.Name] = append(folders[next.Name], last)
			}
			last = FileOrDir{DIR, next_string}
		}
	}
	return folders
}

func (fods FileOrDirs) Contains(fod1 FileOrDir) bool {
	s := fod1.Name
	for _, fod := range fods {
		if fod.Name == s {
			return true
		}
	}
	return false
}

var DownloadQueue chan *IPFilePair
var timeOut *time.Timer

func (ifp IPFilePairs) Len() int {
	return len(ifp)
}

func (ifp IPFilePairs) Swap(i, j int) {
	ifp[i], ifp[j] = ifp[j], ifp[i]
}

func (ifp IPFilePairs) Less(i, j int) bool {
	return ifp[i].FileName < ifp[j].FileName
}

// takes a string and returns true if it should be kept, false otherwise
type Filter interface {
	Filter(IPFilePair) bool
}

type SimpleFilter string

type RegexFilter regexp.Regexp

type FolderFilter string

func (filter FolderFilter) Filter(toCompare IPFilePair) bool {
	path := regexp.QuoteMeta(string(filter))
	regex_string := "^" + path + "[^/]+$"
	regex := regexp.MustCompile(regex_string)
	return (&regex).MatchString(string(toCompare.FileName))
}

func (filter *RegexFilter) Filter(toCompare IPFilePair) bool {
	regex := regexp.Regexp(*filter)
	return (&regex).MatchString(string(toCompare.FileName))
}

func (filter SimpleFilter) Filter(toCompare IPFilePair) bool {
	return strings.Contains(strings.ToLower(toCompare.FileName), strings.ToLower(string(filter)))
}

func ManifestMap() IPFilePairs {
	ManifestLock.Lock()
	if timeOut == nil || !timeOut.Stop() {
		FileManifest = client.CleanManifest(FileManifest)
	}
	atomic.StoreInt32(&IsCleanManifest, 1) //Manifest is clean keep it clean
	fileList := make(IPFilePairs, 0, 100)
	for ipString, tempFileList := range FileManifest {
		ip := net.ParseIP(ipString)
		port := util.GetPort(ip)
		for _, fileItem := range tempFileList.List {
			fileList = append(fileList, &IPFilePair{NetIP(ip), port, fileItem.FileName})
		}
	}
	timeOut = time.AfterFunc(time.Minute, func() {
		atomic.StoreInt32(&IsCleanManifest, 0) //Timed out let the Manifest get dirty
	})
	ManifestLock.Unlock()
	return fileList
}

func ApplyFilter(fileList IPFilePairs, filter Filter) IPFilePairs {
	results := make(IPFilePairs, 0, len(fileList))
	for _, fileName := range fileList {
		if filter.Filter(*fileName) {
			results = append(results, fileName)
		}
	}
	return results
}

func Search(query string, regex bool) (IPFilePairs, error) {
	log.Println("Searching for " + query)
	fileMap := ManifestMap()
	var fileList IPFilePairs
	if regex {
		r, err := regexp.Compile(string(query))
		if err != nil {
			log.Println("Invalid regex")
			return nil, err
		}
		regexFilter := RegexFilter(*r)
		fileList = ApplyFilter(fileMap, &regexFilter)
	} else {
		fileList = ApplyFilter(fileMap, SimpleFilter(query))
	}
	sort.Sort(fileList)
	log.Println("Search completed for " + query)
	return fileList, nil
}

func downloadThread() {
	for {
		select {
		case filePair := <-DownloadQueue:
			log.Println("Downloading file:" + filePair.FileName)
			err := client.DownloadFile(net.IP(filePair.IP), filePair.FileName)
			if err != nil {
				log.Println(err)
			}
			log.Println("Downloading complete")
		}
	}
}

func InitializeUserThread() {
	go downloadThread()
	go downloadThread()
	go downloadThread()
	go downloadThread()
	DownloadQueue = make(chan *IPFilePair, 1000)
	file, err := os.Open(util.ConfPath("manifest-cache.json"))
	if err != nil && os.IsNotExist(err) {
		FileManifest = client.BuildFileManifest()
	} else if err != nil {
		log.Println(err)
		FileManifest = client.BuildFileManifest()
	} else if xml, err := ioutil.ReadAll(file); err != nil {
		log.Println(err)
		FileManifest = client.BuildFileManifest()
		file.Close()
	} else if FileManifest, err = fileindex.UnmarshalManifest(xml); err != nil {
		FileManifest = client.BuildFileManifest()
	} else {
		FileManifest = client.CleanManifest(FileManifest)
		file.Close()
	}
	go InitializeFancyStuff()
	if !Settings.DontOpenBrowserOnStart {
		webbrowser.Open("http://localhost:" + Settings.LocalServerPort)
	}
}
