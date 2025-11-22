package internal

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type WebsocketHub struct {
	upgrade websocket.Upgrader
	in      <-chan string

	mu    sync.Mutex
	conns map[*websocket.Conn]struct{}
}

func NewWebsocketHub(in <-chan string) *WebsocketHub {
	return &WebsocketHub{
		upgrade: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
		in:    in,
		conns: make(map[*websocket.Conn]struct{}),
	}
}

func (h *WebsocketHub) Run() {
	for msg := range h.in {
		h.mu.Lock()
		for c := range h.conns {
			_ = c.SetWriteDeadline(time.Now().Add(5 * time.Second))
			if err := c.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
				log.Printf("websocket write error: %v (dropping conn)", err)
				_ = c.Close()
				delete(h.conns, c)
			}
		}
		h.mu.Unlock()
	}
}

func (h *WebsocketHub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c, err := h.upgrade.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("error upgrading to websocket: %v", err)
		return
	}

	_ = c.WriteMessage(websocket.TextMessage, []byte("[Logs] websocket connected"))

	h.mu.Lock()
	h.conns[c] = struct{}{}
	h.mu.Unlock()

	<-r.Context().Done()

	h.mu.Lock()
	delete(h.conns, c)
	h.mu.Unlock()
	_ = c.Close()
}
