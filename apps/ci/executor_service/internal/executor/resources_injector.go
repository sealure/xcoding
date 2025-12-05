package executor

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"xcoding/apps/ci/executor_service/internal/parser"
)

// BuildResources 从约定的环境变量构造资源请求/限制
// 说明：
// - 支持 CPU/Memory 的 Request 与 Limit
// - 键：XC_RESOURCE_CPU_REQUEST、XC_RESOURCE_MEMORY_REQUEST、XC_RESOURCE_CPU_LIMIT、XC_RESOURCE_MEMORY_LIMIT
func BuildResources(job parser.Job) corev1.ResourceRequirements {
	req := corev1.ResourceList{}
	lim := corev1.ResourceList{}
	if v, ok := job.Env["XC_RESOURCE_CPU_REQUEST"]; ok && v != "" {
		req[corev1.ResourceCPU] = resource.MustParse(v)
	}
	if v, ok := job.Env["XC_RESOURCE_MEMORY_REQUEST"]; ok && v != "" {
		req[corev1.ResourceMemory] = resource.MustParse(v)
	}
	if v, ok := job.Env["XC_RESOURCE_CPU_LIMIT"]; ok && v != "" {
		lim[corev1.ResourceCPU] = resource.MustParse(v)
	}
	if v, ok := job.Env["XC_RESOURCE_MEMORY_LIMIT"]; ok && v != "" {
		lim[corev1.ResourceMemory] = resource.MustParse(v)
	}
	return corev1.ResourceRequirements{Requests: req, Limits: lim}
}
