package server

import (
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var buf chan string

func init() {
	buf = make(chan string, 1500)
}

func logHandlerPost(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, HEAD")
	w.Header().Set("Access-Control-Allow-Headers", "X-PINGOTHER, Content-Type")
	w.Header().Set("Access-Control-Max-Age", "86400")
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Errorf("Error upgrading to websocket conn %s", err)
		return
	}
	defer c.Close()
	for {
		_, msg, err := c.ReadMessage()
		if err != nil {
			log.Errorf("Error reading more messages %s", err.Error())
			break
		}
		buf <- string(msg)
	}
}

func logHandlerGet(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, HEAD")
	w.Header().Set("Access-Control-Allow-Headers", "X-PINGOTHER, Content-Type")
	w.Header().Set("Access-Control-Max-Age", "86400")
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Errorf("Error upgrading to websocket conn %s", err)
		return
	}
	defer c.Close()
	for msg := range buf {
		err := c.WriteMessage(websocket.TextMessage, []byte(msg))
		if err != nil {
			log.Errorf("Error writing to client %s", err.Error())
			break
		}
	}
}
