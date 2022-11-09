package k8s

import (
	"bufio"
	"context"
	"io"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

// TODO podList - 前端根据podList获取的pod查看日志，进行终端

// 终端接口 提供对应终端方法
type TtyHandler interface {
	Stdin() io.Reader
	Stdout() io.Writer
	Stderr() io.Writer
	Tty() bool
	remotecommand.TerminalSizeQueue
	Done()
}

type PodClient struct {
	clientSet *kubernetes.Clientset
	config    *rest.Config // 看实际需求
}

func NewPodClient(clientSet *kubernetes.Clientset, config *rest.Config) *PodClient {
	return &PodClient{
		clientSet: clientSet,
		config:    config,
	}
}

// 获取pod
func (p *PodClient) Get(name, namespace string) (*corev1.Pod, error) {
	opts := metav1.GetOptions{}
	return p.clientSet.CoreV1().Pods(namespace).Get(context.Background(), name, opts)
}

// 获取pod日志
func (p *PodClient) Logs(name, namespace string, opts *corev1.PodLogOptions) *rest.Request {
	return p.clientSet.CoreV1().Pods(namespace).GetLogs(name, opts)
}

// 流获取
func (p *PodClient) LogsStream(name, namespace string, opts *corev1.PodLogOptions, write io.Writer) error {
	req := p.Logs(name, namespace, opts)
	stream, err := req.Stream(context.TODO())
	if err != nil {
		return err
	}
	defer stream.Close()

	buf := bufio.NewReaderSize(stream, 2048)
	for {
		bytes, err := buf.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				_, err = write.Write(bytes)
			}
			return err
		}
		_, err = write.Write(bytes)
		if err != nil {
			return err
		}
	}
}

//

// 命令终端
func (p *PodClient) Exec(cmd []string, handler TtyHandler, nameSpace, pod, container string) error {

	defer func() {
		handler.Done()
	}()

	// client-go提供远程包
	req := p.clientSet.CoreV1().RESTClient().Post().
		Resource("pods").
		Namespace(nameSpace).
		Name(pod).SubResource("exec")

	opts := &corev1.PodExecOptions{
		Container: container, // 交互的容器名
		Command:   cmd,
		Stdin:     handler.Stdin() != nil,
		Stdout:    handler.Stdout() != nil,
		Stderr:    handler.Stderr() != nil,
		TTY:       handler.Tty(),
	}
	req.VersionedParams(opts, scheme.ParameterCodec)

	// 建立长连接
	executor, err := remotecommand.NewSPDYExecutor(p.config, "POST", req.URL())
	if err != nil {
		return err
	}

	return executor.Stream(remotecommand.StreamOptions{
		Stdin:             handler.Stdin(),
		Stdout:            handler.Stdout(), // websocket输出
		Stderr:            handler.Stderr(),
		Tty:               handler.Tty(),
		TerminalSizeQueue: handler,
	})

}
