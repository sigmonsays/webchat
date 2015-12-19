// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"container/list"
	"log"
	"sync"
)

type OpCode int

const (
	MessageOp OpCode = iota
	HistoryOp
)

type Message struct {
	Op      OpCode
	Message string
}

// hub maintains the set of active connections and broadcasts messages to the
// connections.
type hub struct {
	connections map[*connection]bool
	broadcast   chan *Message
	register    chan *connection
	unregister  chan *connection
	mx          sync.Mutex
	history     *list.List
}

func NewHub() *hub {
	h := &hub{
		broadcast:   make(chan *Message),
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

func (h *hub) run() {
	for {
		select {
		case c := <-h.register:
			log.Printf("register connection %#v\n", c)
			h.connections[c] = true

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
			}
		case m := <-h.broadcast:
			log.Printf("broadcast %+v\n", m)
			h.history.PushBack(m)
			if h.history.Len() > 5 {
				if e := h.history.Front(); e != nil {
					h.history.Remove(e)
				}
			}
			for c := range h.connections {
				select {
				case c.send <- m:
				default:
					close(c.send)
					delete(h.connections, c)
				}
			}
		}
	}
}
