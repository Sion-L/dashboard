package k8s

import (
	"github.com/gorilla/websocket"
	"net/http"
)

type KubeLogger struct {
	Conn *websocket.Conn
}

func NewKubeLogger(w http.ResponseWriter, r *http.Request, responseHeader http.Header) (*KubeLogger, error) {
	// 升级http get连接为websocket连接
	conn, err := upGrader.Upgrade(w, r, responseHeader)
	if err != nil {
		return nil, err
	}
	KubeLogger := &KubeLogger{
		Conn: conn,
	}
	return KubeLogger, nil
}

// websocket写入数据
func (k *KubeLogger) Write(data []byte) (int, error) {
	if err := k.Conn.WriteMessage(websocket.TextMessage, data); err != nil {
		return 0, err
	}
	return len(data), nil
}

// 写完关闭
func (k *KubeLogger) Close() error {
	return k.Conn.Close()
}
