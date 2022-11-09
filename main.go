package main

import (
	"devops/k8s"
	"devops/route"
	"flag"
	"github.com/gin-gonic/gin"
	"k8s.io/klog/v2"
)

// 全局k8s初始化入口
func initialize() error {
	var err error
	// TODO 初始化配置文件
	if err = k8s.NewKubeClient(); err != nil {
		return err
	}
	return nil
}

func main() {

	// 解析flag参数,初始化klog，绑定到本地的flagset
	klog.InitFlags(nil)
	defer klog.Flush()
	flag.Set("logtostderr", "false")
	flag.Set("alsoToStderr", "false")
	flag.Parse()

	// k8s全局初始化
	if err := initialize(); err != nil {
		klog.V(2).ErrorS(err, "init global failed")
		return
	}

	serv := gin.Default()

	// 注册路由
	route.InitApi(serv)

	// 启动服务
	if err := serv.Run(":8888"); err != nil {
		klog.V(2).ErrorS(err, "Server run failed")
	}

}
