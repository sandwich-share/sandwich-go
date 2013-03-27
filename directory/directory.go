package directory

import(
	"log"
	"io/ioutil"
	"hash/crc32"
	"os"
	"path"
	"path/filepath"
	"sandwich-go/fileindex"
//	"code.google.com/p/go.exp/fsnotify"
)

var SandwichPath string

func GetFileChecksum(file *os.File) uint32 {
	data, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}
	return crc32.ChecksumIEEE(data)
}

func GetFileItem(filePath string, info os.FileInfo) *fileindex.FileItem {
	fullName := path.Join(filePath, info.Name())
	file, err := os.Open(fullName)
	if err != nil {
		log.Fatal(err)
	}
	relName, err := filepath.Rel(SandwichPath, fullName)
	if err != nil {
		log.Fatal(err)
	}
	return &fileindex.FileItem{relName, uint64(info.Size()), GetFileChecksum(file)}
}

func BuildFileList(filePath, dir string) []*fileindex.FileItem {
	var fileList []*fileindex.FileItem
	newPath := path.Join(filePath, dir)
	infoList, err := ioutil.ReadDir(newPath)
	if err != nil {
		log.Fatal(err)
	}
	for _, fileInfo := range infoList {
		if fileInfo.IsDir() {
			fileList = append(fileList, BuildFileList(newPath, fileInfo.Name())...)
		} else {
			fileList = append(fileList, GetFileItem(newPath, fileInfo))
		}
	}
	return fileList
}

func BuildFileIndex(dir string) *fileindex.FileList {
	SandwichPath = dir
	fileList := &fileindex.FileList{}
	fileList.List = BuildFileList("", dir)
	fileList.UpdateHash()
	return fileList
}

