package models

import (
	"time"

	artifactv1 "xcoding/gen/go/artifact/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

// Namespace 命名空间模型
type Namespace struct {
	ID          uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string         `gorm:"not null;size:100" json:"name"`
	Description string         `gorm:"type:text" json:"description"`
	RegistryID  uint64         `gorm:"not null;index" json:"registry_id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	Registry     Registry     `gorm:"foreignKey:RegistryID;references:ID" json:"registry,omitempty"`
	Repositories []Repository `gorm:"foreignKey:NamespaceID;references:ID" json:"repositories,omitempty"`
}

func (Namespace) TableName() string {
	return "namespaces"
}

func (n *Namespace) ToProto() *artifactv1.Namespace {
	return &artifactv1.Namespace{
		Id:          n.ID,
		Name:        n.Name,
		Description: n.Description,
		RegistryId:  n.RegistryID,
		CreatedAt:   timestamppb.New(n.CreatedAt),
		UpdatedAt:   timestamppb.New(n.UpdatedAt),
	}
}

func (n *Namespace) ToProtoWithRegistry() *artifactv1.Namespace {
	namespace := n.ToProto()
	if n.Registry.ID != 0 {
		namespace.Registry = n.Registry.ToProto()
	}
	return namespace
}

func (n *Namespace) FromProto(namespace *artifactv1.Namespace) {
	if namespace == nil {
		return
	}

	n.ID = namespace.Id
	n.Name = namespace.Name
	n.Description = namespace.Description
	n.RegistryID = namespace.RegistryId

	if namespace.CreatedAt != nil {
		n.CreatedAt = namespace.CreatedAt.AsTime()
	}

	if namespace.UpdatedAt != nil {
		n.UpdatedAt = namespace.UpdatedAt.AsTime()
	}
}
