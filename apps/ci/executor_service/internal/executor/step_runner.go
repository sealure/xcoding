package executor

import (
	"fmt"
	"strings"
	"xcoding/apps/ci/executor_service/internal/parser"
)

// BuildStepCommand 将单步命令包装为捕获退出码并输出标记
// 约定：支持 continue_on_error（通过 step.Env["XC_CONTINUE_ON_ERROR"] == "true"）
func BuildStepCommand(st parser.Step) string {
	cmd := strings.TrimSpace(st.Run)
	if cmd == "" {
		return ""
	}
	// 捕获退出码并输出特定标记，供日志处理器识别
	// 退出码输出格式：__step_exit__ <name> <code>
	var b strings.Builder
	// 若允许继续错误，需要暂时关闭 set -e，以便命令失败不立即终止脚本
	if strings.EqualFold(strings.TrimSpace(st.Env["XC_CONTINUE_ON_ERROR"]), "true") {
		fmt.Fprintf(&b, "set +e\n")
	}
	fmt.Fprintf(&b, "%s\n", cmd)
	fmt.Fprintf(&b, "code=$?; echo %s %s $code\n", MarkerStepExit, st.Name)
	// 若允许继续，则不 set -e 失败；脚本顶层已有 set -e
	if strings.EqualFold(strings.TrimSpace(st.Env["XC_CONTINUE_ON_ERROR"]), "true") {
		// 恢复 set -e，后续步骤仍旧严格
		fmt.Fprintf(&b, "set -e\n")
	} else {
		fmt.Fprintf(&b, "if [ $code -ne 0 ]; then exit $code; fi\n")
	}
	return b.String()
}
