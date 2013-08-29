package directory

import (
	"code.google.com/p/go.exp/fsnotify"
	"hash/crc32"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sandwich-go/fileindex"
	"sync"
	"time"
	"unicode/utf8"
)

const ChunkSize = 256 * 1024

var SandwichPath string
var watcher *fsnotify.Watcher
var CheckSumMaxSize int64

func GetFileChecksum(file *os.File) uint32 {
	fileInfo, err := file.Stat()
	if err != nil {
		log.Println(err)
		return 0
	}
	if fileInfo.Size() > CheckSumMaxSize && CheckSumMaxSize != -1 {
		return 0
	}
	hasher := crc32.New(crc32.MakeTable(crc32.Castagnoli))
	byteBuf := make([]byte, ChunkSize)
	byteChan := make(chan []byte, ChunkSize)
	go func() {
		for val := range byteChan {
			hasher.Write(val)
		}
	}()
	for done := false; !done; {
		numRead, err := file.Read(byteBuf)
		if err != nil && err != io.EOF {
			log.Println(err)
		}
		if numRead < ChunkSize {
			byteBuf = byteBuf[:numRead]
			done = true
		}
		byteChan <- byteBuf
	}
	close(byteChan)
	return hasher.Sum32()
}

func GetFileItemName(name string) (*fileindex.FileItem, error) {
	info, err := os.Stat(name)
	_, ok := err.(*os.PathError)
	if err != nil && ok {
		log.Println(err)
		return nil, err
	}
	dirPath, _ := filepath.Split(name)
	return GetFileItem(dirPath, info)
}

func GetFileItem(filePath string, info os.FileInfo) (*fileindex.FileItem, error) {
	var checksum uint32
	fullName := filepath.Join(filePath, info.Name())
	if CheckSumMaxSize <= info.Size() && CheckSumMaxSize != 0 {
		file, err := os.Open(fullName)
		_, ok := err.(*os.PathError)
		if err != nil && ok {
			log.Println(err)
			return nil, err
		}
		checksum = GetFileChecksum(file)
		file.Close()
	} else {
		checksum = 0
	}
	relName, err := filepath.Rel(SandwichPath, fullName)
	if err != nil {
		log.Fatal(err)
	}
	fileItem := &fileindex.FileItem{filepath.ToSlash(relName), uint64(info.Size()), checksum}
	return fileItem, err
}

func BuildFileList(filePath, dir string) []*fileindex.FileItem {
	watcher.Watch(filepath.Join(filePath, dir))
	var fileList []*fileindex.FileItem
	newPath := filepath.Join(filePath, dir)
	infoList, err := ioutil.ReadDir(newPath)
	if err != nil {
		log.Fatal(err)
	}
	for _, fileInfo := range infoList {
		if fileInfo.IsDir() {
			fileList = append(fileList, BuildFileList(newPath, fileInfo.Name())...)
		} else {
			fileItem, err := GetFileItem(newPath, fileInfo)
			_, ok := err.(*os.PathError)
			if err != nil && ok {
				log.Println(err)
				continue
			}
			if utf8.ValidString(fileItem.FileName) {
				fileList = append(fileList, fileItem)
			} else {
				log.Printf("%s is not properly utf-8 encoded", fileItem.FileName)
			}
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
	var err error
	watcher, err = fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	fileList := BuildFileIndex(dir)
	fileIndex.Copy(fileList)

	mutex := new(sync.Mutex)
	lock := sync.NewCond(mutex)
	go func() {
		for {
			mutex.Lock()
			lock.Wait()
			fileIndex.UpdateHash()
			log.Println("Updated hash")
			time.Sleep(1 * time.Second)
			mutex.Unlock()
		}
	}()

	go func() {
		defer watcher.Close() //This loop should run as long as the program is running
		for event := range watcher.Event {
			name, err := filepath.Rel(SandwichPath, event.Name)
			fullName := filepath.Join(SandwichPath, name)
			if err != nil {
				log.Fatal(err)
			}
			switch {
			case event.IsCreate():
				info, err := os.Stat(fullName)
				if err == nil && info.IsDir() {
					fileList := BuildFileList("", fullName)
					fileIndex.Concat(fileList)
					log.Println(fullName + " was added to the manifest.")
				} else if err == nil {
					fileItem, err := GetFileItemName(fullName)
					if err == nil { //Otherwise the file was deleted before we could create it
						if utf8.ValidString(fileItem.FileName) {
							fileList.Add(fileItem)
						} else {
							log.Println("Hey bra, you cannot have non-utf8 encoded file names")
						}
					}
				}
			case event.IsDelete():
				fileIndex.Remove(name)
				log.Println(name + " was removed from the manifest.")
			case event.IsModify():
				fileIndex.Remove(name)
				fileItem, err := GetFileItemName(fullName)
				if err == nil { //Otherwise the file was deleted before we could create it
					if utf8.ValidString(fileItem.FileName) {
						fileList.Add(fileItem)
						log.Println(fullName + " was added to the manifest.")
					} else {
						log.Println("Hey bra, you cannot have non-utf8 encoded file names")
					}
				}
			case event.IsRename():
				fileIndex.Remove(name)
				log.Println(name + " was removed from the manifest.")
			}
			lock.Signal()
		}
		log.Println("Watch loop exited")
	}()
}
