package fileindex

import(
	"net"
	"encoding/json"
	"log"
)

type FileManifest map[string]*FileList

func NewFileManifest() FileManifest {
	return make(FileManifest)
}

func UnmarshalManifest(data []byte) (FileManifest, error) {
	var manifest FileManifest
	err := json.Unmarshal(data, &manifest)
	if err != nil {
		return nil, err
	}
	return manifest, err
}

func (man FileManifest) Marshal() []byte {
	data, err := json.Marshal(&man)
	if err != nil {
		log.Println(err)
	}
	return data
}

func (man FileManifest) Get(address net.IP) (bool, *FileList) {
	result, ok := man[address.String()]
	return ok, result
}

func (man FileManifest) Put(address net.IP, list *FileList) {
	man[address.String()] = list
}

func (man FileManifest) Delete(address net.IP) {
	delete(man, address.String())
}

