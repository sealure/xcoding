package executor

import (
	"context"
	"fmt"
	"strings"
	"time"
	"xcoding/apps/ci/executor_service/internal/parser"
	"xcoding/apps/ci/executor_service/models"
	civ1 "xcoding/gen/go/ci/v1"

	"gorm.io/gorm"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Scheduler struct {
	Env *K8sEnv
	DB  *gorm.DB
}

// Scheduler 负责单个 Job 的生命周期管理：
// - 创建 K8s Job 并等待容器就绪
// - 持续读取并解析日志，驱动 Step 状态变化
// - 在 Job 结束时（成功/失败/异常）兜底收敛 Step 终态
// 注意：不直接修改 Build 终态，构建结果由引擎在所有 Job 完成后统一计算

// NewScheduler 创建调度器，负责单个 Job 的创建、日志采集与状态落库
func NewScheduler(env *K8sEnv, db *gorm.DB) *Scheduler {
	return &Scheduler{Env: env, DB: db}
}

// finalizeSteps 在 Job 结束时兜底收敛步骤状态
// 语义：
// - failed=true：将该 Job 下所有处于 running 的步骤置为 failed，pending 置为 skipped
// - failed=false：将 running 置为 succeeded；理论上不应有 pending，若存在可按 skipped 处理或维持一致性
// 目的：避免日志标记缺失导致步骤卡在中间态，从而保证 Job/Step 状态与实际结论一致
func finalizeSteps(db *gorm.DB, buildID uint64, jobName string, failed bool) {
	now := time.Now()
	if failed {
		_ = db.Model(&models.BuildStep{}).Where("build_id = ? AND job_name = ? AND status = ?", buildID, jobName, "running").Updates(map[string]any{"status": "failed", "finished_at": &now}).Error
		_ = db.Model(&models.BuildStep{}).Where("build_id = ? AND job_name = ? AND status = ?", buildID, jobName, "pending").Updates(map[string]any{"status": "skipped", "finished_at": &now}).Error
		return
	}
	_ = db.Model(&models.BuildStep{}).Where("build_id = ? AND job_name = ? AND status = ?", buildID, jobName, "running").Updates(map[string]any{"status": "succeeded", "finished_at": &now}).Error
	_ = db.Model(&models.BuildStep{}).Where("build_id = ? AND job_name = ? AND status = ?", buildID, jobName, "pending").Updates(map[string]any{"status": "skipped", "finished_at": &now}).Error
}

// RunSingleJob 运行指定 job（不处理 needs），并把日志写入 Append
// 返回：
// - nil：job 成功完成
// - error：job 失败或日志流出错（用于通知上层引擎标记失败）
func (s *Scheduler) RunSingleJob(ctx context.Context, buildID uint64, jobName string, job parser.Job) error {
	ns := s.Env.Namespace
	name := fmt.Sprintf("build-%d-%s", buildID, jobName)
	nowStart := time.Now()
	// 标记该 Job 为 running 并记录开始时间（用于前端实时展示）
	_ = s.DB.Model(&models.BuildJob{}).Where("build_id = ? AND name = ?", buildID, jobName).Updates(map[string]any{"status": "running", "started_at": &nowStart}).Error
	spec := BuildJobSpecWithExtensions(ns, buildID, name, job)
	// 创建 K8s Job，失败则直接返回错误并由引擎判定该 Job 失败
	if _, err := s.Env.Clientset.BatchV1().Jobs(ns).Create(ctx, spec, metav1.CreateOptions{}); err != nil {
		return fmt.Errorf("create job: %w", err)
	}
	var podName string
	// 轮询查找第一个 Pod 名（依据 job-name 标签），避免立即读取日志失败
	for i := 0; i < 30; i++ {
		time.Sleep(500 * time.Millisecond)
		p, perr := s.Env.FirstPodNameByJob(ctx, name)
		if perr == nil && p != "" {
			podName = p
			break
		}
	}
	if podName == "" {
		return fmt.Errorf("pod not found for job %s", name)
	}
	// 等待容器就绪；若出现不可调度（Unschedulable）等错误或其它未就绪情况，判定 Job 失败
	// 注意：此处不写 Build 终态，让引擎在所有 Job 完成后统一计算构建结果
	if err := s.Env.WaitForContainerReady(ctx, podName, ns, "runner", 15*time.Second); err != nil {
		// 快速检查 Pod 条件，捕获不可调度场景
		pod, _ := s.Env.Clientset.CoreV1().Pods(ns).Get(ctx, podName, metav1.GetOptions{})
		if pod != nil {
			if isUnschedulable(pod) {
				now := time.Now()
				// Job 标记为失败并兜底步骤终态
				_ = s.DB.Model(&models.BuildJob{}).Where("build_id = ? AND name = ?", buildID, jobName).Updates(map[string]any{"status": "failed", "finished_at": &now}).Error
				finalizeSteps(s.DB, buildID, jobName, true)
				return fmt.Errorf("job unschedulable: %s", jobName)
			}
		}
		// 其它未就绪情况，继续按失败处理（仅更新 Job 状态；Build 终态由引擎统一落库）
		now := time.Now()
		_ = s.DB.Model(&models.BuildJob{}).Where("build_id = ? AND name = ?", buildID, jobName).Updates(map[string]any{"status": "failed", "finished_at": &now}).Error
		finalizeSteps(s.DB, buildID, jobName, true)
		return fmt.Errorf("container not ready: %s", jobName)
	}
	proc := NewLogProcessor(s.DB, buildID, jobName)
	// 持续读取 Pod 日志：
	// - 识别内部标记驱动 Step 状态（begin/end/exit）
	// - 非标记行按用户日志写入数据库
	if err := s.Env.StreamPodLogs(ctx, podName, ns, func(line string) {
		// 更新数据库状态
		statusEvent := proc.OnLine(ctx, line)

		// 只记录普通日志（非状态事件）
		if statusEvent == civ1.StepStatus_STEP_STATUS_UNSPECIFIED {
			// 过滤内部标记
			if !strings.HasPrefix(strings.TrimSpace(line), MarkerStepExit) {
				proc.SaveLog(ctx, line)
			}
		}
	}); err != nil {
		return fmt.Errorf("logs stream: %w", err)
	}
	// 流日志结束后轮询 K8s Job 状态，直至观察到 Succeeded/Failed 或达到上限
	var status civ1.BuildStatus = civ1.BuildStatus_BUILD_STATUS_RUNNING
	for i := 0; i < 40; i++ { // 最长约 20s
		j, _ := s.Env.Clientset.BatchV1().Jobs(ns).Get(ctx, name, metav1.GetOptions{})
		if j != nil {
			if j.Status.Succeeded > 0 {
				status = civ1.BuildStatus_BUILD_STATUS_SUCCEEDED
				break
			}
			if j.Status.Failed > 0 {
				status = civ1.BuildStatus_BUILD_STATUS_FAILED
				break
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
	now := time.Now()
	if status == civ1.BuildStatus_BUILD_STATUS_SUCCEEDED {
		// Job 成功：更新 Job 终态并兜底收敛步骤状态为成功/跳过
		_ = s.DB.Model(&models.BuildJob{}).Where("build_id = ? AND name = ?", buildID, jobName).Updates(map[string]any{"status": "succeeded", "finished_at": &now}).Error
		finalizeSteps(s.DB, buildID, jobName, false)
		return nil
	}
	if status == civ1.BuildStatus_BUILD_STATUS_FAILED {
		// Job 失败：更新 Job 终态并兜底收敛步骤状态为失败/跳过
		_ = s.DB.Model(&models.BuildJob{}).Where("build_id = ? AND name = ?", buildID, jobName).Updates(map[string]any{"status": "failed", "finished_at": &now}).Error
		finalizeSteps(s.DB, buildID, jobName, true)
		return fmt.Errorf("job failed: %s", jobName)
	}
	// 状态仍未刷新，保守返回错误以便引擎标记失败
	finalizeSteps(s.DB, buildID, jobName, true)
	return fmt.Errorf("job status unknown after logs: %s", jobName)
}

// isUnschedulable 判断 Pod 是否不可调度（根据 PodScheduled 条件）
func isUnschedulable(pod *corev1.Pod) bool {
	if pod == nil {
		return false
	}
	for _, c := range pod.Status.Conditions {
		if c.Type == corev1.PodScheduled && c.Status == corev1.ConditionFalse && c.Reason == "Unschedulable" {
			return true
		}
	}
	return false
}
