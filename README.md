##sandwich-go

A client for Sandwich, written in Go.

(previously known as "how to insert penis into sandwich", thanks Jacob).


###About Sandwich

Sandwich is a distributed file-sharing application which operates entirely over HTTP/HTTPS. The current canonical implementation is in Go, but other implementations are welcomed and encouraged, as long as they are interoperable with other clients through the spec below.


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
