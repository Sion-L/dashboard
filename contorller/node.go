package contorller

import (
	"devops/k8s"
	"github.com/gin-gonic/gin"
	"k8s.io/klog/v2"
)

func GetNodeList(c *gin.Context) {
	nodes, err := k8s.Client.Node.List("")
	if err != nil {
		// 后面再追加表示哪块的代码
		klog.V(2).ErrorS(err, "get nodeList failed", "controller", "GetNodeList")
		WriteError(c, err.Error())
		return
	}
	// 传递给前端，前端处理nodes数据
	WriteOk(c, gin.H{"nodes": nodes})
}
