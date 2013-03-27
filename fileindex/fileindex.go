package fileindex

import(
	"log"
	"encoding/json"
	"time"
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
	var fileList *FileList
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

func (list *FileList) Concat(newList *FileList) {
	list.List = append(list.List, newList.List...)
}

func (list *FileList) Copy() *FileList {
	newList := &FileList{list.IndexHash, list.TimeStamp, make([]*FileItem, len(list.List))}
	for i, elem := range list.List {
		newList.List[i] = elem.Copy()
	}
	return newList
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

