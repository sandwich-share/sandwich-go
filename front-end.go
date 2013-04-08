package main

import(
	"fmt"
	"net"
	"log"
	"os"
	"strings"
)

func PrintFileManifest() {
	for _, fileList := range FileManifest {
		if fileList == nil {
			log.Println("FileList is null")
			continue
		}
		for _, fileItem := range fileList.List {
			fmt.Println(fileItem.FileName)
		}
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
			default:
				fmt.Println("Input should be in the form: =>command args")
			}
		}
	}()
}

