package models

import (
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
	coderepositoryv1 "xcoding/gen/go/code_repository/v1"
)

// RepositoryBranch 表示某仓库的一个分支
// 设计为 (repository_id, name) 唯一，以便同仓库下分支名不重复
// 通过 GORM 的复合唯一索引 idx_repo_branch_repo_name 约束实现
// 通过 is_default 标记仓库的默认分支（应用层保证仅一个为 true）
type RepositoryBranch struct {
	ID           uint64 `gorm:"primaryKey"`
	RepositoryID uint64 `gorm:"index:idx_repo_branch_repo_name,unique"`
	Name         string `gorm:"size:255;index:idx_repo_branch_repo_name,unique"`
	IsDefault    bool   `gorm:"default:false"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (RepositoryBranch) TableName() string { return "branches" }

// ToProto 将模型转换为 proto Branch
func ToProtoBranch(m *RepositoryBranch) *coderepositoryv1.Branch {
	if m == nil {
		return nil
	}
	return &coderepositoryv1.Branch{
		Id:           m.ID,
		RepositoryId: m.RepositoryID,
		Name:         m.Name,
		IsDefault:    m.IsDefault,
		CreatedAt:    timestamppb.New(m.CreatedAt),
		UpdatedAt:    timestamppb.New(m.UpdatedAt),
	}
}
