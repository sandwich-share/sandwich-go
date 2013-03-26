package main

import(
	"net"
	"net/http"
	"log"
	"io/ioutil"
	"sandwich-go/addresslist"
)

func GetPeerList(address net.IP) addresslist.IPSlice {
	resp, err := http.Get("http://" + address.String() + "/peerlist" + GetPort(address))
	if err != nil {
		log.Println(err)
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
	}
	return addresslist.FromString(string(data))
}

