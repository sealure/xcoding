package models

import (
	"time"

	artifactv1 "xcoding/gen/go/artifact/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

// Tag 标签模型
type Tag struct {
	ID           uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	Name         string         `gorm:"not null;size:100" json:"name"`
	Digest       string         `gorm:"not null;size:255" json:"digest"`
	SizeBytes    int64          `gorm:"not null;default:0" json:"size_bytes"`
	RepositoryID uint64         `gorm:"not null;index" json:"repository_id"`
	IsLatest     bool           `gorm:"not null;default:false" json:"is_latest"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	Repository Repository `gorm:"foreignKey:RepositoryID;references:ID" json:"repository,omitempty"`
}

func (Tag) TableName() string {
	return "tags"
}

func (t *Tag) ToProto() *artifactv1.Tag {
	return &artifactv1.Tag{
		Id:           t.ID,
		Name:         t.Name,
		Digest:       t.Digest,
		SizeBytes:    uint64(t.SizeBytes),
		RepositoryId: t.RepositoryID,
		CreatedAt:    timestamppb.New(t.CreatedAt),
		UpdatedAt:    timestamppb.New(t.UpdatedAt),
	}
}

func (t *Tag) ToProtoWithRepository() *artifactv1.Tag {
	tag := t.ToProto()
	if t.Repository.ID != 0 {
		tag.Repository = t.Repository.ToProto()
	}
	return tag
}

func (t *Tag) FromProto(tag *artifactv1.Tag) {
	if tag == nil {
		return
	}

	t.ID = tag.Id
	t.Name = tag.Name
	t.Digest = tag.Digest
	t.SizeBytes = int64(tag.SizeBytes)
	t.RepositoryID = tag.RepositoryId

	if tag.CreatedAt != nil {
		t.CreatedAt = tag.CreatedAt.AsTime()
	}

	if tag.UpdatedAt != nil {
		t.UpdatedAt = tag.UpdatedAt.AsTime()
	}
}
