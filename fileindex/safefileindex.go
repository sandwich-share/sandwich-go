package fileindex

import(
	"sync"
	"time"
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

func (list *SafeFileList) Contents() *FileList {
	list.m.RLock()
	retVal := list.fileList.Copy()
	list.m.RUnlock()
	return retVal
}

func (list *SafeFileList) Copy(fileList *FileList) {
	list.m.Lock()
	list.fileList = fileList
	list.m.Unlock()
}

func (list *SafeFileList) TimeStamp() time.Time {
	list.m.RLock()
	retVal := list.fileList.TimeStamp
	list.m.RUnlock()
	return retVal
}

func (list *SafeFileList) IndexHash() uint32 {
	list.m.RLock()
	retVal := list.fileList.IndexHash
	list.m.RUnlock()
	return retVal
}

func (list *SafeFileList) Remove(itemList ...string) {
	list.m.Lock()
	list.fileList.Remove(itemList...)
	list.m.Unlock()
}

func (list *SafeFileList) RemoveAt(indexs ...int) {
	list.m.Lock()
	list.fileList.RemoveAt(indexs...)
	list.m.Unlock()
}

func (list *SafeFileList) UpdateHash() {
	list.m.Lock()
	list.fileList.UpdateHash()
	list.m.Unlock()
}

