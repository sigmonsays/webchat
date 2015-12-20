package webchat

import (
	"fmt"
	"log"
	"sync"
)

// a low level message
type message struct {
	data       []byte
	connection *Connection
}

func (m *message) String() string {
	return fmt.Sprintf("<conn:%d data:%d>", m.connection.id, len(m.data))
}

type CallbackFn func(op OpCode, hub *Hub, c *Connection, m *Message) error

// hub maintains the set of active connections and broadcasts messages to the
// connections.
type Hub struct {
	connections map[*Connection]bool
	broadcast   chan *message
	register    chan *Connection
	unregister  chan *Connection
	mx          sync.Mutex

	callbacks map[OpCode][]CallbackFn
}

func NewHub() *Hub {
	callbacks := make(map[OpCode][]CallbackFn, 0)
	h := &Hub{
		broadcast:   make(chan *message, 50),
		register:    make(chan *Connection),
		unregister:  make(chan *Connection),
		connections: make(map[*Connection]bool),
		callbacks:   callbacks,
	}
	return h
}

func (h *Hub) OnCallback(callback OpCode, fn CallbackFn) {
	ls, ok := h.callbacks[callback]
	if ok == false {
		ls = make([]CallbackFn, 0)
	}
	ls = append(ls, fn)
	h.callbacks[callback] = ls
}

func (h *Hub) SendMessage(c *Connection, m *Message) error {
	select {
	case c.send <- m:
	default:
		close(c.send)
		delete(h.connections, c)
	}
	return nil
}

func (h *Hub) Send(op OpCode, msg string) {
	m := &Message{Op: op, Message: msg}
	h.SendBroadcast(m)
}

func (h *Hub) SendBroadcast(m *Message) {
	for c := range h.connections {
		m.Id = c.id
		select {
		case c.send <- m:
		default:
			close(c.send)
			delete(h.connections, c)
		}
	}
}

func (h *Hub) findConnection(id int64) (*Connection, error) {
	for c := range h.connections {
		if c.id == id {
			return c, nil
		}
	}
	return nil, fmt.Errorf("not found id:%d", id)
}

func (h *Hub) dispatch(op OpCode, c *Connection, m *Message) error {
	callbacks, ok := h.callbacks[op]
	if ok == false {
		return nil
	}
	log.Printf("dispatch %s: callbacks:%d", op, len(callbacks))

	var err error
	for _, callback := range callbacks {
		err = callback(op, h, c, m)
		if err != nil {
			log.Printf("callback %s: %s", op, err)
		}
	}

	return nil
}

func (h *Hub) Start() {
	for {
		select {
		case c := <-h.register:
			log.Printf("register connection id:%d remote:%s\n", c.id, c.ws.RemoteAddr())
			h.connections[c] = true

			h.dispatch(RegisterOp, c, nil)

		case c := <-h.unregister:
			log.Printf("unregister connection id:%d remote:%s\n", c.id, c.ws.RemoteAddr())
			if _, ok := h.connections[c]; ok {
				delete(h.connections, c)
				close(c.send)
				h.dispatch(UnregisterOp, c, nil)
			}

		case data := <-h.broadcast:
			m := &Message{
				connection: data.connection,
				Id:         data.connection.id,
			}

			err := m.FromJson(data.data)
			if err != nil {
				log.Printf("ERROR: FromJson [ %s ]: %s", data, err)
				continue
			}
			log.Printf("dispatch %s %s %s\n", m.Op, data, string(data.data))
			err = h.dispatch(m.Op, data.connection, m)
			if err != nil {
				log.Printf("dispatch %s: %s\n", m.Op, err)
			}
		}
	}
}
