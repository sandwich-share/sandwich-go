package main

import (
	"github.com/toqueteos/webbrowser"
	"log"
	"net"
	"regexp"
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

var DownloadQueue chan *IPFilePair
var timeOut *time.Timer

// takes a string and returns true if it should be kept, false otherwise
type Filter interface {
	Filter(string) bool
}

type SimpleFilter string

type RegexFilter string

func (filter RegexFilter) Filter(toCompare string) bool {
	r, err := regexp.Compile(string(filter))
	if err != nil {
		log.Println("Invalid regex")
		return false
	}
	return r.MatchString(string(toCompare))
}

func (filter SimpleFilter) Filter(toCompare string) bool {
	return strings.Contains(strings.ToLower(toCompare), strings.ToLower(string(filter)))
}

func ManifestMap() map[string]string {
	ManifestLock.Lock()
	if timeOut == nil || !timeOut.Stop() {
		CleanManifest()
	}
	atomic.StoreInt32(&IsCleanManifest, 1) //Manifest is clean keep it clean
	fileMap := make(map[string]string)
	for ip, fileList := range FileManifest {
		if fileList == nil {
			log.Println("FileList is null")
			continue
		}
		for _, fileItem := range fileList.List {
			fileMap[fileItem.FileName] = ip
		}
	}
	timeOut = time.AfterFunc(time.Minute, func() {
		atomic.StoreInt32(&IsCleanManifest, 0) //Timed out let the Manifest get dirty
	})
	ManifestLock.Unlock()
	return fileMap
}

func SortedManifest(fileMap map[string]string) []string {
	fileList := make([]string, 0, len(fileMap))
	for fileName := range fileMap {
		fileList = append(fileList, fileName)
	}
	sort.Strings(fileList)
	return fileList
}

func ApplyFilter(fileList []string, filter Filter) []string {
	results := make([]string, 0, len(fileList))
	for _, fileName := range fileList {
		if filter.Filter(fileName) {
			results = append(results, fileName)
		}
	}
	return results
}

func Search(query string, regex bool) []*IPFilePair {
	log.Println("Searching for " + query)
	fileMap := ManifestMap()
	fileList := SortedManifest(fileMap)
	if regex {
		fileList = ApplyFilter(fileList, RegexFilter(query))
	} else {
		fileList = ApplyFilter(fileList, SimpleFilter(query))
	}
	result := make([]*IPFilePair, 0, len(fileList))
	for _, fileName := range fileList {
		ip := net.ParseIP(fileMap[fileName])
		result = append(result, &IPFilePair{ip, GetPort(ip), fileName})
	}
	log.Println("Search completed for " + query)
	return result
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
	BuildFileManifest()
	go InitializeFancyStuff()
	webbrowser.Open("http://localhost:8000")
}
