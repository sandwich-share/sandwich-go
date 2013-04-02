package fileindex

import(
	"net"
)

type FileManifest map[string]*FileList

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

