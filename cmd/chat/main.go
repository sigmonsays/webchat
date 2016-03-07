package main

import (
	"container/list"
	"flag"
	"fmt"
	"log"
	"net/http"
	"sync"
	"text/template"

	"github.com/sigmonsays/webchat"
)

const (
	HistoryOp = iota + 100
)

var addr = flag.String("addr", ":8080", "http service address")

type chatHandler struct {
	mx      sync.Mutex
	history *list.List
}

func (h *chatHandler) serveHome(w http.ResponseWriter, r *http.Request) {
	log.Printf("request %s", r.URL)
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	homeTempl := template.Must(template.ParseFiles("home.html"))
	homeTempl.Execute(w, r.Host)
}

func (h *chatHandler) addHistory(m *webchat.Message) {
	h.mx.Lock()
	defer h.mx.Unlock()
	h.history.PushBack(m)
	if h.history.Len() > 5 {
		if e := h.history.Front(); e != nil {
			h.history.Remove(e)
		}
	}
}

func (h *chatHandler) getHistory() []*webchat.Message {
	h.mx.Lock()
	defer h.mx.Unlock()
	ret := make([]*webchat.Message, 0)
	for e := h.history.Front(); e != nil; e = e.Next() {
		ret = append(ret, e.Value.(*webchat.Message))
	}
	return ret
}

func (h *chatHandler) handleMessage(op webchat.OpCode, hub *webchat.Hub, c *webchat.Connection, m *webchat.Message) error {
	log.Printf("handleMessage op:%s\n", op)

	if op == webchat.MessageOp {
		hub.SendBroadcast(m)

	} else if op == webchat.RegisterOp {
		// play back history
		for _, hm := range h.getHistory() {
			hm.Op = HistoryOp
			hub.SendMessage(c, hm)
		}
	} else if op == webchat.UnregisterOp {
		hub.Send(webchat.NoticeOp, fmt.Sprintf("%s has left", c.Name))
		//} else if m.Op == HistoryOp {
		//		hub.sendBroadcast(m)

	} else if op == webchat.NickOp {

		if c.Name == "" {
			hub.Send(webchat.NoticeOp, fmt.Sprintf("%s has joined", m.From))
		} else {
			hub.Send(webchat.NoticeOp, fmt.Sprintf("%s has changed their name to %s", c.Name, m.From))
		}
		c.Name = m.From

	} else if op == webchat.NoticeOp {
		hub.SendBroadcast(m)
	} else {
		log.Printf("Unhandled op %+v\n", m)
	}
	return nil
}

func main() {
	flag.Parse()

	hub := webchat.NewHub()

	srv, err := webchat.NewHandler(hub)
	if err != nil {
		log.Fatal("NewHandler: ", err)
	}

	handler := &chatHandler{
		history: list.New(),
	}

	opcodes := []webchat.OpCode{
		webchat.RegisterOp,
		webchat.UnregisterOp,
		webchat.NoticeOp,
		webchat.NickOp,
		webchat.MessageOp,
	}
	for _, op := range opcodes {
		hub.OnCallback(op, handler.handleMessage)
	}
	go hub.Start()

   mx := http.NewServeMux()

	mx.HandleFunc("/", handler.serveHome)
	mx.HandleFunc("/ws", srv.ServeWebSocket)

   alias := "/chat"
	mx.HandleFunc(alias, handler.serveHome)
	mx.HandleFunc(alias + "/ws", srv.ServeWebSocket)

   hs := &http.Server{
      Addr: *addr,
      Handler: mx,
   }

	err = hs.ListenAndServe()
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
