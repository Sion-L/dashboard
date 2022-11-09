package contorller

import (
	"context"
	"devops/models"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/klog/v2"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"time"
)

func GetMetrics(c *gin.Context) {
	// post - json
	var metricsQuery models.MetricsQuery
	if err := c.ShouldBindJSON(&metricsQuery); err != nil {
		klog.V(2).ErrorS(err, "bind models.MetricsQuery to json failed")
		WriteOk(c, gin.H{})
		return
	}
	// 从数据库配置中去获取prometheus 的服务地址
	// 先判断prometheus服务是否可以 (配置超时时间
	ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Millisecond)
	defer cancel()
	readyRequest, err := http.NewRequest("GET", "http://prometheus.firecloud.wan/-/ready", nil)
	if err != nil {
		klog.V(2).ErrorS(err, "check prometheus service ready failed")
		WriteOk(c, gin.H{}) // 返回空
		return
	}

	readyResp, err := http.DefaultClient.Do(readyRequest.WithContext(ctx))
	if err != nil {
		klog.V(2).ErrorS(err, "check prometheus server response failed")
		WriteOk(c, gin.H{})
		return
	}

	// 如果还没成功，则返回前端空数据
	if readyResp.StatusCode != http.StatusOK {
		WriteOk(c, gin.H{})
		return
	}

	// 正常则执行真正的查询任务 - 断言判断
	wg := sync.WaitGroup{}
	tracker := models.NewPrometheusTracker()
	step := 60
	end := time.Now().Unix()
	start := end - 3600
	e := reflect.ValueOf(&metricsQuery).Elem()
	for i := 0; i < e.NumField(); i++ {
		wg.Add(i)
		go func(i int) {
			// 执行查询
			defer wg.Done()
			fName := e.Type().Field(i).Name
			fValue := e.Field(i).Interface().(*models.MetricsCategory)
			fTag := e.Type().Field(i).Tag // 决定返回的是结构体的tag名称
			if fValue == nil {
				return
			}
			klog.V(2).InfoS("start request prometheus data", "field", fName)
			prometheusQueries := fValue.GenerateQuery()
			if prometheusQueries == nil {
				klog.V(2).InfoS("no Promql", "field", fName)
				return
			}
			// 拿到真正的查询语句
			promql := prometheusQueries.GetValueByField(fName)
			resp, err := http.Get(fmt.Sprintf("http://prometheus.firecloud.wan/api/v1/query_range?query=%s&start=%d&end=%d&step=%d", promql, start, end, step))

			if err != nil {
				klog.V(2).ErrorS(err, "request promql data error", "field", fName, "promql", promql)
				return
			}
			body, err1 := ioutil.ReadAll(resp.Body)
			defer func(Body io.ReadCloser) {
				_ = Body.Close()
			}(resp.Body)
			if err1 != nil {
				klog.V(2).ErrorS(err, "read response body error", "field", fName, "promql", promql)
				return
			}

			// 解析返回数据
			var prometheusResp models.PrometheusQueryResp
			if err := json.Unmarshal(body, &prometheusResp); err != nil {
				klog.V(2).ErrorS(err, "Json unmarshal promql response is failed", "field", fName, "promql", promql)
				return
			}

			// 把数据组装到最后要返回的json中
			tag := fTag.Get("json")
			tracker.Set(tag[:strings.Index(tag, ",omitempty")], &prometheusResp) // 获取数据字段tag，并去掉,omitempty
		}(i)
	}
	// 等待所有查询完成
	wg.Wait()
	WriteOk(c, gin.H{
		"metrics": tracker.Metrics,
	})
}
