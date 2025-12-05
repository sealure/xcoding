package models

import (
	"time"
	civ1 "xcoding/gen/go/ci/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/datatypes"
)

type Build struct {
	ID          uint64            `gorm:"primaryKey;autoIncrement" json:"id"`                              // 构建 ID, 加json是websocket
	PipelineID  uint64            `gorm:"index;index:idx_build_pid_created,priority:1" json:"pipeline_id"` // 流水线 ID
	Name        string            `gorm:"size:255;not null" json:"name"`                                   // 构建名称
	Status      int32             `gorm:"not null" json:"status"`                                          // civ1.BuildStatus 枚举值
	TriggeredBy string            `gorm:"size:128" json:"triggered_by"`                                    // 触发者
	CommitSHA   string            `gorm:"size:64" json:"commit_sha"`                                       // 提交 SHA
	Branch      string            `gorm:"size:128" json:"branch"`
	Variables   datatypes.JSONMap `gorm:"type:jsonb" json:"variables"`
	CreatedAt   time.Time         `gorm:"autoCreateTime;index:idx_build_pid_created,priority:2" json:"created_at"`
	StartedAt   *time.Time        `gorm:"" json:"started_at"`
	FinishedAt  *time.Time        `gorm:"" json:"finished_at"`
}

func (b *Build) ToProto() *civ1.Build {
	if b == nil {
		return nil
	}
	var vars map[string]string
	if b.Variables != nil {
		vars = make(map[string]string, len(b.Variables))
		for k, v := range b.Variables {
			if sv, ok := v.(string); ok {
				vars[k] = sv
			}
		}
	}
	pb := &civ1.Build{
		Id:          b.ID,
		PipelineId:  b.PipelineID,
		Name:        b.Name,
		Status:      civ1.BuildStatus(b.Status),
		TriggeredBy: b.TriggeredBy,
		CommitSha:   b.CommitSHA,
		Branch:      b.Branch,
		Variables:   vars,
		CreatedAt:   timestamppb.New(b.CreatedAt),
	}
	if b.StartedAt != nil {
		pb.StartedAt = timestamppb.New(*b.StartedAt)
	}
	if b.FinishedAt != nil {
		pb.FinishedAt = timestamppb.New(*b.FinishedAt)
	}
	return pb
}

func (b *Build) FromProto(pb *civ1.Build) {
	if pb == nil {
		return
	}
	b.ID = pb.GetId()
	b.PipelineID = pb.GetPipelineId()
	b.Name = pb.GetName()
	b.Status = int32(pb.GetStatus())
	b.TriggeredBy = pb.GetTriggeredBy()
	b.CommitSHA = pb.GetCommitSha()
	b.Branch = pb.GetBranch()
	if pb.GetVariables() != nil {
		jm := datatypes.JSONMap{}
		for k, v := range pb.GetVariables() {
			jm[k] = v
		}
		b.Variables = jm
	}
	if pb.GetCreatedAt() != nil {
		b.CreatedAt = pb.GetCreatedAt().AsTime()
	}
	if pb.GetStartedAt() != nil {
		t := pb.GetStartedAt().AsTime()
		b.StartedAt = &t
	}
	if pb.GetFinishedAt() != nil {
		t := pb.GetFinishedAt().AsTime()
		b.FinishedAt = &t
	}
}
