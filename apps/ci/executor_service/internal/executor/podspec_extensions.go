package executor

import (
	"strconv"
	"xcoding/apps/ci/executor_service/internal/parser"

	batchv1 "k8s.io/api/batch/v1"
)

// BuildJobSpecWithExtensions 注入扩展：Job 超时、NodeSelector 与 TTL 自动清理
func BuildJobSpecWithExtensions(ns string, buildID uint64, jobName string, job parser.Job) *batchv1.Job {
	spec := BuildJobSpec(ns, buildID, jobName, job)

	// TTL：仅在配置时设置，默认不清理，便于调试
	ttl := ParseTTLFromEnv(job.Env)
	ApplyTTLExtension(spec, ttl)

	// Timeout 秒：与 TTL 区分，超时会使 Job 失败
	// 读取超时：XC_JOB_TIMEOUT_SECONDS
	if v, ok := job.Env["XC_JOB_TIMEOUT_SECONDS"]; ok && v != "" {
		if i, err := strconv.ParseInt(v, 10, 64); err == nil && i > 0 {
			ii := i
			spec.Spec.ActiveDeadlineSeconds = &ii
		}
	}

	/*
		// 当前集群节点不支持，所以关闭
		// 场景：调度到特定硬件类型的节点
		NodeSelector: map[string]string{
		    "accelerator": "gpu-nvidia",
		    "storage": "ssd",
		}

	*/

	//// NodeSelector：约定 `XC_NODE_SELECTOR_KEY` and `XC_NODE_SELECTOR_VALUE`
	//k, kOk := job.Env["XC_NODE_SELECTOR_KEY"]
	//val, vOk := job.Env["XC_NODE_SELECTOR_VALUE"]
	//if kOk && vOk && k != "" && val != "" {
	//	if spec.Spec.Template.Spec.NodeSelector == nil {
	//		spec.Spec.Template.Spec.NodeSelector = map[string]string{}
	//	}
	//	spec.Spec.Template.Spec.NodeSelector[k] = val
	//}

	return spec
}
