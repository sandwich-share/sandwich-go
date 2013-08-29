package frontend

import (
	"code.google.com/p/go.net/websocket"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sandwich-go/client"
	"sandwich-go/util"
	"strconv"
	"time"
)

type IPPort struct {
	IP   string
	Port string
}

var cache IPFilePairs
var peerCache map[string]FileOrDirs
var peerCacheIP string

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
		if timeOut == nil || !timeOut.Stop() {
			fileManifest = client.CleanManifest(fileManifest)
		}
		x := fileManifest[peer_ip]
		if x != nil {
			peerCache = makeFolders(x.List)
			peerCacheIP = peer_ip
		} else {
			fmt.Fprintf(w, "[]")
			return
		}
		timeOut = time.AfterFunc(time.Minute, func() {})
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

func localVersionHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, util.VERSION)
}

func killHandler(w http.ResponseWriter, r *http.Request) {
	shutdown(fileManifest)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/index.html")
}

func settingsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		dirname := r.FormValue("dir")
		localport := r.FormValue("port")
		dontopenbrowser := r.FormValue("openBrowser") == "false"
		new_settings := sandwichSettings
		new_settings.SandwichDirName = dirname
		new_settings.LocalServerPort = localport
		new_settings.DontOpenBrowserOnStart = dontopenbrowser
		new_settings.Save()
	} else {
		json_res, _ := json.Marshal(sandwichSettings)
		w.Write(json_res)
	}
}

func writePeers() {
	peerList := make([]IPPort, 0, 10)
	for _, peer := range addressList.Contents() {
		peerList = append(peerList, IPPort{peer.IP.String(), util.GetPort(peer.IP)})
	}
	json_res, _ := json.Marshal(peerList)
	peerHub.broadcast <- string(json_res)
}

func InitializeFancyStuff() {
	go peerHub.run()
	mux := http.NewServeMux()
	mux.HandleFunc("/", homeHandler)
	mux.HandleFunc("/search", searchHandler)
	mux.HandleFunc("/peer", peerHandler)
	mux.HandleFunc("/download", downloadHandler)
	mux.HandleFunc("/version", localVersionHandler)
	mux.HandleFunc("/kill", killHandler)
	mux.HandleFunc("/settings", settingsHandler)
	mux.Handle("/peerSocket", websocket.Handler(peerSocketHandler))
	mux.Handle("/static/", http.FileServer(http.Dir("./")))
	srv := &http.Server{Handler: mux, Addr: "localhost:" + sandwichSettings.LocalServerPort}
	srv.ListenAndServe()
}
