package executor

import (
	"context"
	"strconv"
	"strings"
	"time"
	"xcoding/apps/ci/executor_service/models"
	civ1 "xcoding/gen/go/ci/v1"

	"gorm.io/gorm"
)

type LogProcessor struct {
	db            *gorm.DB
	buildID       uint64
	jobName       string
	currentStepID uint64
}

// NewLogProcessor 创建日志处理器：按标记更新步骤状态与退出码
func NewLogProcessor(db *gorm.DB, buildID uint64, jobName string) *LogProcessor {
	return &LogProcessor{db: db, buildID: buildID, jobName: jobName}
}

// OnLine 处理日志行：识别 __step_begin__/__step_end__/__step_exit__ 并更新数据库
// 返回值：status event (UNSPECIFIED if normal log)
func (p *LogProcessor) OnLine(ctx context.Context, line string) civ1.StepStatus {
	s := strings.TrimSpace(line)
	if strings.HasPrefix(s, MarkerStepBegin+" ") {
		name := strings.TrimSpace(strings.TrimPrefix(s, MarkerStepBegin+" "))
		now := time.Now()
		var step models.BuildStep
		if err := p.db.Model(&models.BuildStep{}).
			Where("build_id = ? AND job_name = ? AND name = ?", p.buildID, p.jobName, name).
			First(&step).Error; err == nil {
			p.currentStepID = step.ID
			_ = p.db.Model(&step).Updates(map[string]any{"status": "running", "started_at": &now}).Error
		}
		return civ1.StepStatus_STEP_STATUS_RUNNING
	}
    if strings.HasPrefix(s, MarkerStepEnd+" ") {
        name := strings.TrimSpace(strings.TrimPrefix(s, MarkerStepEnd+" "))
        now := time.Now()
        var step models.BuildStep
        if err := p.db.Model(&models.BuildStep{}).
            Where("build_id = ? AND job_name = ? AND name = ?", p.buildID, p.jobName, name).
            First(&step).Error; err == nil {
            _ = p.db.Model(&step).Updates(map[string]any{"status": "succeeded", "finished_at": &now}).Error
            if step.ID == p.currentStepID {
                p.currentStepID = 0
            }
        }
        return civ1.StepStatus_STEP_STATUS_SUCCEEDED
    }
	if strings.HasPrefix(s, MarkerStepExit+" ") {
		parts := strings.Split(s, " ")
		if len(parts) >= 3 {
			name := strings.TrimSpace(parts[1])
			code := strings.TrimSpace(parts[2])
			var exit int32 = 0
			if n, err := strconv.ParseInt(code, 10, 32); err == nil {
				exit = int32(n)
			}
			_ = p.db.Model(&models.BuildStep{}).
				Where("build_id = ? AND job_name = ? AND name = ?", p.buildID, p.jobName, name).
				Updates(map[string]any{"exit_code": exit}).Error
		}
		return civ1.StepStatus_STEP_STATUS_UNSPECIFIED
	}
	return civ1.StepStatus_STEP_STATUS_UNSPECIFIED
}

// SaveLog 将日志写入 BuildStepLogChunk
func (p *LogProcessor) SaveLog(ctx context.Context, content string) {
	if p.currentStepID == 0 {
		return // 忽略不在步骤内的日志
	}
	chunk := models.BuildStepLogChunk{
		BuildStepID: p.currentStepID,
		Content:     content,
		CreatedAt:   time.Now(),
	}
	_ = p.db.Create(&chunk).Error
}

// Finalize 在 Job 结束时兜底标记步骤状态
// failed=true: running->failed, pending->skipped
// failed=false: running->succeeded (兜底), pending->skipped (理论上不应有 pending，除非逻辑错误，但保持一致性可设为 skipped 或忽略)
func (p *LogProcessor) Finalize(ctx context.Context, failed bool) {
    now := time.Now()
    if failed {
        _ = p.db.Model(&models.BuildStep{}).
            Where("build_id = ? AND job_name = ? AND status = ?", p.buildID, p.jobName, "running").
            Updates(map[string]any{"status": "failed", "finished_at": &now}).Error
        _ = p.db.Model(&models.BuildStep{}).
            Where("build_id = ? AND job_name = ? AND status = ?", p.buildID, p.jobName, "pending").
            Updates(map[string]any{"status": "skipped", "finished_at": &now}).Error
        return
    }
    _ = p.db.Model(&models.BuildStep{}).
        Where("build_id = ? AND job_name = ? AND status = ?", p.buildID, p.jobName, "running").
        Updates(map[string]any{"status": "succeeded", "finished_at": &now}).Error
    _ = p.db.Model(&models.BuildStep{}).
        Where("build_id = ? AND job_name = ? AND status = ?", p.buildID, p.jobName, "pending").
        Updates(map[string]any{"status": "succeeded", "finished_at": &now}).Error
}
