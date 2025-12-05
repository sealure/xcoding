package models

import (
	"time"

	projectv1 "xcoding/gen/go/project/v1"
	userv1 "xcoding/gen/go/user/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

type Project struct {
	ID          uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string         `gorm:"uniqueIndex;not null;size:100" json:"name"`
	Description string         `gorm:"type:text" json:"description"`
	Language    string         `gorm:"size:50" json:"language"`
	Framework   string         `gorm:"size:50" json:"framework"`
	IsPublic    bool           `gorm:"not null;default:false" json:"is_public"`
	Status      string         `gorm:"size:50" json:"status"`
	OwnerID     uint64         `gorm:"not null;index" json:"owner_id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Project) TableName() string { return "projects" }

func (p *Project) ToProto() *projectv1.Project {
	return &projectv1.Project{
		Id:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Language:    p.Language,
		Framework:   p.Framework,
		IsPublic:    p.IsPublic,
		Status:      p.Status,
		OwnerId:     p.OwnerID,
		CreatedAt:   timestamppb.New(p.CreatedAt),
		UpdatedAt:   timestamppb.New(p.UpdatedAt),
	}
}

// Use composite primary key on (project_id, user_id) to support updates
// and prevent duplicate rows for the same member.
type ProjectMember struct {
	UserID    uint64    `gorm:"primaryKey;not null" json:"user_id"`
	Username  string    `gorm:"size:100" json:"username"`
	ProjectID uint64    `gorm:"primaryKey;not null" json:"project_id"`
	Role      int32     `gorm:"not null;default:0" json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (ProjectMember) TableName() string { return "project_members" }

func (pm *ProjectMember) ToProto() *projectv1.ProjectMember {
	proto := &projectv1.ProjectMember{
		UserId:    pm.UserID,
		ProjectId: pm.ProjectID,
		Role:      projectv1.ProjectMemberRole(pm.Role),
		CreatedAt: timestamppb.New(pm.CreatedAt),
		UpdatedAt: timestamppb.New(pm.UpdatedAt),
	}
	if pm.Username != "" {
		proto.User = &userv1.User{Username: pm.Username}
	}
	return proto
}
