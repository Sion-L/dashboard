package k8s

import (
	"fmt"
	"github.com/gorilla/websocket"
	"io"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/client-go/tools/remotecommand"
	"net/http"
)

var (
	EOT            = "\u0004" // 字符编码
	_   TtyHandler = &kubeExec{}
)

type kubeExec struct {
	Conn     *websocket.Conn
	sizeChan chan remotecommand.TerminalSize
	stopChan chan struct{}
	tty      bool
}

// 前后端定义的消息格式 -类似协议
type execMsgType struct {
	Type string `json:"type"`
	Data string `json:"data,omitempty"`
	Rows uint16 `json:"rows,omitempty"` // 行
	Cols uint16 `json:"cols,omitempty"` // 列
}

func (k *kubeExec) Stdin() io.Reader {
	// kubeExec实现了ttyHandler中的函数，就是实现了reader和write
	return k
}

func (k *kubeExec) Stdout() io.Writer {
	return k
}

func (k *kubeExec) Stderr() io.Writer {
	return k
}

func (k *kubeExec) Tty() bool {
	return k.tty
}

func (k *kubeExec) Next() *remotecommand.TerminalSize {
	select {
	case size := <-k.sizeChan:
		return &size
	case <-k.stopChan:
		return nil
	}
}

func (k *kubeExec) Done() {
	close(k.stopChan)
}

// 升级为websocket
func NewkubeExec(w http.ResponseWriter, r *http.Request, responseHeader http.Header) (*kubeExec, error) {
	conn, err := upGrader.Upgrade(w, r, responseHeader)
	if err != nil {
		return nil, err
	}
	kubeExec := &kubeExec{
		Conn:     conn,
		sizeChan: make(chan remotecommand.TerminalSize),
		stopChan: make(chan struct{}),
		tty:      true,
	}
	return kubeExec, nil
}

func (k *kubeExec) Read(p []byte) (n int, err error) {
	// 数据读出来,放到p里面
	_, msg, err := k.Conn.ReadMessage()
	if err != nil {
		return copy(p, EOT), err
	}
	// 序列化
	var msgType execMsgType
	if err := json.Unmarshal([]byte(msg), &msgType); err != nil {
		return copy(p, EOT), nil
	}
	// 判断处理类型
	switch msgType.Type {
	case "read": // 读
		return copy(p, msgType.Data), nil
	case "resize": // resize类型去拿TerminalSize的数据放到sizechan
		k.sizeChan <- remotecommand.TerminalSize{Width: msgType.Cols, Height: msgType.Rows}
		return 0, nil
	default:
		return copy(p, EOT), fmt.Errorf("unknown message type: %s", err)
	}
}

func (k *kubeExec) Write(p []byte) (n int, err error) {
	msg, err := json.Marshal(execMsgType{
		Type: "write",
		Data: string(p),
	})
	if err != nil {
		return 0, err
	}
	// 得到后通过websocket发送数据
	if err := k.Conn.WriteMessage(websocket.TextMessage, msg); err != nil {
		return 0, err
	}

	return len(p), nil
}

func (k *kubeExec) Close() error {
	return k.Conn.Close()
}
