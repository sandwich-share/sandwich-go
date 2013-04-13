package directory

import(
	"log"
	"io"
	"io/ioutil"
	"hash/crc32"
	"os"
	"time"
	"sync"
	"sync/atomic"
	"path/filepath"
	"sandwich-go/fileindex"
	"code.google.com/p/go.exp/fsnotify"
)

const ChunkSize = 256*1024

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
		file.Close();
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
			fileList = append(fileList, fileItem)
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
	var okToUpdate int32 = 0
	go func() {
		for {
			mutex.Lock()
			lock.Wait()
			for update := false; update; update = !atomic.CompareAndSwapInt32(&okToUpdate, 0, 1) {
				time.Sleep(5 * time.Second)
			}
			fileIndex.UpdateHash()
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
				} else if err == nil {
					fileItem, err := GetFileItemName(fullName)
					if err == nil { //Otherwise the file was deleted before we could create it
						fileIndex.Add(fileItem)
					}
				}
			case event.IsDelete():
				fileIndex.Remove(name)
			case event.IsModify():
				fileIndex.Remove(name)
				fileItem, err := GetFileItemName(fullName)
				if err == nil { //Otherwise the file was deleted before we could create it
					fileIndex.Add(fileItem)
				}
			case event.IsRename():
				fileIndex.Remove(name)
			}
			atomic.CompareAndSwapInt32(&okToUpdate, 1, 0)
			lock.Signal()
		}
		log.Println("Watch loop exited")
	}()
}

