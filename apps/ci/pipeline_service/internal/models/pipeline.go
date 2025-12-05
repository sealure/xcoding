package models

import (
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	civ1 "xcoding/gen/go/ci/v1"
)

// Pipeline 表示 CI 流水线/工作流的定义（与 proto 保持一致）
type Pipeline struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement"`
	ProjectID    uint64    `gorm:"index;uniqueIndex:ux_pipeline_project_name"`
	Name         string    `gorm:"size:255;index;uniqueIndex:ux_pipeline_project_name"`
	Description  string    `gorm:"type:text"`
	WorkflowYAML string    `gorm:"type:text"`
	IsActive     bool      `gorm:"default:true"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime"`
}

// ==== Conversions to/from proto ====

func (p *Pipeline) ToProto() *civ1.Pipeline {
	if p == nil {
		return nil
	}
	return &civ1.Pipeline{
		Id:           p.ID,
		ProjectId:    p.ProjectID,
		Name:         p.Name,
		Description:  p.Description,
		WorkflowYaml: p.WorkflowYAML,
		IsActive:     p.IsActive,
		CreatedAt:    timestamppb.New(p.CreatedAt),
		UpdatedAt:    timestamppb.New(p.UpdatedAt),
	}
}

func (p *Pipeline) FromProto(pp *civ1.Pipeline) {
	if pp == nil {
		return
	}
	p.ID = pp.GetId()
	p.ProjectID = pp.GetProjectId()
	p.Name = pp.GetName()
	p.Description = pp.GetDescription()
	p.WorkflowYAML = pp.GetWorkflowYaml()
	p.IsActive = pp.GetIsActive()
	if pp.GetCreatedAt() != nil {
		p.CreatedAt = pp.GetCreatedAt().AsTime()
	}
	if pp.GetUpdatedAt() != nil {
		p.UpdatedAt = pp.GetUpdatedAt().AsTime()
	}
}
