package main

import(
	"fmt"
	"net"
	"log"
)

func PrintFileManifest() {
	for ip, fileList := range FileManifest {
		if fileList == nil {
			log.Println("FileList is null")
			continue
		}
		for _, fileItem := range fileList.List {
			fmt.Println(ip + ":" + fileItem.FileName)
		}
	}
}

func InitializeUserThread() {
	fmt.Println("Hello!")
	go func() {
		for {
			fmt.Print("=>")
			input := make([]string, 3)
			fmt.Scanln(&input[0], &input[1], &input[2])
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

