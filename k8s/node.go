package k8s

import (
	"context"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type NodeClient struct {
	clientSet *kubernetes.Clientset
}

func NewNodeClient(clientSet *kubernetes.Clientset) *NodeClient {
	return &NodeClient{
		clientSet: clientSet,
	}
}

// 获取node列表
func (n *NodeClient) List(labels string) ([]corev1.Node, error) {
	// options用于过滤,是否由label标签
	opts := metav1.ListOptions{}
	if labels != "" {
		opts.LabelSelector = labels
	}
	nodeList, err := n.clientSet.CoreV1().Nodes().List(context.Background(), opts)
	if err != nil {
		return nil, err
	}
	return nodeList.Items, nil
}
