package main

import(
	"fmt"
	"strings"
	"net"
	"log"
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
			fmt.Print("=>")
			var input string
			_, err := fmt.Scanln(&input)
			if err != nil {
				fmt.Println(err)
			}
			parsedInput := strings.Split(input, " ")
			if len(parsedInput) < 1 {
				fmt.Println("Input should be in the form: =>command argument")
				continue
			}
			switch(parsedInput[0]) {
			case "print":
				PrintFileManifest()
			case "update":
				BuildFileManifest()
			case "get":
				if len(parsedInput) != 3 {
					fmt.Println("Input should be in the form: =>command argument")
					continue
				}
				err := DownloadFile(net.ParseIP(parsedInput[1]), parsedInput[2])
				if err != nil {
					fmt.Println(err)
				}
			default:
				fmt.Println("Input should be in the form: =>command args")
			}
		}
	}()
}

