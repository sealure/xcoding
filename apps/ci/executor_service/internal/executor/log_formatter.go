package executor

import (
	"fmt"
	"strings"
)

// FormatStepLog 将内部标记转换为用户友好的日志格式
// 返回：(格式化后的日志, 是否需要追加到用户日志)
func FormatStepLog(line string) (string, bool) {
	s := strings.TrimSpace(line)

	// 处理步骤开始
	if strings.HasPrefix(s, MarkerStepBegin+" ") {
		name := strings.TrimSpace(strings.TrimPrefix(s, MarkerStepBegin+" "))
		return fmt.Sprintf("🔹 Step [%s] Running", name), true
	}

	// 处理步骤结束（可选，如果觉得太吵可以返回 false）
	if strings.HasPrefix(s, MarkerStepEnd+" ") {
		// name := strings.TrimSpace(strings.TrimPrefix(s, MarkerStepEnd+" "))
		// return fmt.Sprintf("🔹 Step [%s] Succeeded", name), true
		return "", false // 暂时不展示结束，保持简洁
	}

	// 处理步骤退出（不展示）
	if strings.HasPrefix(s, MarkerStepExit+" ") {
		return "", false
	}

	// 普通日志，原样返回
	return line, true
}
