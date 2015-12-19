// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"container/list"
	"encoding/json"
	"fmt"
	"log"
	"sync"
)

type OpCode int

const (
	MessageOp OpCode = iota
	HistoryOp
	NoticeOp
	JoinOp
)

type Message struct {
	Id      int64  `json:"id"`
	Op      OpCode `json:"op"`
	From    string `json:"from"`
	Message string `json:"message"`

	// pop up a notification
	Notify bool `json:"notify"`
}

func (m *Message) Json() []byte {
	data, _ := json.Marshal(m)
	return data
}
func (m *Message) FromJson(data []byte) error {
	err := json.Unmarshal(data, m)
	return err
}

// hub maintains the set of active connections and broadcasts messages to the
// connections.
type hub struct {
	connections map[*connection]bool
	broadcast   chan []byte
	register    chan *connection
	unregister  chan *connection
	mx          sync.Mutex
	history     *list.List
}

func NewHub() *hub {
	h := &hub{
		broadcast:   make(chan []byte, 50),
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
	h.broadcast <- m.Json()
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

func (h *hub) run() {
	for {
		select {
		case c := <-h.register:
			log.Printf("register connection %#v\n", c)
			h.connections[c] = true

			h.send(NoticeOp, fmt.Sprintf("someone has joined"))

			// play back history
			for _, m := range h.getHistory() {
				select {
				case c.send <- m:
				default:
					close(c.send)
					delete(h.connections, c)
				}
			}

		case c := <-h.unregister:

			log.Printf("unregister connection %#v\n", c)
			if _, ok := h.connections[c]; ok {
				delete(h.connections, c)
				close(c.send)
				h.send(NoticeOp, fmt.Sprintf("someone has left"))
			}
		case data := <-h.broadcast:
			log.Printf("broadcast %s\n", string(data))
			m := &Message{}
			err := m.FromJson(data)
			if err != nil {
				log.Printf("ERROR: FromJson [ %s ]: %s", data, err)
				continue
			}

			if m.Op == MessageOp {
				h.history.PushBack(m)
				if h.history.Len() > 5 {
					if e := h.history.Front(); e != nil {
						h.history.Remove(e)
					}
				}
				h.sendBroadcast(m)
			} else if m.Op == NoticeOp {
				h.sendBroadcast(m)
			} else {
				log.Printf("Unhandled op %+v\n", m)
			}
		}
	}
}
