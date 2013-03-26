package main

import(
	"log"
	"io"
	"path"
)

const SandwichFolderName = "Sandwich"
const ConfigFolderName = "conf"

// Quick way to make a path for a config file
func ConfPath(newPath string) string {
	return path.Join("conf", newPath)
}


