package main

import(
	"os"
	"log"
	"io/ioutil"
	"sandwich-go/addresslist"
	"net"
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
	iplist := addresslist.FromString(string(data))
	AddressList = addresslist.New(iplist)
	log.Println("Loaded AddressList from file")
}

//TODO: Make a BootStrap that does something reasonable
func BootStrap() {
	iplist := make(addresslist.IPSlice, 1)
	iplist[0] = net.IPv4(127, 0, 0, 1)
	AddressList = addresslist.New(iplist)
}

