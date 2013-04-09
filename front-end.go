package main

import(
	"net"
	"log"
	"sort"
	"strings"
)

type IPFilePair struct {
	IP net.IP
  Port string
	FileName string
}

var DownloadQueue chan *IPFilePair

// takes a string and returns true if it should be kept, false otherwise
type Filter interface {
	Filter(string) bool
}

type SimpleFilter string

func (filter SimpleFilter) Filter(toCompare string) bool {
	return strings.Contains(toCompare, string(filter))
}

func ManifestMap() map[string]string {
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
	return fileMap
}

func SortedManifest(fileMap map[string]string) []string {
	fileList := make([]string, 1)
	for fileName := range fileMap {
		fileList = append(fileList, fileName)
	}
	sort.Strings(fileList)
	return fileList
}

func ApplyFilter(fileList []string, filter Filter) []string {
	results := make([]string, 1)
	for _, fileName := range fileList {
		if filter.Filter(fileName) {
			results = append(results, fileName)
		}
	}
	return results
}

func Search(query string) []*IPFilePair {
	fileMap := ManifestMap()
	fileList := SortedManifest(fileMap)
	fileList = ApplyFilter(fileList, SimpleFilter(query))
	result := make([]*IPFilePair, 0, len(fileList))
	for _, fileName := range fileList {
    ip := net.ParseIP(fileMap[fileName])
		result = append(result, &IPFilePair{ip, GetPort(ip), fileName})
	}
	return result
}

func InitializeUserThread() {
	DownloadQueue = make(chan *IPFilePair, 1000)
	go func() {
		for {
			select {
			case filePair := <-DownloadQueue:
				DownloadFile(filePair.IP, filePair.FileName)
			}
		}
	}()
  BuildFileManifest()
  go InitializeFancyStuff()
}

