package main

import (
	"flag"
	"log"
	"net/http"
	"text/template"

	"github.com/sigmonsays/webchat"
)

var addr = flag.String("addr", ":8080", "http service address")

type chatHandler struct {
}

func (h *chatHandler) serveHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "Not found", 404)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	homeTempl := template.Must(template.ParseFiles("home.html"))
	homeTempl.Execute(w, r.Host)
}

func main() {
	flag.Parse()

	h := webchat.NewHub()
	go h.Start()

	srv, err := webchat.NewHandler(h)
	if err != nil {
		log.Fatal("NewHandler: ", err)
	}

	index := &chatHandler{}

	http.HandleFunc("/", index.serveHome)
	http.HandleFunc("/ws", srv.ServeWebSocket)
	err = http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
