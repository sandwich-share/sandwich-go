package main

import "net/http"
import "log"

// this will eventually resize, but right now you can't have more than 500 peers.
var defaultPeerListSize = 500

func pingHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("pong\n"))
}

func indexForHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("This is an example server.\n"))
}

func peerListHandlerGenerator(peerMetaChan chan chan string) func(http.ResponseWriter, *http.Request) {

	return func (w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/plain")

		log.Print("url handler received request for peer list")

		peerChan := make(chan string)

		peerMetaChan <- peerChan

		log.Print("sent request for peer list...")

		for addr := <-peerChan; len(addr) != 0; {
			log.Print("recieved value from monitor...")
			w.Write([]byte(addr))
		}

		log.Print("url handler finished sending peer list")

		peerChan <- "foo"
	}
}

// TODO: needs to be changed to start up communication on the received channel with yet another goroutine
// otherwise we are blocking per request.
func peerListMonitor(peerMetaChan chan chan string) {

	log.Print("Starting up peerListMonitor")

	peerList := make([]string, defaultPeerListSize)

	peerList = append(peerList, "foo", "bar", "baz")

	select {
	case peerRequest := <-peerMetaChan:
		log.Print("Peer List monitor recieved request for peer list")
		for _,v := range peerList {
			log.Print("Sending value from monitor...")
			peerRequest <-v
		}

		//requesterIP := <-peerRequest
		//<-peerRequest
	}
}

func main() {

	peerMetaChan := make(chan chan string)

	peerListHandler := peerListHandlerGenerator(peerMetaChan)

	go peerListMonitor(peerMetaChan)

	http.HandleFunc("/peerlist", peerListHandler)
	http.HandleFunc("/ping", pingHandler)
	http.HandleFunc("/indexfor", indexForHandler)

	log.Printf("About to listen on 8000. Go to http://127.0.0.1:8000/")
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal(err)
	}
}

