package fileindex

import(
	"log"
	"encoding/json"
	"time"
	"encoding/binary"
	"hash/crc32"
)

type FileItem struct {
	FileName string
	Size uint64
	CheckSum uint32
}

type FileList struct {
	IndexHash uint32
	TimeStamp time.Time
	List []*FileItem
}

func (item *FileItem) Copy() *FileItem {
	return &FileItem{item.FileName, item.Size, item.CheckSum}
}

func Unmarshal(jsonList []byte) *FileList {
	fileList := &FileList{}
	err := json.Unmarshal(jsonList, fileList)
	if err != nil {
		log.Println(err)
	}
	return fileList
}

func (list *FileList) Marshal() []byte {
	jsonList, err := json.Marshal(list)
	if err != nil {
		log.Println(err)
	}
	return jsonList
}

func (list *FileList) At(index int) *FileItem {
	return list.List[index]
}

func (list *FileList) Add(item *FileItem) {
	list.List = append(list.List, item)
}

func (list *FileList) Concat(newList []*FileItem) {
	list.List = append(list.List, newList...)
}

func (list *FileList) Copy() *FileList {
	newList := &FileList{list.IndexHash, list.TimeStamp, make([]*FileItem, len(list.List))}
	for i, elem := range list.List {
		newList.List[i] = elem.Copy()
	}
	return newList
}

func (list *FileList) Remove(newList ...string) {
	subtract := 0
	i := 0
	for j, elem := range list.List {
		list.List[j - subtract] = elem
		if i < len(newList) && newList[i] == list.List[j].FileName {
			i++
			subtract++
		}
	}
	list.List = list.List[:len(list.List) - subtract]
}

func (list *FileList) RemoveAt(indexList ...int) {
	subtract := 0
	i := 0
	for j := 0; j < len(list.List); j++ {
		elem := list.List[j]
		list.List[j - subtract] = elem
		if i < len(indexList) && indexList[i] == j {
			i++
			subtract++
		}
	}
	list.List = list.List[:len(list.List) - subtract]
}

func (list *FileList) UpdateHash() {
	data := make([]byte, 1)
	for _, item := range list.List {
		buffer := make([]byte, 8)
		data = append(data, []byte(item.FileName)...)
		binary.PutUvarint(buffer, item.Size)
		data = append(data, buffer...)
		binary.PutUvarint(buffer, uint64(item.CheckSum))
		data = append(data, buffer...)
	}
	list.IndexHash = crc32.ChecksumIEEE(data)
	list.TimeStamp = time.Now()
}

