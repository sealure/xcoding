package executor

import (
	"context"
	"sync"
	"time"
	"xcoding/apps/ci/executor_service/internal/parser"
	"xcoding/apps/ci/executor_service/models"
	civ1 "xcoding/gen/go/ci/v1"

	"gorm.io/gorm"
)

type Engine struct {
    Env *K8sEnv
    DB  *gorm.DB
}

// NewEngine 创建工作流执行引擎：负责 DAG 并发运行
func NewEngine(env *K8sEnv, db *gorm.DB, _ func(ctx context.Context, buildID uint64, seq uint64, line string)) *Engine {
    return &Engine{Env: env, DB: db}
}

// RunWorkflow 并发运行工作流（按 needs 约束），基础版本：每个 Job 仅运行第一步的 run
// 主要职责：
// - 基于 workflow 构建 DAG，识别就绪的 Job 并发执行
// - 跟踪每个 Job 的运行状态（pending/running/succeeded/failed）
// - Job 完成后检查其 dependents 是否满足依赖，从而推进下一批就绪 Job
// - 所有 Job 完成后，按严格规则计算构建终态：
//   * 存在任意 failed → Build=FAILED
//   * 全部 succeeded → Build=SUCCEEDED
//   * 其它（仍有 running/pending）→ Build=RUNNING（不写 finished_at）
func (e *Engine) RunWorkflow(ctx context.Context, buildID uint64, wf *parser.Workflow) error {
	dag := BuildDAG(wf)
	// 将全局 workflow 环境变量合并到每个 job 的环境变量中
	if len(wf.Env) > 0 {
		for name, job := range dag.Jobs {
			newEnv := make(map[string]string)
			// 1. 添加全局环境变量
			for k, v := range wf.Env {
				newEnv[k] = v
			}
			// 2. 使用 Job 级环境变量覆盖（优先级更高）
			for k, v := range job.Env {
				newEnv[k] = v
			}
			job.Env = newEnv
			dag.Jobs[name] = job
		}
	}
	state := map[string]string{} // pending/running/succeeded/failed
	ready := []string{}
	for name := range dag.Jobs {
		if len(dag.Needs[name]) == 0 {
			ready = append(ready, name)
		} else {
			state[name] = "pending"
		}
	}
	var mu sync.Mutex
	var wg sync.WaitGroup

	// 标记状态函数
	mark := func(job string, st string) {
		mu.Lock()
		state[job] = st
		mu.Unlock()
	}

	runJob := func(name string) {
		defer wg.Done()
		j := dag.Jobs[name]
		sched := NewScheduler(e.Env, e.DB)
		if err := sched.RunSingleJob(ctx, buildID, name, j); err != nil {
			mark(name, "failed")
		} else {
			mark(name, "succeeded")
		}
		mu.Lock()
		for _, dep := range dag.Dependents[name] {
			// 检查 dep 的 needs 是否均完成
			ok := true
			for _, n := range dag.Needs[dep] {
				if state[n] != "succeeded" {
					ok = false
					break
				}
			}
			if ok && state[dep] == "pending" {
				ready = append(ready, dep)
			}
		}
		mu.Unlock()
	}
	for len(ready) > 0 {
		mu.Lock()
		batch := append([]string{}, ready...)
		ready = ready[:0]
		mu.Unlock()
		for _, name := range batch {
			wg.Add(1)
			mark(name, "running")
			go runJob(name)
		}
		wg.Wait()
	}
    // 结束状态更新：仅当所有 Job 成功时才标记构建为 SUCCEEDED；存在失败则 FAILED；否则 RUNNING
    now := time.Now()
    total := len(dag.Jobs)
    succeeded := 0
    failed := false
    for name := range dag.Jobs {
        st := state[name]
        switch st {
        case "failed":
            failed = true
        case "succeeded":
            succeeded++
        case "running", "pending", "":
            // 未完成或未调度，保持运行态
        }
        if failed {
            break
        }
    }
    var status civ1.BuildStatus = civ1.BuildStatus_BUILD_STATUS_RUNNING
    if failed {
        status = civ1.BuildStatus_BUILD_STATUS_FAILED
    } else if succeeded == total {
        status = civ1.BuildStatus_BUILD_STATUS_SUCCEEDED
    }
    if status == civ1.BuildStatus_BUILD_STATUS_SUCCEEDED || status == civ1.BuildStatus_BUILD_STATUS_FAILED {
        _ = e.DB.Model(&models.Build{}).Where("id = ?", buildID).Updates(map[string]any{"status": int32(status), "finished_at": &now}).Error
    } else {
        _ = e.DB.Model(&models.Build{}).Where("id = ?", buildID).Updates(map[string]any{"status": int32(status)}).Error
    }
	return nil
}
