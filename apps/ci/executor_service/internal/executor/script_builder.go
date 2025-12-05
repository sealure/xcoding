package executor

import (
	"fmt"
	"strings"
	act "xcoding/apps/ci/executor_service/internal/executor/actions"
	"xcoding/apps/ci/executor_service/internal/parser"
)

// BuildScript 生成在容器中执行的 Bash 脚本
// 说明：
// - 顶层启用 set -e；step 层按需覆盖
// - 导出 Job 级非敏感环境变量
// - 按步骤输出 __step_begin__/__step_end__/__step_exit__ 标记，便于日志解析
func BuildScript(job parser.Job) string {
	var b strings.Builder
	fmt.Fprintf(&b, "set -e\n")
	//b.WriteString("mkdir -p /workspace\n")
	//b.WriteString("cd /workspace\n")

	// 统一：不在脚本中 export Job 级 env，均通过 K8s EnvVar 注入

	//  添加step
	for _, st := range job.Steps {
		fmt.Fprintf(&b, "echo %s %s\n", MarkerStepBegin, st.Name)

		if strings.TrimSpace(st.Uses) != "" {
			frag, err := act.BuildUsesScript(st, job)

			if err != nil {
				fmt.Fprintf(&b, "echo \"action error: %s\"\nexit 1\n", strings.ReplaceAll(err.Error(), "\"", "\\\""))
			} else {
				fmt.Fprintf(&b, "%s\n", frag)
			}
		} else if strings.TrimSpace(st.Run) != "" {
			var inline []string
			for k, v := range st.Env {
				if strings.HasPrefix(strings.TrimSpace(v), "secret://") {
					continue
				}
				inline = append(inline, k+"="+v)
			}
			if len(inline) > 0 {
				fmt.Fprintf(&b, "%s ", strings.Join(inline, " "))
			}
			fmt.Fprintf(&b, "%s\n", BuildStepCommand(st))
		}
		fmt.Fprintf(&b, "echo %s %s\n", MarkerStepEnd, st.Name)
	}
	return b.String()
}
