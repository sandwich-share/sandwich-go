package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
)

var cache IPFilePairs
var peerCache map[string]FileOrDirs
var peerCacheIP string

type IPPort struct {
	IP   string
	Port string
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	search := r.FormValue("search")
	regex := r.FormValue("regex") == "true"
	start, _ := strconv.Atoi(r.FormValue("start"))
	var step int
	step, _ = strconv.Atoi(r.FormValue("step"))
	if start == 0 {
		var err error
		cache, err = Search(search, regex)
		if err != nil {
			fmt.Fprintf(w, "Invalid regex")
		}
	}
	var end int
	if step != 0 {
		end = start + step
		if len(cache) < end {
			end = len(cache)
		}
	} else {
		end = len(cache)
	}
	if start > len(cache) {
		fmt.Fprintf(w, "")
	} else {
		json_res, _ := json.Marshal(cache[start:end])
		w.Write(json_res)
	}
}

func peerHandler(w http.ResponseWriter, r *http.Request) {
	peer_ip := r.FormValue("peer")
	path := r.FormValue("path")
	if peerCacheIP != peer_ip {
		x := FileManifest[peer_ip]
		if x != nil {
			peerCache = makeFolders(x.List)
			peerCacheIP = peer_ip
		} else {
			fmt.Fprintf(w, "[]")
			return
		}
	}
	step, _ := strconv.Atoi(r.FormValue("step"))
	start, _ := strconv.Atoi(r.FormValue("start"))
	files := peerCache[path]
	var end int
	if step != 0 {
		end = start + step
		if len(files) < end {
			end = len(files)
		}
	} else {
		end = len(files)
	}
	if start > len(files) {
		fmt.Fprintf(w, "")
	} else {
		json_res, _ := json.Marshal(files[start:end])
		w.Write(json_res)
	}
}

func peersHandler(w http.ResponseWriter, r *http.Request) {
	peerList := make([]IPPort, 0, 10)
	for _, peer := range AddressList.Contents() {
		peerList = append(peerList, IPPort{peer.IP.String(), GetPort(peer.IP)})
	}
	json_res, _ := json.Marshal(peerList)
	w.Write(json_res)
}

func downloadHandler(w http.ResponseWriter, r *http.Request) {
	var file_type int
	ip := NetIP(net.ParseIP(r.FormValue("ip")))
	if r.FormValue("type") == "0" {
		file_type = DIR
	} else {
		file_type = FILE
	}
	queue := make(FileOrDirs, 0, 10)
	queue = append(queue, FileOrDir{file_type, r.FormValue("file")})
	for len(queue) > 0 {
		last_i := len(queue) - 1
		last := queue[last_i]
		queue = queue[0:last_i]
		if last.Type == DIR {
			queue = append(queue, peerCache[last.Name]...)
		} else {
			DownloadQueue <- &IPFilePair{IP: ip, FileName: last.Name}
		}
	}

}

func killHandler(w http.ResponseWriter, r *http.Request) {
	Shutdown()
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/index.html")
}

func InitializeFancyStuff() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", homeHandler)
	mux.HandleFunc("/search", searchHandler)
	mux.HandleFunc("/peer", peerHandler)
	mux.HandleFunc("/peers", peersHandler)
	mux.HandleFunc("/download", downloadHandler)
	mux.HandleFunc("/kill", killHandler)
	mux.Handle("/static/", http.FileServer(http.Dir("./")))
	srv := &http.Server{Handler: mux, Addr: ":" + Settings.LocalServerPort}
	srv.ListenAndServe()
}
