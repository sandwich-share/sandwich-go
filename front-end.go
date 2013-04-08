package main

import(
	"fmt"
	"net"
	"log"
	"os"
	"sort"
	"strings"
)

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

func PrintFileManifest() {
	fileMap := ManifestMap()
	for _, fileName := range SortedManifest(fileMap) {
		fmt.Println(fileMap[fileName] + " " + fileName)
	}
}

func Search(query string) {
	fileMap := ManifestMap()
	fileList := SortedManifest(fileMap)
	fileList = ApplyFilter(fileList, SimpleFilter(query))
	for _, fileName := range fileList {
		fmt.Println(fileMap[fileName] + " " + fileName)
	}
}

func InitializeUserThread() {
	fmt.Println("Hello!")
	go func() {
		for {
			rdbuf := make([]byte, 1)
			inputstr := ""
			fmt.Print("=>")
			for {
				readLength, err := os.Stdin.Read(rdbuf)
				if err != nil || readLength == 0 {
					fmt.Println("Shit fucked up");
					return
				}

				inputchar := rdbuf[0]
				if (inputchar == '\n') {
					break
				} else {
					inputstr += string(inputchar)
				}
			}

			input := make([]string, 3)

			splitstring := strings.Split(inputstr, " ")
			for i, substring := range splitstring {
				if i < 2 {
					input[i] = substring
				} else {
					if input[2] != "" {
						input[2] += " "+substring
					} else {
						input[2] = substring
					}
				}
			}

			if len(input) < 1 {
				fmt.Println("Input should be in the form: =>command argument")
				continue
			}
			switch(input[0]) {
			case "print":
				PrintFileManifest()
			case "update":
				BuildFileManifest()
			case "search":
				Search(input[1])
			case "get":
				if len(input) != 3 {
					fmt.Println("Input should be in the form: =>command argument")
					fmt.Printf("Length is: %d\n", len(input))
					continue
				}
				err := DownloadFile(net.ParseIP(input[1]), input[2])
				if err != nil {
					fmt.Println(err)
				}
			case "exit":
				Shutdown()
			default:
				fmt.Println("Input should be in the form: =>command args")
			}
		}
	}()
}

