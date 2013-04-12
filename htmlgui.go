package main

import (
	"fmt"
	"html/template"
	"net"
	"net/http"
	"strconv"
)

var templates = template.Must(template.ParseFiles("templates/index.html", "templates/query_result.html"))

var cache []*IPFilePair

func homeHandler(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "index.html", nil)
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	search := r.FormValue("search")
	regex := r.FormValue("regex") == "true"
	start, _ := strconv.Atoi(r.FormValue("start"))
	if start == 0 {
		var err error
		cache, err = Search(search, regex)
		if err != nil {
			fmt.Fprintf(w, "Invalid regex")
		}
	}
	end := start + 100
	if (len(cache) - 1) < end {
		end = len(cache) - 1
	}
	if start > len(cache)-1 {
		fmt.Fprintf(w, "")
	} else {
		templates.ExecuteTemplate(w, "query_result.html", cache[start:end])
	}
}

func downloadHandler(w http.ResponseWriter, r *http.Request) {
	DownloadQueue <- &IPFilePair{IP: net.ParseIP(r.FormValue("ip")), FileName: r.FormValue("file")}
}

func killHandler(w http.ResponseWriter, r *http.Request) {
	Shutdown()
}

func InitializeFancyStuff() {
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/search", searchHandler)
	http.HandleFunc("/download", downloadHandler)
	http.HandleFunc("/kill", killHandler)
	http.Handle("/static/", http.FileServer(http.Dir("./")))
	http.ListenAndServe(":8000", nil)
}
