package models

import (
	"time"

	userv1 "xcoding/gen/go/user/v1"

	"github.com/lib/pq"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

type User struct {
	ID        uint64          `gorm:"primaryKey;autoIncrement" json:"id"`
	Username  string          `gorm:"uniqueIndex;not null;size:50" json:"username"`
	Email     string          `gorm:"uniqueIndex;not null;size:100" json:"email"`
	Password  string          `gorm:"not null" json:"-"`
	Avatar    string          `gorm:"size:255" json:"avatar"`
	Role      userv1.UserRole `gorm:"not null;default:0" json:"role"`
	IsActive  bool            `gorm:"not null;default:true" json:"is_active"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	DeletedAt gorm.DeletedAt  `gorm:"index" json:"-"`
}

func (User) TableName() string {
	return "users"
}

type APIToken struct {
	ID          uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID      uint64         `gorm:"not null;index" json:"user_id"`
	Name        string         `gorm:"not null;size:100" json:"name"`
	TokenHash   string         `gorm:"uniqueIndex;not null;size:255" json:"-"`
	Description string         `gorm:"type:text" json:"description"`
	Scopes      pq.StringArray `gorm:"type:text[];default:'{}'" json:"scopes"`
	ExpiresAt   *time.Time     `json:"expires_at"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	User User `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`
}

func (APIToken) TableName() string {
	return "api_tokens"
}

func (u *User) ToProto() *userv1.User {
	return &userv1.User{
		Id:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		Avatar:    u.Avatar,
		Role:      u.Role,
		IsActive:  u.IsActive,
		CreatedAt: timestamppb.New(u.CreatedAt),
		UpdatedAt: timestamppb.New(u.UpdatedAt),
	}
}

func (u *User) FromProto(user *userv1.User) {
	if user == nil {
		return
	}

	u.ID = user.Id
	u.Username = user.Username
	u.Email = user.Email
	u.Avatar = user.Avatar
	u.Role = user.Role
	u.IsActive = user.IsActive

	if user.CreatedAt != nil {
		u.CreatedAt = user.CreatedAt.AsTime()
	}

	if user.UpdatedAt != nil {
		u.UpdatedAt = user.UpdatedAt.AsTime()
	}
}

func (t *APIToken) ToProto() *userv1.CreateAPITokenResponse {
	token := &userv1.CreateAPITokenResponse{
		Id:          t.ID,
		Name:        t.Name,
		Description: t.Description,
		Scopes:      make([]userv1.Scope, 0, len(t.Scopes)),
		CreatedAt:   timestamppb.New(t.CreatedAt),
	}

	if t.ExpiresAt != nil {
		token.ExpiresAt = timestamppb.New(*t.ExpiresAt)
	}

	for _, scope := range t.Scopes {
		if s, ok := userv1.Scope_value[scope]; ok {
			token.Scopes = append(token.Scopes, userv1.Scope(s))
		}
	}

	return token
}

func (t *APIToken) FromProto(token *userv1.CreateAPITokenResponse) {
	if token == nil {
		return
	}

	t.ID = token.Id
	t.Name = token.Name
	t.Description = token.Description
	t.Scopes = make([]string, 0, len(token.Scopes))

	for _, scope := range token.Scopes {
		t.Scopes = append(t.Scopes, scope.String())
	}

	if token.ExpiresAt != nil {
		expiresAt := token.ExpiresAt.AsTime()
		t.ExpiresAt = &expiresAt
	}

	if token.CreatedAt != nil {
		t.CreatedAt = token.CreatedAt.AsTime()
	}
}
