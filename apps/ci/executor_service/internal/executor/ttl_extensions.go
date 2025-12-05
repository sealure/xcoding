package executor

import (
	"strconv"

	batchv1 "k8s.io/api/batch/v1"
)

// ApplyTTLExtension 为 Job 注入 TTLSecondsAfterFinished（仅当提供有效 TTL 时）
// 说明：
// - 当 ttlSeconds > 0 时，设置 Job 在完成后自动清理的 TTL 秒数
// - 当 ttlSeconds <= 0 时，不设置 TTL，便于调试与保留 Job 记录
func ApplyTTLExtension(job *batchv1.Job, ttlSeconds int32) {
	if job == nil {
		return
	}
	if ttlSeconds > 0 {
		job.Spec.TTLSecondsAfterFinished = &ttlSeconds
	}
}

// ParseTTLFromEnv 从 Job 的 env 映射中解析用户提供的 TTL（键：XC_JOB_TTL_SECONDS）
// 返回值：
// - 成功解析且 >0：返回对应 int32 值
// - 未提供或无效：返回 0（不设置 TTL）
func ParseTTLFromEnv(env map[string]string) int32 {
	if env == nil {
		return 0
	}
	if v, ok := env["XC_JOB_TTL_SECONDS"]; ok && v != "" {
		if i, err := strconv.ParseInt(v, 10, 32); err == nil && i > 0 {
			return int32(i)
		}
	}
	// 一小时
	return 60 * 60
}
