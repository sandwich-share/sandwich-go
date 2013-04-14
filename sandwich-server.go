package main

import (
	"compress/gzip"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
)

func pingHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("pong\n"))
	ip := net.ParseIP(strings.Split(req.RemoteAddr, ":")[0])
	if !AddressList.Contains(ip) {
		AddressSet.Add(ip)
	}
}

func indexForHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/json")
	listCopy := FileIndex.Contents()
	w.Write(listCopy.Marshal())
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
	mux.HandleFunc("/peerlist/", makeGzipHandler(peerListHandler))
	mux.HandleFunc("/ping/", pingHandler)
	mux.HandleFunc("/fileindex/", makeGzipHandler(indexForHandler))
	mux.Handle("/files/", http.StripPrefix("/files/", http.FileServer(http.Dir(SandwichPath))))

	log.Printf("About to listen on %s.\n", GetPort(LocalIP))
	srv := &http.Server{Handler: mux, Addr: GetPort(LocalIP)}
	err := srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
		return err
	}
	return nil
}
