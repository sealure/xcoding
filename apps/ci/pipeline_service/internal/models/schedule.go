package models

import (
	"time"
	civ1 "xcoding/gen/go/ci/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// PipelineSchedule 表示流水线的定时（cron）计划（与 proto 保持一致）
type PipelineSchedule struct {
	ID              uint64 `gorm:"primaryKey;autoIncrement"`
	PipelineID      uint64 `gorm:"index"`
	Cron            string `gorm:"size:128"`
	Timezone        string `gorm:"size:64;default:UTC"`
	Enabled         bool   `gorm:"default:true"`
	LastTriggeredAt *time.Time
	CreatedAt       time.Time `gorm:"autoCreateTime"`
	UpdatedAt       time.Time `gorm:"autoUpdateTime"`
}

func (s *PipelineSchedule) FromProto(ps *civ1.PipelineSchedule) {
	if ps == nil {
		return
	}
	s.ID = ps.GetId()
	s.PipelineID = ps.GetPipelineId()
	s.Cron = ps.GetCron()
	s.Timezone = ps.GetTimezone()
	s.Enabled = ps.GetEnabled()
	if ps.GetCreatedAt() != nil {
		s.CreatedAt = ps.GetCreatedAt().AsTime()
	}
	if ps.GetUpdatedAt() != nil {
		s.UpdatedAt = ps.GetUpdatedAt().AsTime()
	}
	if ps.GetLastTriggeredAt() != nil {
		t := ps.GetLastTriggeredAt().AsTime()
		s.LastTriggeredAt = &t
	}
}

func (s *PipelineSchedule) ToProto() *civ1.PipelineSchedule {
	if s == nil {
		return nil
	}
	ps := &civ1.PipelineSchedule{
		Id:         s.ID,
		PipelineId: s.PipelineID,
		Cron:       s.Cron,
		Timezone:   s.Timezone,
		Enabled:    s.Enabled,
		CreatedAt:  timestamppb.New(s.CreatedAt),
		UpdatedAt:  timestamppb.New(s.UpdatedAt),
	}
	if s.LastTriggeredAt != nil {
		ps.LastTriggeredAt = timestamppb.New(*s.LastTriggeredAt)
	}
	return ps
}
