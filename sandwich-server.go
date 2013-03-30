package main

import(
	"net/http"
	"log"
	"net/url"
	"strings"
)

// this will eventually resize, but right now you can't have more than 500 peers.
var defaultPeerListSize = 500

func pingHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("pong\n"))
}

func indexForHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/json")
	listCopy := FileIndex.Contents()
	w.Write(listCopy.Marshal())
	log.Println("Sent index")
}

func peerListHandler(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "text/json")
	addressList := AddressList.Contents() //Gets a copy of the underlying IPSlice
	addressList = append(addressList, MakeLocalPeerItem())
	log.Println("Copied list")
	json := addressList.Marshal()
	writer.Write(json)
}

func fileHandler(writer http.ResponseWriter, request *http.Request) {
	var err error
	query := request.URL.RawQuery
	split := strings.Split(query, "=")
	request.URL, err = url.Parse(split[1])
	if err != nil {
		log.Fatal(err)
	}
	http.FileServer(http.Dir(SandwichPath)).ServeHTTP(writer, request)
}

func InitializeServer() {
	http.HandleFunc("/peerlist/", peerListHandler)
	http.HandleFunc("/ping/", pingHandler)
	http.HandleFunc("/indexfor/", indexForHandler)
	http.HandleFunc("/file", fileHandler)

	log.Printf("About to listen on 8000. Go to http://127.0.0.1:8000/")
	if Settings.ListenLocal {
		go http.ListenAndServe(":8000", nil)
	}
	err := http.ListenAndServe(GetPort(LocalIP), nil)
	if err != nil {
		log.Fatal(err)
	}
}

