package main

import(
	"net"
	"net/http"
	"log"
	"io/ioutil"
	"sandwich-go/addresslist"
)

func GetPeerList(address net.IP) (addresslist.PeerList, error) {
	resp, err := http.Get("http://" + address.String() + "/peerlist" + GetPort(address))
	if err != nil {
		log.Println(err)
		return nil, err
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	peerlist := addresslist.Unmarshal(data)
	return peerlist, err
}

