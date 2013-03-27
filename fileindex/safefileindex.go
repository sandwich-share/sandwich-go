package fileindex

import(
	"sync"
	"hash/crc32"
	"time"
	"encoding/binary"
)

type SafeFileList struct {
	fileList *FileList
	m sync.RWMutex
}

func New(list *FileList) *SafeFileList {
	var mutex sync.RWMutex
	return &SafeFileList{list, mutex}
}

func (list *SafeFileList) At(index int) *FileItem {
	list.m.RLock()
	retVal := list.fileList.List[index].Copy()
	list.m.RUnlock()
	return retVal
}

func (list *SafeFileList) Add(item *FileItem) {
	list.m.Lock()
	list.fileList.Add(item)
	list.m.Unlock()
}

func (list *SafeFileList) Concat(newList *FileList) {
	list.m.Lock()
	list.fileList.Concat(newList)
	list.m.Unlock()
}

func (list *SafeFileList) Copy() *FileList {
	list.m.RLock()
	retVal := list.fileList.Copy()
	list.m.RUnlock()
	return retVal
}

func (list *SafeFileList) RemoveAt(indexs ...int) {
	list.m.Lock()
	list.fileList.RemoveAt(indexs...)
	list.m.Unlock()
}

func (list *SafeFileList) UpdateHash() {
	list.m.RLock()
	data := make([]byte, 1)
	for _, item := range list.fileList.List {
		buffer := make([]byte, 8)
		data = append(data, []byte(item.FileName)...)
		binary.PutUvarint(buffer, item.Size)
		data = append(data, buffer...)
		binary.PutUvarint(buffer, uint64(item.CheckSum))
		data = append(data, buffer...)
	}
	list.fileList.IndexHash = crc32.ChecksumIEEE(data)
	list.fileList.TimeStamp = time.Now()
	list.m.RUnlock()
}

