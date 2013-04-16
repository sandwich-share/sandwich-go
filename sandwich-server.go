package main

import (
	"compress/gzip"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"bytes"
)

var jsonFileIndexCache []byte
var gzipFileIndexCache []byte
var indexHash uint32
var cacheLock sync.RWMutex

func updateCache() {
	cacheLock.Lock()
	fileList := FileIndex.Contents()
	jsonFileIndexCache = fileList.Marshal()
	indexHash = fileList.IndexHash
	buffer := new(bytes.Buffer)
	gwriter := gzip.NewWriter(buffer)
	_, err := gwriter.Write(jsonFileIndexCache)
	gwriter.Close()
	if err != nil {
		log.Println(err)
	}
	gzipFileIndexCache = buffer.Bytes()
	log.Println("Updated cache")
	cacheLock.Unlock()
}

func makeBWListHandler(function http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ip := net.ParseIP(strings.Split(req.RemoteAddr, ":")[0])
		if !BlackWhiteList.OK(ip) {
			log.Println("Forbid " + ip.String() + " from accessing service")
			http.Error(w, "403 Forbidden", http.StatusForbidden)
		} else {
			function(w, req)
		}
	}
}

func pingHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("pong\n"))
	ip := net.ParseIP(strings.Split(req.RemoteAddr, ":")[0])
	if !AddressList.Contains(ip) {
		AddressSet.Add(ip)
	}
}

func indexForHandler(w http.ResponseWriter, req *http.Request) {
	cacheLock.RLock()
	if indexHash != FileIndex.IndexHash() {
		log.Println("Need to update cache")
		cacheLock.RUnlock()
		updateCache()
		cacheLock.RLock()
	}
	if !strings.Contains(req.Header.Get("Accept-Encoding"), "gzip") {
		w.Header().Set("Content-Type", "text/json")
		w.Write(jsonFileIndexCache)
	} else {
		w.Header().Set("Content-Encoding", "gzip")
		w.Write(gzipFileIndexCache)
	}
	cacheLock.RUnlock()
	log.Println("Sent index")
	ip := net.ParseIP(strings.Split(req.RemoteAddr, ":")[0])
	if !AddressList.Contains(ip) {
		AddressSet.Add(ip)
	}
}

func peerListHandler(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "text/json")
	addressList := AddressList.Contents() //Gets a copy of the underlying IPSlice
	addressList = append(addressList, MakeLocalPeerItem())
	json := addressList.Marshal()
	writer.Write(json)
	AddressSet.Add(net.ParseIP(strings.Split(request.RemoteAddr, ":")[0]))
}

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func makeGzipHandler(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			fn(w, r)
			return
		} else {
			w.Header().Set("Content-Encoding", "gzip")
			gz, err := gzip.NewWriterLevel(w, gzip.DefaultCompression)
			if err != nil {
				http.Error(w, "", http.StatusInternalServerError)
				return
			}
			defer gz.Close()
			fn(gzipResponseWriter{Writer: gz, ResponseWriter: w}, r)
		}
	}
}

func InitializeServer() error {
	mux := http.NewServeMux()
	fileHandler, _ := http.StripPrefix("/files/", http.FileServer(http.Dir(SandwichPath))).(http.HandlerFunc)
	mux.HandleFunc("/peerlist", makeBWListHandler(makeGzipHandler(peerListHandler)))
	mux.HandleFunc("/ping", makeBWListHandler(pingHandler))
	mux.HandleFunc("/fileindex", makeBWListHandler(indexForHandler))
	mux.HandleFunc("/files/", makeBWListHandler(fileHandler))

	log.Printf("About to listen on %s.\n", GetPort(LocalIP))
	srv := &http.Server{Handler: mux, Addr: GetPort(LocalIP)}
	err := srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
		return err
	}
	return nil
}

