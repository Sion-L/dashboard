package models

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
)

type MetricsCategory struct {
	Category  string `json:"Category"`
	Nodes     string `json:"Nodes,omitempty"`
	PVC       string `json:"pvc,omitempty"`
	Pods      string `json:"pods,omitempty"`
	Selector  string `json:"selector,omitempty"`
	Namespace string `json:"namespace,omitempty"`
}

type MetricsQuery struct {
	MemoryUsage               *MetricsCategory `json:"memoryUsage,omitempty"`
	MemoryRequest             *MetricsCategory `json:"memoryRequest,omitempty"`
	MemoryLimits              *MetricsCategory `json:"memoryLimits,omitempty"`
	MemoryCapacity            *MetricsCategory `json:"memoryCapacity,omitempty"`
	MemoryAllocatableCapacity *MetricsCategory `json:"memoryAllocatableCapacity,omitempty"`
	CpuUsage                  *MetricsCategory `json:"cpuUsage,omitempty"`
	CpuLimits                 *MetricsCategory `json:"cpuLimits,omitempty"`
	CpuRequest                *MetricsCategory `json:"cpuRequest,omitempty"`
	CpuCapacity               *MetricsCategory `json:"cpuCapacity,omitempty"`
	CpuAllocatableCapacity    *MetricsCategory `json:"cpuAllocatableCapacity,omitempty"`
	FsSize                    *MetricsCategory `json:"fsSize,omitempty"`
	FsUsage                   *MetricsCategory `json:"fsUsage,omitempty"`
	PodUsage                  *MetricsCategory `json:"podUsage,omitempty"`
	PodCapacity               *MetricsCategory `json:"podCapacity,omitempty"`
	PodAllocatableCapacity    *MetricsCategory `json:"podAllocatableCapacity,omitempty"`
}

type PrometheusQuery struct {
	CpuUsage                  string
	CpuRequests               string
	CpuLimits                 string
	CpuCapacity               string
	CpuAllocatableCapacity    string
	MemoryUsage               string
	MemoryCapacity            string
	MemoryRequests            string
	MemoryLimits              string
	MemoryAllocatableCapacity string
	FsUsage                   string
	FsSize                    string
	NetworkReceive            string
	NetworkTransmit           string
	PodUsage                  string
	PodCapacity               string
	PodAllocatableCapacity    string
	DiskUsage                 string
	DiskCapacity              string
}

// ???????????????
type PrometheusQueryResp struct {
	Status string                   `json:"status"`
	Data   *PrometheusQueryRespData `json:"data"`
}

type PrometheusQueryRespData struct {
	ResultType string                      `json:"resultType"`
	Result     []PrometheusQueryRespResult `json:"result"`
}

type PrometheusQueryRespResult struct {
	Metric interface{}   `json:"metric"`
	Value  []interface{} `json:"value"`
}

// ??????????????????
type PrometheusTracker struct {
	// ?????????????????????????????????map
	sync.RWMutex
	Metrics map[string]*PrometheusQueryResp
}

func NewPrometheusTracker() *PrometheusTracker {
	return &PrometheusTracker{Metrics: map[string]*PrometheusQueryResp{}}
}

func (p *PrometheusTracker) Get(key string) (*PrometheusQueryResp, bool) {
	p.RLock()
	defer p.RUnlock()
	val, ext := p.Metrics[key] // ext?????????????????????
	return val, ext
}

func (p *PrometheusTracker) Set(key string, val *PrometheusQueryResp) {
	p.Lock()
	defer p.Unlock()
	p.Metrics[key] = val
}

// ????????????????????????????????????promql??????
func (p *PrometheusQuery) GetValueByField(field string) string {
	e := reflect.ValueOf(p).Elem()
	for i := 0; i < e.NumField(); i++ {
		if e.Type().Field(i).Name == field {
			return e.Field(i).Interface().(string)
		}
	}
	return ""
}

//sum(node_memory_MemTotal_bytes - (node_memory_MemFree_bytes + node_memory_Buffers_bytes + node_memory_Cached_bytes)) by (instance)
func (mc *MetricsCategory) GenerateQuery() *PrometheusQuery {
	switch mc.Category {
	case "cluster":
		// %s ??????????????????????????????????????????????????????
		return &PrometheusQuery{
			MemoryUsage:               strings.Replace("sum(node_memory_MemTotal_bytes - (node_memory_MemFree_bytes + node_memory_Buffers_bytes + node_memory_Cached_bytes)) by (kubernetes_name)", "_bytes", fmt.Sprintf("_bytes{kubernetes_node=~\"%s\"}", mc.Nodes), -1),
			MemoryRequests:            fmt.Sprintf(`sum(kube_pod_container_resource_requests{node=~"%s", resource="memory"}) by (component)`, mc.Nodes),
			MemoryLimits:              fmt.Sprintf(`sum(kube_pod_container_resource_limits{node=~"%s", resource="memory"}) by (component)`, mc.Nodes),
			MemoryCapacity:            fmt.Sprintf(`sum(kube_node_status_capacity{node=~"%s", resource="memory"}) by (component)`, mc.Nodes),
			MemoryAllocatableCapacity: fmt.Sprintf(`sum(kube_node_status_allocatable{node=~"%s", resource="memory"}) by (component)`, mc.Nodes),
			CpuUsage:                  fmt.Sprintf(`sum(rate(node_cpu_seconds_total{kubernetes_node=~"%s", mode=~"user|system"}[1m]))`, mc.Nodes),
			CpuRequests:               fmt.Sprintf(`sum(kube_pod_container_resource_requests{node=~"%s", resource="cpu"}) by (component)`, mc.Nodes),
			CpuLimits:                 fmt.Sprintf(`sum(kube_pod_container_resource_limits{node=~"%s", resource="cpu"}) by (component)`, mc.Nodes),
			CpuCapacity:               fmt.Sprintf(`sum(kube_node_status_capacity{node=~"%s", resource="cpu"}) by (component)`, mc.Nodes),
			CpuAllocatableCapacity:    fmt.Sprintf(`sum(kube_node_status_allocatable{node=~"%s", resource="cpu"}) by (component)`, mc.Nodes),
			PodUsage:                  fmt.Sprintf(`sum({__name__=~"kubelet_running_pod_count|kubelet_running_pods", instance=~"%s"})`, mc.Nodes),
			PodCapacity:               fmt.Sprintf(`sum(kube_node_status_capacity{node=~"%s", resource="pods"}) by (component)`, mc.Nodes),
			PodAllocatableCapacity:    fmt.Sprintf(`sum(kube_node_status_allocatable{node=~"%s", resource="pods"}) by (component)`, mc.Nodes),
			FsSize:                    fmt.Sprintf(`sum(node_filesystem_size_bytes{kubernetes_node=~"%s", mountpoint="/"}) by (kubernetes_node)`, mc.Nodes),
			FsUsage:                   fmt.Sprintf(`sum(node_filesystem_size_bytes{kubernetes_node=~"%s", mountpoint="/"} - node_filesystem_avail_bytes{kubernetes_node=~"%s", mountpoint="/"}) by (kubernetes_node)`, mc.Nodes, mc.Nodes),
		}
		// ??????????????????????????????????????????
	case "nodes":
		return &PrometheusQuery{
			MemoryUsage:    `sum(node_memory_MemTotal_bytes - (node_memory_MemFree_bytes + node_memory_Buffers_bytes + node_memory_Cached_bytes)) by (instance)`,
			MemoryCapacity: `sum(kube_node_status_capacity{resource="memory"}) by (node)`,
			CpuUsage:       `sum(rate(node_cpu_seconds_total{mode=~"user|system"}[1m])) by (instance)`,
			CpuCapacity:    `sum(kube_node_status_capacity{resource="cpu"}) by (node)`,
			FsSize:         `sum(node_filesystem_size_bytes{mountpoint="/"}) by (instance)`,
			FsUsage:        `sum(node_filesystem_size_bytes{mountpoint="/"} - node_filesystem_avail_bytes{mountpoint="/"}) by (instance)`,
		}
	}
	return nil
}
