package models

import "time"

type BuildJob struct {
	ID         uint64 `gorm:"primaryKey;autoIncrement"`
	BuildID    uint64 `gorm:"index"`
	Name       string
	Status     string
	StartedAt  *time.Time
	FinishedAt *time.Time
	Index      int32
}
type BuildJobEdge struct {
	ID      uint64 `gorm:"primaryKey;autoIncrement"`
	BuildID uint64 `gorm:"index"`
	FromJob string
	ToJob   string
}
type BuildStep struct {
	ID         uint64 `gorm:"primaryKey;autoIncrement"`
	BuildID    uint64 `gorm:"index"`
	JobName    string
	Index      int32
	Name       string
	Status     string
	StartedAt  *time.Time
	FinishedAt *time.Time
	ExitCode   *int32
}

type BuildStepLogChunk struct {
	ID          uint64 `gorm:"primaryKey;autoIncrement"`
	BuildStepID uint64 `gorm:"index"`
	Content     string `gorm:"type:text"`
	CreatedAt   time.Time
}
