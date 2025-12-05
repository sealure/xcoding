package models

import (
	"time"

	artifactv1 "xcoding/gen/go/artifact/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

// Repository 仓库模型
type Repository struct {
	ID          uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string         `gorm:"not null;size:100" json:"name"`
	Description string         `gorm:"type:text" json:"description"`
	NamespaceID uint64         `gorm:"not null;index" json:"namespace_id"`
	IsPublic    bool           `gorm:"not null;default:true" json:"is_public"`
	Path        string         `gorm:"size:255" json:"path"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	Namespace Namespace `gorm:"foreignKey:NamespaceID;references:ID" json:"namespace,omitempty"`
	Tags      []Tag     `gorm:"foreignKey:RepositoryID;references:ID" json:"tags,omitempty"`
}

func (Repository) TableName() string {
	return "repositories"
}

func (r *Repository) ToProto() *artifactv1.Repository {
	return &artifactv1.Repository{
		Id:          r.ID,
		Name:        r.Name,
		Description: r.Description,
		NamespaceId: r.NamespaceID,
		IsPublic:    r.IsPublic,
		CreatedAt:   timestamppb.New(r.CreatedAt),
		UpdatedAt:   timestamppb.New(r.UpdatedAt),
		Path:        r.Path,
	}
}

func (r *Repository) ToProtoWithNamespace() *artifactv1.Repository {
	repository := r.ToProto()
	if r.Namespace.ID != 0 {
		repository.Namespace = r.Namespace.ToProto()
	}
	return repository
}

func (r *Repository) FromProto(repository *artifactv1.Repository) {
	if repository == nil {
		return
	}

	r.ID = repository.Id
	r.Name = repository.Name
	r.Description = repository.Description
	r.NamespaceID = repository.NamespaceId
	r.IsPublic = repository.IsPublic
	r.Path = repository.Path

	if repository.CreatedAt != nil {
		r.CreatedAt = repository.CreatedAt.AsTime()
	}

	if repository.UpdatedAt != nil {
		r.UpdatedAt = repository.UpdatedAt.AsTime()
	}
}
