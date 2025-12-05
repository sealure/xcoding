package executor

import (
	"bufio"
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

// StreamPodLogs 读取 Pod 日志并按行回调
// 说明：Follow 模式持续读取直到日志结束；单行回调，供日志处理器解析标记
func (e *K8sEnv) StreamPodLogs(ctx context.Context, podName, namespace string, onLine func(string)) error {
	ns := namespace
	if ns == "" {
		ns = e.Namespace
	}
	req := e.Clientset.CoreV1().Pods(ns).GetLogs(podName, &corev1.PodLogOptions{Follow: true})
	stream, err := req.Stream(ctx)
	if err != nil {
		return err
	}
	defer stream.Close()
	scanner := bufio.NewScanner(stream)
	scanner.Buffer(make([]byte, 0, 1024), 1024*1024)
	for scanner.Scan() {
		onLine(scanner.Text())
	}
	return nil
}

// FirstPodNameByJob 获取某 Job 的第一个 Pod 名
// 说明：通过标签 job-name=<job> 选择器定位 Pod
func (e *K8sEnv) FirstPodNameByJob(ctx context.Context, jobName string) (string, error) {
	pods, err := e.Clientset.CoreV1().Pods(e.Namespace).List(ctx, metav1.ListOptions{LabelSelector: "job-name=" + jobName})
	if err != nil {
		return "", err
	}
	if len(pods.Items) == 0 {
		return "", nil
	}
	return pods.Items[0].Name, nil
}

// WaitForContainerReady 等待指定 Pod 的容器就绪或进入终止态
// 参数：container 留空表示任意容器；timeout 总等待上限
func (e *K8sEnv) WaitForContainerReady(ctx context.Context, podName, namespace, container string, timeout time.Duration) error {
	ns := namespace
	if ns == "" {
		ns = e.Namespace
	}
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		pod, err := e.Clientset.CoreV1().Pods(ns).Get(ctx, podName, metav1.GetOptions{})
		if err == nil {
			for _, cs := range pod.Status.ContainerStatuses {
				if container == "" || cs.Name == container {
					if cs.State.Running != nil || cs.State.Terminated != nil || cs.Ready {
						return nil
					}
				}
			}
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		time.Sleep(300 * time.Millisecond)
	}
	return fmt.Errorf("container not ready: pod=%s container=%s", podName, container)
}
