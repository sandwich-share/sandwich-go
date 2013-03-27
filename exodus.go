package main

import(
	"os"
	"log"
	"io/ioutil"
	"sandwich-go/addresslist"
	"net"
	"time"
)

var AddressList *addresslist.SafeIPList //Thread safe

func InitializeAddressList() {
	path := ConfPath("peerlist")
	file, err := os.Open(path)

	pathErr, ok := err.(*os.PathError)
	if err != nil && ok && pathErr.Err.Error() == "no such file or directory" { //Yeah, this is pretty bad but the library 
		// did not expose a constant to represent this

		log.Println(err)
		BootStrap() //This bootstraps us into the network
		return
	} else if err != nil {
		log.Fatal(err)
	}

	data, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}
	ipList := addresslist.Unmarshal(data)
	AddressList = addresslist.New(ipList)
	log.Println("Loaded AddressList from file")
}

//TODO: Make a BootStrap that does something reasonable
func BootStrap() {
	iplist := make(addresslist.PeerList, 1)
	iplist[0] = &addresslist.PeerItem{net.IPv4(127, 0, 0, 1), 0, time.Now()}
	AddressList = addresslist.New(iplist)
	log.Println("Created new peerlist")
}

