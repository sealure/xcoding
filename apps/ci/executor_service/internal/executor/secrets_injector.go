package executor

import (
	corev1 "k8s.io/api/core/v1"
	"strings"
	"xcoding/apps/ci/executor_service/internal/parser"
)

// BuildEnvVars 将 env 映射转换为 K8s EnvVar，支持 secret:// 注入
// 说明：secret://<name>/<key> 转换为 SecretKeyRef；其它值直接作为明文
func BuildEnvVars(env map[string]string) []corev1.EnvVar {
	out := make([]corev1.EnvVar, 0, len(env))
	for k, v := range env {
		s := strings.TrimSpace(v)
		if strings.HasPrefix(s, "secret://") {
			p := strings.TrimPrefix(s, "secret://")
			parts := strings.SplitN(p, "/", 2)
			if len(parts) == 2 {
				out = append(out, corev1.EnvVar{
					Name:      k,
					ValueFrom: &corev1.EnvVarSource{SecretKeyRef: &corev1.SecretKeySelector{LocalObjectReference: corev1.LocalObjectReference{Name: parts[0]}, Key: parts[1]}},
				})
				continue
			}
		}
		out = append(out, corev1.EnvVar{Name: k, Value: v})
	}
	return out
}

// CollectSecretEnvVars 收集包含 secret:// 前缀的环境变量
func CollectSecretEnvVars(env map[string]string) map[string]string {
	out := map[string]string{}
	for k, v := range env {
		if strings.HasPrefix(strings.TrimSpace(v), "secret://") {
			out[k] = v
		}
	}
	return out
}

// BuildEnvVarsForJob 合并 Job 与 Step 的敏感环境变量并转换为 EnvVar
func BuildEnvVarsForJob(job parser.Job) []corev1.EnvVar {
	merged := map[string]string{}
	for k, v := range job.Env {
		merged[k] = v
	}
	for _, st := range job.Steps {
		for k, v := range st.Env {
			if strings.HasPrefix(strings.TrimSpace(v), "secret://") {
				merged[k] = v
			}
		}
	}
	return BuildEnvVars(merged)
}
