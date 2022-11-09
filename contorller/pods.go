package contorller

import (
	"devops/k8s"
	"github.com/gin-gonic/gin"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	"net/http"
	"strconv"
)

// 实时获取pod日志
func GetKubeLogs(c *gin.Context) {
	// 接口参数拼接 - api/v1/namespaces/namespace/pods/pod/logs?tailLines=50&timestamps=true&previous=false&container=xxx
	// tailLines 类似 kubectl logs -f --tail=xxxx , previous 是查前面一个容器的日志,默认false
	nameSpace := c.Param("namespace")
	podName := c.Param("pod")
	container := c.Query("container")
	tailLines, _ := strconv.ParseInt(c.DefaultQuery("tailLines", "1000"), 10, 64) // 前端没传递，给默认值,转成int64
	timestamps, _ := strconv.ParseBool(c.DefaultQuery("timestamps", "true"))
	previous, _ := strconv.ParseBool(c.DefaultQuery("previous", "false"))

	klog.V(2).InfoS("get kube logs request params", "namespace", nameSpace, "pod", podName, "container", container, "tailLines", tailLines, "timestemps", timestamps, "previous", previous)

	// 判断参数是否传递进来
	if nameSpace == "" || podName == "" || container == "" {
		c.String(http.StatusBadRequest, "must specific namespace,pod and container query params")
		return
	}

	// 获取pod日志 - 升级成websocket连接
	kubeLogger, err := k8s.NewKubeLogger(c.Writer, c.Request, nil)
	if err != nil {
		klog.V(2).ErrorS(err, "upgrade websocket failed")
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	// 构造获取日志的结构体
	opts := corev1.PodLogOptions{
		Container:  container,
		Follow:     true,
		TailLines:  &tailLines,
		Timestamps: timestamps,
		Previous:   previous,
	}

	if err := k8s.Client.Pod.LogsStream(podName, nameSpace, &opts, kubeLogger); err != nil {
		klog.V(2).ErrorS(err, "write logs Stream failed")
		_, _ = kubeLogger.Write([]byte(err.Error())) // 错误日志给到前端
	}

}

// pod终端
func HandlerTerminal(c *gin.Context) {
	namespace := c.Param("namespace")
	podName := c.Param("pod")
	container := c.Query("container")
	cmd := []string{
		"/bin/sh", "-c", "clear;(bash || sh)",
	}

	klog.V(2).InfoS("get kube logs request params",
		"namespace", namespace, "pod", podName, "container", container, "cmd", cmd)

	// 获取pod
	if _, err := k8s.Client.Pod.Get(podName, namespace); err != nil {
		klog.ErrorS(err, "get pods failed")
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	// 校验pod - 是否运行
	kubeExec, err := k8s.NewkubeExec(c.Writer, c.Request, nil)
	if err != nil {
		klog.ErrorS(err, "init kube exec failed")
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	defer func() {
		_ = kubeExec.Close()
	}()

	// 调用pod的exec
	if err := k8s.Client.Pod.Exec(cmd, kubeExec, namespace, podName, container); err != nil {
		klog.ErrorS(err, "exec pods failed")
		c.String(http.StatusBadRequest, err.Error())
		return
	}
}
