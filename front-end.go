package main

import (
	"github.com/toqueteos/webbrowser"
	"io/ioutil"
	"log"
	"net"
	"os"
	"regexp"
	"sandwich-go/fileindex"
	"sort"
	"strings"
	"sync/atomic"
	"time"
)

type IPFilePair struct {
	IP       net.IP
	Port     string
	FileName string
}

type IPFilePairs []*IPFilePair

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
		CleanManifest()
	}
	atomic.StoreInt32(&IsCleanManifest, 1) //Manifest is clean keep it clean
	fileList := make(IPFilePairs, 0, 100)
	for ipString, tempFileList := range FileManifest {
		ip := net.ParseIP(ipString)
		port := GetPort(ip)
		for _, fileItem := range tempFileList.List {
			fileList = append(fileList, &IPFilePair{ip, port, fileItem.FileName})
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

func InitializeUserThread() {
	DownloadQueue = make(chan *IPFilePair, 1000)
	go func() {
		for {
			select {
			case filePair := <-DownloadQueue:
				log.Println("Downloading file:" + filePair.FileName)
				err := DownloadFile(filePair.IP, filePair.FileName)
				if err != nil {
					log.Println(err)
				}
				log.Println("Downloading complete")
			}
		}
	}()
	file, err := os.Open(ConfPath("manifest-cache.json"))
	if err != nil && os.IsNotExist(err) {
		BuildFileManifest()
	} else if err != nil {
		log.Println(err)
		BuildFileManifest()
	} else if xml, err := ioutil.ReadAll(file); err != nil {
		log.Println(err)
		BuildFileManifest()
		file.Close()
	} else if FileManifest, err = fileindex.UnmarshalManifest(xml); err != nil {
		FileManifest = fileindex.NewFileManifest()
		BuildFileManifest()
	} else {
		CleanManifest()
		file.Close()
	}
	go InitializeFancyStuff()
	if !Settings.DontOpenBrowserOnStart {
		webbrowser.Open("http://localhost:" + Settings.LocalServerPort)
	}
}
