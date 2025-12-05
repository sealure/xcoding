package models

import (
	"time"

	artifactv1 "xcoding/gen/go/artifact/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

// Registry 注册表模型
type Registry struct {
	ID             uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	Name           string         `gorm:"uniqueIndex;not null;size:100" json:"name"`
	URL            string         `gorm:"not null;size:255" json:"url"`
	Description    string         `gorm:"type:text" json:"description"`
	IsPublic       bool           `gorm:"not null;default:true" json:"is_public"`
	Username       string         `gorm:"size:100" json:"username"`
	Password       string         `gorm:"size:255" json:"-"`
	ProjectID      uint64         `gorm:"not null;index" json:"project_id"`
	ArtifactType   ArtifactType   `gorm:"not null;default:0" json:"artifact_type"`
	ArtifactSource ArtifactSource `gorm:"not null;default:0" json:"artifact_source"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`

	Namespaces []Namespace `gorm:"foreignKey:RegistryID;references:ID" json:"namespaces,omitempty"`
}

func (Registry) TableName() string {
	return "registries"
}

func (r *Registry) ToProto() *artifactv1.Registry {
	return &artifactv1.Registry{
		Id:             r.ID,
		Name:           r.Name,
		Url:            r.URL,
		Description:    r.Description,
		IsPublic:       r.IsPublic,
		Username:       r.Username,
		Password:       r.Password,
		ProjectId:      r.ProjectID,
		CreatedAt:      timestamppb.New(r.CreatedAt),
		UpdatedAt:      timestamppb.New(r.UpdatedAt),
		ArtifactType:   r.ArtifactType.ToProto(),
		ArtifactSource: r.ArtifactSource.ToProto(),
	}
}

func (r *Registry) FromProto(registry *artifactv1.Registry) {
	if registry == nil {
		return
	}

	r.ID = registry.Id
	r.Name = registry.Name
	r.URL = registry.Url
	r.Description = registry.Description
	r.IsPublic = registry.IsPublic
	r.Username = registry.Username
	r.Password = registry.Password
	r.ProjectID = registry.ProjectId
	r.ArtifactType = ArtifactTypeFromProto(registry.ArtifactType)
	r.ArtifactSource = ArtifactSourceFromProto(registry.ArtifactSource)

	if registry.CreatedAt != nil {
		r.CreatedAt = registry.CreatedAt.AsTime()
	}

	if registry.UpdatedAt != nil {
		r.UpdatedAt = registry.UpdatedAt.AsTime()
	}
}
