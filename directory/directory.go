package directory

import(
	"log"
	"io/ioutil"
	"hash/crc32"
	"os"
	"path"
	"path/filepath"
	"sandwich-go/fileindex"
	"code.google.com/p/go.exp/fsnotify"
)

var SandwichPath string

func GetFileChecksum(file *os.File) uint32 {
	data, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}
	return crc32.ChecksumIEEE(data)
}

func GetFileItemName(name string) (*fileindex.FileItem, error) {
	fullName := path.Join(SandwichPath, name)
	info, err := os.Stat(fullName)
	pathErr, ok := err.(*os.PathError)
	if err != nil && ok && pathErr.Err.Error() == "no such file or directory" {
		log.Println(err)
		return nil, err
	}
	return GetFileItem(SandwichPath, info), nil
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

func StartWatch(dir string, fileIndex *fileindex.SafeFileList) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for event := range watcher.Event {
			name, err := filepath.Rel(SandwichPath, event.Name)
			if err != nil {
				log.Fatal(err)
			}
			log.Println(name)
			switch {
			case event.IsCreate():
				log.Println("Created")
				fileItem, err := GetFileItemName(name)
				if err == nil { //Otherwise the file was deleted before we could create it
					fileIndex.Add(fileItem)
				}
			case event.IsDelete():
				log.Println("Deleted")
				fileIndex.Remove(name)
			case event.IsModify():
				log.Println("Modified")
				fileIndex.Remove(name)
				fileItem, err := GetFileItemName(name)
				if err == nil { //Otherwise the file was deleted before we could create it
					fileIndex.Add(fileItem)
				}
			case event.IsRename():
				log.Println("Renamed")
				fileIndex.Remove(name)
			}
			fileIndex.UpdateHash()
		}
		log.Println("Watch loop exited")
		watcher.Close() //This loop should run as long as the program is running
	}()

	err = watcher.Watch(dir)
	if err != nil {
		log.Fatal(err)
	}
}

