package actions

import (
	"xcoding/apps/ci/executor_service/internal/parser"
)

// Action 抽象接口：将一个 uses 步骤转为可执行脚本片段
// 约定：返回的脚本可直接拼接到 runner 的 /bin/sh -c 脚本中
type Action interface {
	Build(step parser.Step, job parser.Job) (string, error)
}

// ParsedRef 结构化的 uses 引用
// 例如：owner/name@version -> Owner=owner, Name=name, Version=version
type ParsedRef struct {
	Owner   string
	Name    string
	Path    string
	Version string
}

// ResolvedAction 解析后的远端 Action 元数据（精简版）
// Using: composite|node|docker
// Path: 本地缓存路径（包含 action.yml 等）
type ResolvedAction struct {
	Using     string
	Path      string
	Main      string
	Composite []CompositeStepMeta
}
