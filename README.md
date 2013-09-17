##sandwich-go

A client and server for Sandwich, written in Go.

(previously known as "how to insert penis into sandwich", thanks Jacob).


###About Sandwich

Sandwich is a distributed file-sharing application which operates entirely over HTTP/HTTPS. The current canonical implementation is in Go, but other implementations are welcomed and encouraged, as long as they are interoperable with other clients through the spec below.

# Get Go and Sandwich

```
sudo apt-get install golang-go mercurial # This is specific to Ubuntu
echo "export GOPATH=$HOME/go" >> ~/.profile
mkdir -p $HOME/go/src
cd $HOME/go/src
git clone git://github.com/sandwich-share/sandwich-go
```


# To Build

```
./build.sh update
./build.sh
```

# Sandwich Spec

We've completely scrapped the old design for Sandwich. Here's the new gig:

This specification is provided so that alternate api-compatible sandwich clients can be written.

Sandwich is divided into a handful of submodules:

1. Local Client
2. Local Server
3. IRC Client

Each section provides its own URIs to which it will respond.

**DISCLAIMER: sandwich HEAD is not guaranteed to fully implement all of the below features.**

#Local Server

The local server is the most integral part of Sandwich. The local server is responsible for interacting with the greater Sandwich network. It maintains an updated peer list and stores the file indexes for all remote peers on the network. It is also responsible for the bootstrapping operation.

`GET /`

***Response:*** Serves up the static HTML index page.

---

`GET /search`

***Params:*** {'search': [searchterm],
               'regex': [true|false],
               'start': [startlocation],
               'step': [stepsize]}

`search` is either a plain text string or a regular expression to
search for. This will search across all peers. `regex` should be the string
'true' if you want it to be a regex search. `start` is the number to start
returning from, and step is the number of items to return. To do pagination,
you can return the first 100 with `start=0` and `step=100`, and then return
the next 100 with `start=100`, `step=100`.

***Response:*** JSON list: [{'IP': [ip], 'Port': [port], 'FileName': [filename]}]

This will return a list of all files matching the query, giving you the IP,
the Port number, and the full file path of the file.

---

`GET /peer`

***Params:*** {'peer': [ip], 'path': [path], 'start': [start], 'step': [step}

Specify the peer to query, as well as the directory to display. See the
`search` handler for an explanation of start/stop.

***Response:*** JSON list: [{'Type': [FileType], 'Name': [FileName]}

This will return a list of the files. For each file, it will specify the
`FileType`, which is `0` for directories, and `1` for files. The name is the
full path to the file.

---

`GET /download`

***Params:*** {'ip': [ip], 'type': [FileType], 'file': [FileName]}

Specify the IP to download from, the full path to download, and the FileType
(0 for Dir, 1 for File). Downloading a directory will recursively download the
directory. All files downloaded using this will download the files with the
mirrored path of the source. This has no response.

---

`GET /version`

***Response:*** [version]

This just returns the current version number of Sandwich.

---

`GET /kill`

Has no response. Just shuts down Sandwich.


##URIs

`GET /ping`

***Response:*** `pong`

Used to check whether a server is alive. Generally requested by the sandwich client prior to displaying search results.

---

`GET /peerlist`

***Response:*** JSON list: `[{"ip": [ip], "indexhash": [hash], "lastseen": [time]} ...{} ...{}]`

Note that `ip` and `hash` should be transmitted as strings. `time` should be transmitted in seconds since epoch (not as a string)

***Usage:***
If the local server is bootstrapping into the network, it will call `/peerlist` on whatever nodes it has in its list of bootstrap nodes. This will cause it to receive a peerlist containing every other node on the network. The nodes from which it requested `/peerlist` will now contain the new local server in their peerlist, where it will eventually disperse across the network as the network syncs.

The network sync operation is based around eventual consistency. Every 2 seconds, Each node in the network will query `/peerlist` for another node in its current list of peers. It will then merge that peer list with it's own, in the process discovering new node join/drops on the network. Any nodes that the local server has which were not in the requested node's peer list are pinged with `/ping` to make sure they are still alive.

Through this model, the network will eventually become consistent.

---

`GET /fileindex`

**Response**: A file which is the index for the host.

After receiving another node's peer list, the hashes for each peer's index in the new peer list are compared with the previous values the local server has. If these values differ and the `lastseen` value for the IP is more recent on the received peer list, then the local server will request `/indexfor` on the peer from which the new peer list was received, passing in the IP of the peer whose hashes differ as the parameter `ip`.

---

`GET /files/<path>`

**Response** The file at `path` with the root at the root of the share folder.


#Local Client

The local client is responsible for building the local file index, executing searches, and displaying a UI. The default UI is via the browser, but other implementations are possible.

This section is still under construction, come back later.


#IRC Client

Ideally we want to have an imbedded IRC client in the UI. This hasn't really been fleshed out much, so come back later :)
