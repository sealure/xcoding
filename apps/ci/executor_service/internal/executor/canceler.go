package executor

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CancelBuild 删除与构建关联的所有 Job（前缀匹配 build-<id>-* 以及 build-<id>）
func (e *K8sEnv) CancelBuild(ctx context.Context, buildID uint64) error {
	ns := e.Namespace
	jobs, err := e.Clientset.BatchV1().Jobs(ns).List(ctx, metav1.ListOptions{LabelSelector: fmt.Sprintf("xcoding.io/build-id=%d", buildID)})
	if err != nil {
		return err
	}
	for _, j := range jobs.Items {
		_ = e.Clientset.BatchV1().Jobs(ns).Delete(ctx, j.Name, metav1.DeleteOptions{})
	}
	return nil
}
