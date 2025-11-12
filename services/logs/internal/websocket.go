package internal

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type socketHandler struct {
	upgrade websocket.Upgrader
	logger  <-chan string
}

func NewWebsocketHandler(logger <-chan string) *socketHandler {
	return &socketHandler{
		upgrade: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		logger: logger,
	}
}

func (s *socketHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c, err := s.upgrade.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("error %s when upgrading connection to websocket", err)
		return
	}
	defer c.Close()

	done := make(chan struct{})

	go func() {
		for msg := range s.logger {
			err := c.WriteMessage(websocket.TextMessage, []byte(msg))
			if err != nil {
				log.Printf("Error %s when sending message to client", err)
				break
			}
		}
		close(done)
	}()

	select {

	case <-done:
	case <-r.Context().Done():
	}

}
