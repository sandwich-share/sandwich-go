package server

import (
	"bytes"
	"compress/gzip"
	"io"
	"log"
	"net"
	"net/http"
  "sandwich-go/addresslist"
  "sandwich-go/fileindex"
  "sandwich-go/util"
	"strings"
	"sync"
	"time"
)

var addressList *addresslist.SafeIPList        //Thread safe
var addressSet *addresslist.AddressSet         //Thread safe
var blackWhiteList *addresslist.BlackWhiteList //Thread safe
var cacheLock sync.RWMutex
var fileIndex *fileindex.SafeFileList          //Thread safe
var gzipFileIndexCache []byte
var indexHash uint32
var jsonFileIndexCache []byte
var localIP net.IP

func updateCache() {
	cacheLock.Lock()
	fileList := fileIndex.Contents()
	fileList.TimeStamp = time.Now()
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
		if !blackWhiteList.OK(ip) {
			log.Println("Forbid " + ip.String() + " from accessing service")
			http.Error(w, "403 Forbidden", http.StatusForbidden)
		} else {
			function(w, req)
		}
	}
}

func versionHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(util.VERSION + "\n"))
}

func pingHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("pong\n"))
	ip := net.ParseIP(strings.Split(req.RemoteAddr, ":")[0])
	if !addressList.Contains(ip) {
		addressSet.Add(ip)
	}
}

func indexForHandler(w http.ResponseWriter, req *http.Request) {
	cacheLock.RLock()
	if indexHash != fileIndex.IndexHash() {
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
	if !addressList.Contains(ip) {
		addressSet.Add(ip)
	}
}

func peerListHandler(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "text/json")
	ipSlice := addressList.Contents() //Gets a copy of the underlying IPSlice
	ipSlice = append(ipSlice, util.MakePeerItem(localIP, fileIndex))
	json := ipSlice.Marshal()
	writer.Write(json)
	addressSet.Add(net.ParseIP(strings.Split(request.RemoteAddr, ":")[0]))
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

func Initialize(newAddressList *addresslist.SafeIPList,
    newAddressSet *addresslist.AddressSet,
    newBlackWhiteList *addresslist.BlackWhiteList,
    newFileIndex *fileindex.SafeFileList,
    newLocalIP net.IP) {
  addressList = newAddressList
  addressSet = newAddressSet
  blackWhiteList = newBlackWhiteList
  fileIndex = newFileIndex
  localIP = newLocalIP
	mux := http.NewServeMux()

	mux.HandleFunc("/peerlist", makeBWListHandler(makeGzipHandler(peerListHandler)))
	mux.HandleFunc("/ping", makeBWListHandler(pingHandler))
	mux.HandleFunc("/fileindex", makeBWListHandler(indexForHandler))
	mux.HandleFunc("/version", versionHandler)
	mux.Handle("/files/", http.StripPrefix("/files/", http.FileServer(http.Dir(util.SandwichPath))).(http.HandlerFunc))

	log.Printf("About to listen on %s.\n", util.GetPort(localIP))
	srv := &http.Server{Handler: mux, Addr: util.GetPort(localIP)}
	err := srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
