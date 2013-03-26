package main

import(
	"os"
	"log"
	"io"
	"addresslist"
	"net"
)

var AddressList *addresslist.SafeIPList

func InitializeAddressList() {
	path := ConfPath("addresslist")
	file, err := Open(path)
	if err != nil && err != ErrNotExist {
		//The address list does not exist
		log.Println("The addresslist must be created")
		BootStrap() //This bootstraps us into the network
		return
	} else if err != nil {
		log.Fatal(err)
	}
	data, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}
	iplist := FromString(data)
	AddressList = addresslist.New(iplist)
}

//TODO: Make a BootStrap that does something reasonable
func BootStrap() {
	iplist := make(addresslist.IPSlice, 1)
	iplist[0] = net.ParseIP("127.0.0.1")
	AddressList = addresslist.New(iplist)
}

