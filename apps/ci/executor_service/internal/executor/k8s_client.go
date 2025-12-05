package executor

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"os"
)

type K8sEnv struct {
	Clientset *kubernetes.Clientset
	Namespace string
}

// NewK8sEnv 创建 K8s 环境封装
// 说明：在集群内读取配置，初始化 Clientset，并从环境变量 POD_NAMESPACE 获取命名空间
func NewK8sEnv() (*K8sEnv, error) {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	cs, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	ns := os.Getenv("POD_NAMESPACE")
	if ns == "" {
		ns = "xcoding"
	}
	return &K8sEnv{Clientset: cs, Namespace: ns}, nil
}
