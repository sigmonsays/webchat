//go:generate stringer -type=OpCode

package webchat

import (
	"container/list"
	"fmt"
	"log"
	"sync"
)

// a low level message
type message struct {
	data       []byte
	connection *connection
}

func (m *message) String() string {
	return fmt.Sprintf("<conn:%d data:%d>", m.connection.id, len(m.data))
}

// hub maintains the set of active connections and broadcasts messages to the
// connections.
type hub struct {
	connections map[*connection]bool
	broadcast   chan *message
	register    chan *connection
	unregister  chan *connection
	mx          sync.Mutex
	history     *list.List
}

func NewHub() *hub {
	h := &hub{
		broadcast:   make(chan *message, 50),
		register:    make(chan *connection),
		unregister:  make(chan *connection),
		connections: make(map[*connection]bool),
		history:     list.New(),
	}
	return h
}

func (h *hub) getHistory() []*Message {
	h.mx.Lock()
	defer h.mx.Unlock()
	ret := make([]*Message, 0)
	for e := h.history.Front(); e != nil; e = e.Next() {
		ret = append(ret, e.Value.(*Message))
	}
	return ret
}

func (h *hub) send(op OpCode, msg string) {
	m := &Message{Op: op, Message: msg}
	h.sendBroadcast(m)
}

func (h *hub) sendBroadcast(m *Message) {
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
func (h *hub) findConnection(id int64) (*connection, error) {
	for c := range h.connections {
		if c.id == id {
			return c, nil
		}
	}
	return nil, fmt.Errorf("not found id:%d", id)
}

func (h *hub) Start() {
	for {
		select {
		case c := <-h.register:
			log.Printf("register connection id:%d remote:%s\n", c.id, c.ws.RemoteAddr())
			h.connections[c] = true

			// play back history
			for _, m := range h.getHistory() {
				m.Op = HistoryOp
				select {
				case c.send <- m:
				default:
					close(c.send)
					delete(h.connections, c)
				}
			}

		case c := <-h.unregister:
			log.Printf("unregister connection id:%d remote:%s\n", c.id, c.ws.RemoteAddr())
			if _, ok := h.connections[c]; ok {
				delete(h.connections, c)
				close(c.send)
				h.send(NoticeOp, fmt.Sprintf("%s has left", c.Name))
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
			log.Printf("broadcast %s %s %s\n", m.Op, data, string(data.data))

			if m.Op == MessageOp {
				h.history.PushBack(m)
				if h.history.Len() > 5 {
					if e := h.history.Front(); e != nil {
						h.history.Remove(e)
					}
				}
				h.sendBroadcast(m)
			} else if m.Op == HistoryOp {
				h.sendBroadcast(m)
			} else if m.Op == JoinOp {

			} else if m.Op == NickOp {
				conn, err := h.findConnection(m.Id)
				if err != nil {
					log.Printf("findConnection %d: %s", m.Id, err)
					continue
				}

				if conn.Name == "" {
					h.send(NoticeOp, fmt.Sprintf("%s has joined", m.From))
				} else {
					h.send(NoticeOp, fmt.Sprintf("%s has changed their name to %s", conn.Name, m.From))
				}
				conn.Name = m.From

			} else if m.Op == NoticeOp {
				h.sendBroadcast(m)
			} else {
				log.Printf("Unhandled op %+v\n", m)
			}
		}
	}
}
