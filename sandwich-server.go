package main

import(
	"net"
	"net/http"
	"log"
	"net/url"
	"strings"
	"io"
	"compress/gzip"
)

// this will eventually resize, but right now you can't have more than 500 peers.
var defaultPeerListSize = 500

func pingHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("pong\n"))
	AddressSet.Add(net.ParseIP(strings.Split(req.RemoteAddr, ":")[0]))
}

func indexForHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/json")
	listCopy := FileIndex.Contents()
	w.Write(listCopy.Marshal())
	log.Println("Sent index")
	AddressSet.Add(net.ParseIP(strings.Split(req.RemoteAddr, ":")[0]))
}

func peerListHandler(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "text/json")
	addressList := AddressList.Contents() //Gets a copy of the underlying IPSlice
	addressList = append(addressList, MakeLocalPeerItem())
	log.Println("Copied list")
	json := addressList.Marshal()
	writer.Write(json)
	AddressSet.Add(net.ParseIP(strings.Split(request.RemoteAddr, ":")[0]))
}

func fileHandler(writer http.ResponseWriter, request *http.Request) {
	var err error
	query := request.URL.RawQuery
	split := strings.Split(query, "=")
	request.URL, err = url.Parse(split[1])
	if err != nil {
		log.Fatal(err)
		return
	}
	http.FileServer(http.Dir(SandwichPath)).ServeHTTP(writer, request)
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
	http.HandleFunc("/peerlist/", makeGzipHandler(peerListHandler))
	http.HandleFunc("/ping/", pingHandler)
	http.HandleFunc("/indexfor/", makeGzipHandler(indexForHandler))
	http.HandleFunc("/file", fileHandler)

	log.Printf("About to listen on %s.\n", GetPort(LocalIP))
	err := http.ListenAndServe(GetPort(LocalIP), nil)
	if err != nil {
		log.Fatal(err)
		return err
	}
	return nil
}

