package executor

import (
	"xcoding/apps/ci/executor_service/internal/parser"

	"github.com/sirupsen/logrus"
)

type DAG struct {
	Jobs       map[string]parser.Job // 任务信息
	Needs      map[string][]string   // 前置依赖
	Dependents map[string][]string   // 后置依赖
}

// BuildDAG 根据工作流构造简单 DAG 结构（用于并发调度）
func BuildDAG(wf *parser.Workflow) *DAG {
	d := &DAG{Jobs: map[string]parser.Job{}, Needs: map[string][]string{}, Dependents: map[string][]string{}}
	for name, j := range wf.Jobs {
		d.Jobs[name] = j
		d.Needs[name] = append([]string{}, j.Needs...)
		for _, n := range j.Needs {
			d.Dependents[n] = append(d.Dependents[n], name)
		}
	}
	logrus.Infof("build dag: %v", d)
	return d
}
