package models

import (
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
	coderepositoryv1 "xcoding/gen/go/code_repository/v1"
)

type Repository struct {
	ID          uint64 `gorm:"primaryKey"`
	ProjectID   uint64 `gorm:"index:idx_repo_project_name,unique"`
	Name        string `gorm:"size:255;index:idx_repo_project_name,unique"`
	Description string `gorm:"type:text"`
	GitURL      string `gorm:"size:1024"`
	AuthType    int32  `gorm:"type:int"`
	GitUsername string `gorm:"size:255"`
	GitPassword string `gorm:"size:1024"`
	GitSSHKey   string `gorm:"type:text"`
	IsActive    bool   `gorm:"default:true"`
	LastSyncAt  *time.Time
	SyncStatus  int32  `gorm:"type:int"`
	SyncMessage string `gorm:"type:text"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func ToProto(m *Repository) *coderepositoryv1.Repository {
	if m == nil {
		return nil
	}
	var lastSyncAt *timestamppb.Timestamp
	if m.LastSyncAt != nil {
		lastSyncAt = timestamppb.New(*m.LastSyncAt)
	}
	return &coderepositoryv1.Repository{
		Id:          m.ID,
		ProjectId:   m.ProjectID,
		Name:        m.Name,
		Description: m.Description,
		GitUrl:      m.GitURL,
		Branch:      "",
		AuthType:    coderepositoryv1.RepositoryAuthType(m.AuthType),
		GitUsername: m.GitUsername,
		IsActive:    m.IsActive,
		LastSyncAt:  lastSyncAt,
		SyncStatus:  coderepositoryv1.RepositorySyncStatus(m.SyncStatus),
		SyncMessage: m.SyncMessage,
		CreatedAt:   timestamppb.New(m.CreatedAt),
		UpdatedAt:   timestamppb.New(m.UpdatedAt),
	}
}

func ToProtoWithBranch(m *Repository, branch string) *coderepositoryv1.Repository {
	if m == nil {
		return nil
	}
	var lastSyncAt *timestamppb.Timestamp
	if m.LastSyncAt != nil {
		lastSyncAt = timestamppb.New(*m.LastSyncAt)
	}
	return &coderepositoryv1.Repository{
		Id:          m.ID,
		ProjectId:   m.ProjectID,
		Name:        m.Name,
		Description: m.Description,
		GitUrl:      m.GitURL,
		Branch:      branch,
		AuthType:    coderepositoryv1.RepositoryAuthType(m.AuthType),
		GitUsername: m.GitUsername,
		IsActive:    m.IsActive,
		LastSyncAt:  lastSyncAt,
		SyncStatus:  coderepositoryv1.RepositorySyncStatus(m.SyncStatus),
		SyncMessage: m.SyncMessage,
		CreatedAt:   timestamppb.New(m.CreatedAt),
		UpdatedAt:   timestamppb.New(m.UpdatedAt),
	}
}

func UpdateModelFromRequest(m *Repository, req *coderepositoryv1.UpdateRepositoryRequest) {
	if req.Name != "" {
		m.Name = req.Name
	}
	if req.Description != "" {
		m.Description = req.Description
	}
	if req.GitUrl != "" {
		m.GitURL = req.GitUrl
	}
	// Branch 更新由分支表维护，不在仓库模型上处理
	// AuthType 0 is unspecified; only set when non-zero
	if req.AuthType != coderepositoryv1.RepositoryAuthType_REPOSITORY_AUTH_TYPE_UNSPECIFIED {
		m.AuthType = int32(req.AuthType)
	}
	if req.GitUsername != "" {
		m.GitUsername = req.GitUsername
	}
	if req.GitPassword != "" {
		m.GitPassword = req.GitPassword
	}
	if req.GitSshKey != "" {
		m.GitSSHKey = req.GitSshKey
	}
	// is_active 的布尔值无法判断是否提供，默认覆盖
	m.IsActive = req.IsActive
}
