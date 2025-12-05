package executor

import (
	"fmt"
	"xcoding/apps/ci/executor_service/internal/parser"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// BuildJobSpec 构建单个 K8s Job 规范
// 说明：合成容器镜像、环境变量、脚本与资源限制，并设置标签用于检索
func BuildJobSpec(ns string, buildID uint64, jobName string, job parser.Job) *batchv1.Job {
	backoff := int32(0)
	image := job.Container
	if image == "" {
		image = "alpine:latest"
	}
	envs := BuildEnvVarsForJob(job)
	script := BuildScript(job)
	pod := BuildPodSpec(image, script, envs)
	if len(pod.Containers) > 0 {
		pod.Containers[0].Resources = BuildResources(job)
	}
	labels := map[string]string{"app": "ci-executor-build", "xcoding.io/build-id": fmt.Sprintf("%d", buildID)}
	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{Name: jobName, Namespace: ns, Labels: labels},
		Spec:       batchv1.JobSpec{BackoffLimit: &backoff, Template: corev1.PodTemplateSpec{ObjectMeta: metav1.ObjectMeta{Labels: labels}, Spec: pod}},
	}
}
