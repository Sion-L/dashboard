package main

import (
	"devops/k8s"
	"io"
	"k8s.io/client-go/tools/remotecommand"
	"os"
)

var _ k8s.TtyHandler = &TestExec{}
var _ remotecommand.TerminalSizeQueue = &SizeQueue{}

type SizeQueue struct {
	resizeChan   chan remotecommand.TerminalSize
	stopResizing chan struct{}
}

func (s *SizeQueue) Next() *remotecommand.TerminalSize {
	select {
	case size := <-s.resizeChan:
		return &size
	case <-s.stopResizing:
		return nil
	}
}

type TestExec struct {
	tty bool
	SizeQueue
}

func (t *TestExec) Stdin() io.Reader {
	return os.Stdin
}

func (t *TestExec) Stdout() io.Writer {
	return os.Stdout
}

func (t *TestExec) Stderr() io.Writer {
	return os.Stderr
}

func (t *TestExec) Tty() bool {
	return t.tty
}

func (t *TestExec) Done() {
	close(t.stopResizing)
}

func main() {
	if err := k8s.NewKubeClient(); err != nil {
		panic(err)
	}

	cmd := []string{
		"/bin/sh",
	}
	handler := TestExec{
		tty: false,
		SizeQueue: SizeQueue{
			resizeChan:   make(chan remotecommand.TerminalSize),
			stopResizing: make(chan struct{}),
		},
	}
	if err := k8s.Client.Pod.Exec(cmd, &handler, "firecloud", "fc-asset-64689c79fb-9bchf", "fc-asset"); err != nil {
		panic(err)
	}

}
