package main

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
	"encoding/json"
)

var cache []*IPFilePair

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
		if (len(cache) - 1) < end {
			end = len(cache) - 1
		}
	} else {
		end = len(cache)-1
	}
	if start > len(cache)-1 {
		fmt.Fprintf(w, "")
	} else {
		json_res, _ := json.Marshal(cache[start:end])
		w.Write(json_res)
	}
}

func downloadHandler(w http.ResponseWriter, r *http.Request) {
	DownloadQueue <- &IPFilePair{IP: NetIP(net.ParseIP(r.FormValue("ip"))), FileName: r.FormValue("file")}
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
	mux.HandleFunc("/download", downloadHandler)
	mux.HandleFunc("/kill", killHandler)
	mux.Handle("/static/", http.FileServer(http.Dir("./")))
	srv := &http.Server{Handler: mux, Addr: ":" + Settings.LocalServerPort}
	srv.ListenAndServe()
}
