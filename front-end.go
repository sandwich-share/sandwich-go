package main

import(
	"fmt"
)

func PrintFileManifest() {
	for _, fileList := range FileManifest {
		for _, fileItem := range fileList.List {
			fmt.Println(fileItem.FileName)
		}
	}
}

