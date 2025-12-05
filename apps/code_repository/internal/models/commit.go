package models

import (
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
	coderepositoryv1 "xcoding/gen/go/code_repository/v1"
)

// Commit 表示一次代码提交，属于某个仓库的某个分支
// 通过 (branch_id, hash) 唯一约束确保同分支提交哈希不重复
// 如需联表查询仓库，可通过关联 Branch -> Repository
type Commit struct {
	ID             uint64 `gorm:"primaryKey"`
	BranchID       uint64 `gorm:"index:idx_branch_hash,unique"`
	Hash           string `gorm:"size:64;index:idx_branch_hash,unique"`
	Message        string `gorm:"type:text"`
	AuthorName     string `gorm:"size:255"`
	AuthorEmail    string `gorm:"size:255"`
	AuthoredAt     *time.Time
	CommitterName  string     `gorm:"size:255"`
	CommitterEmail string     `gorm:"size:255"`
	CommittedAt    *time.Time `gorm:"index"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (Commit) TableName() string { return "commits" }

// ToProtoCommit 将模型转换为 proto Commit
func ToProtoCommit(m *Commit) *coderepositoryv1.Commit {
	if m == nil {
		return nil
	}
	var authoredAtTs, committedAtTs *timestamppb.Timestamp
	if m.AuthoredAt != nil {
		authoredAtTs = timestamppb.New(*m.AuthoredAt)
	}
	if m.CommittedAt != nil {
		committedAtTs = timestamppb.New(*m.CommittedAt)
	}
	return &coderepositoryv1.Commit{
		Id:             m.ID,
		BranchId:       m.BranchID,
		Hash:           m.Hash,
		Message:        m.Message,
		AuthorName:     m.AuthorName,
		AuthorEmail:    m.AuthorEmail,
		AuthoredAt:     authoredAtTs,
		CommitterName:  m.CommitterName,
		CommitterEmail: m.CommitterEmail,
		CommittedAt:    committedAtTs,
		CreatedAt:      timestamppb.New(m.CreatedAt),
		UpdatedAt:      timestamppb.New(m.UpdatedAt),
	}
}
