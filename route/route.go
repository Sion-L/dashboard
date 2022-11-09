package route

import (
	"devops/contorller"
	"github.com/gin-gonic/gin"
	"net/http"
)

func InitApi(eng *gin.Engine) {

	// 使用跨域中间件
	eng.Use(CoreMiddleware)

	// 注册一个check health的接口
	eng.GET("/ping", contorller.Ping)
	// 接口分组
	api := eng.Group("/api/v1")

	// 获取node列表接口
	api.GET("nodes", contorller.GetNodeList)

	// 获取metrics指标数据
	api.POST("metrics", contorller.GetMetrics)

	// 获取pod日志
	// api/v1/namespaces/kube-system/pods/fc-auth-sdasdaxasd  前端通过这样会获取 - 动态的
	api.GET("namespaces/:namespace/pods/:pod/logs", contorller.GetKubeLogs)

	// 远程执行pod的终端的接口
	// websocket://127.0.0.1:8888/api/v1/namespaces/namespace/pods/pod/shell
	api.GET("namespaces/:namespace/pods/:pod/shell", contorller.HandlerTerminal)
}

// 由于前后端端口不一样，可能导致跨域问题，此处解决跨域
func CoreMiddleware(c *gin.Context) {
	method := c.Request.Method
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "POST,GET,OPTIONS,PUT,DELETE,UPDATE")
	c.Header("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization")
	c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Cache-Control, Content-Language, Content-Type")
	c.Header("Access-Control-Allow-Credentials", "true")
	c.Set("content-type", "application/json")
	if method == "OPTIONS" {
		c.AbortWithStatus(http.StatusNoContent)
	}
	c.Next()
}
