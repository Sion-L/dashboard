package k8s

import (
	"github.com/gorilla/websocket"
	"net/http"
)

// 升级
var upGrader = websocket.Upgrader{
	ReadBufferSize:  2048,
	WriteBufferSize: 2048,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}
