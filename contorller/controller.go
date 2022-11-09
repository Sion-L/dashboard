package contorller

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func Ping(c *gin.Context) {
	c.JSON(http.StatusOK, "pong")
}

// 包装错误日志 - json
func WriteError(c *gin.Context, msg string) {
	c.JSON(http.StatusOK, gin.H{
		"code":    1,
		"message": msg,
	})
}

// 包装成功日志 - json
func WriteOk(c *gin.Context, data interface{}) {
	// 断言，判断传进来的是否就是gin.H
	ret, ok := data.(gin.H)
	if !ok {
		ret = gin.H{}
		ret["data"] = data
	}
	ret["code"] = 0
	ret["message"] = "success"
	c.JSON(http.StatusOK, ret)
	c.String(200, "%v", data)
}
