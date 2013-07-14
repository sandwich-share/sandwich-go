package main

import (
	"code.google.com/p/go.net/websocket"
)

type connection struct {
	ws   *websocket.Conn
	send chan string
}

func (c *connection) reader() {
	for {
		var message string
		err := websocket.Message.Receive(c.ws, &message)
		if err != nil {
			break
		}
	}
	c.ws.Close()
}

func (c *connection) writer() {
	for json_message := range c.send {
		err := websocket.Message.Send(c.ws, json_message)
		if err != nil {
			break
		}
	}
	c.ws.Close()
}

type hub struct {
	connections map[*connection]bool
	broadcast   chan string
	register    chan *connection
	unregister  chan *connection
}

var peerHub = hub{
	connections: make(map[*connection]bool),
	broadcast:   make(chan string),
	register:    make(chan *connection),
	unregister:  make(chan *connection),
}

func (h *hub) run() {
	for {
		select {
		case c := <-h.register:
			h.connections[c] = true
		case c := <-h.unregister:
			delete(h.connections, c)
			close(c.send)
		case m := <-h.broadcast:
			for c := range h.connections {
				select {
				case c.send <- m:
				default:
					delete(h.connections, c)
					close(c.send)
					go c.ws.Close()
				}
			}
		}
	}
}

func peerSocketHandler(ws *websocket.Conn) {
	c := &connection{send: make(chan string, 256), ws: ws}
	peerHub.register <- c
	go writePeers()
	defer func() { peerHub.unregister <- c }()
	go c.writer()
	c.reader()
}
