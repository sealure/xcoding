package models

import "time"

// BuildSnapshot 保存一次构建时的 Workflow YAML 快照
// 设计意图：保证可重现性，不改动外部 proto，仅在内部持久化。
type BuildSnapshot struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement"`
	BuildID      uint64    `gorm:"uniqueIndex;not null"` // 每个构建仅一份快照
	PipelineID   uint64    `gorm:"index;not null"`
	Name         string    `gorm:"size:255;not null"` // 构建名称
	WorkflowYAML string    `gorm:"type:text"`
	YamlSHA256   string    `gorm:"size:64;index"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
}
