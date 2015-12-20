package webchat

import (
	"log"
	"net/http"
	"sync"
	"sync/atomic"
)

func NewHandler(hub *Hub) (*Handler, error) {
	h := &Handler{
		hub: hub,
	}
	return h, nil

}

type Handler struct {
	mx          sync.Mutex
	connections int64
	hub         *Hub
}

// serveWs handles websocket requests from the peer.
func (h *Handler) ServeWebSocket(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	id := atomic.AddInt64(&h.connections, 1)
	c := &Connection{
		id:   id,
		hub:  h.hub,
		send: make(chan *Message, 256),
		ws:   ws,
	}
	h.hub.register <- c
	go c.writePump()
	c.readPump()
}
