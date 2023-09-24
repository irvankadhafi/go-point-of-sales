package model

import (
	"context"
	"errors"
	"github.com/irvankadhafi/go-point-of-sales/rbac"
	"gorm.io/gorm"
	"time"
)

// ErrPasswordMismatch error
var ErrPasswordMismatch = errors.New("password mismatch")

type UserRepository interface {
	Create(ctx context.Context, userID int64, user *User) error
	Update(ctx context.Context, userID int64, user *User) (*User, error)
	UpdatePasswordByID(ctx context.Context, userID int64, password string) error
	FindByID(ctx context.Context, id int64) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)

	IsLoginByEmailPasswordLocked(ctx context.Context, email string) (bool, error)
	IncrementLoginByEmailPasswordRetryAttempts(ctx context.Context, email string) error
	FindPasswordByID(ctx context.Context, id int64) ([]byte, error)
}

type UserUsecase interface {
	FindByID(ctx context.Context, requester *User, id int64) (*User, error)
	Create(ctx context.Context, requester *User, input CreateUserInput) (*User, error)
	ChangePassword(ctx context.Context, requester *User, input ChangePasswordInput) (*User, error)
	UpdateProfile(ctx context.Context, requester *User, input UpdateProfileInput) (*User, error)
}

// User :nodoc:
type User struct {
	ID        int64          `json:"id"`
	Name      string         `json:"name"`
	Email     string         `json:"email"`
	Password  string         `json:"password" gorm:"->:false;<-"` // gorm create & update only (disabled read from db)
	Role      rbac.Role      `json:"role"`
	Status    UserStatus     `json:"status"`
	CreatedBy int64          `json:"created_by" gorm:"->;<-:create"` // create & read only
	UpdatedBy int64          `json:"updated_by"`
	CreatedAt time.Time      `json:"created_at" gorm:"->;<-:create"` // create & read only
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at"`

	SessionID int64                `json:"session_id" gorm:"-"`
	rolePerm  *rbac.RolePermission `gorm:"-"`
}

// UserStatus user status
type UserStatus string

// Status constants
const (
	StatusPending  UserStatus = "PENDING"
	StatusActive   UserStatus = "ACTIVE"
	StatusInactive UserStatus = "INACTIVE"
)

// SetPermission set permission to user
func (u *User) SetPermission(perm *rbac.Permission) {
	if perm == nil {
		return
	}
	u.rolePerm = rbac.NewRolePermission(u.Role, perm)
}

// SetRolePermission set role permission to user
func (u *User) SetRolePermission(rolePerm *rbac.RolePermission) {
	u.rolePerm = rolePerm
}

// GetRolePermission get the role permission
func (u *User) GetRolePermission() *rbac.RolePermission {
	return u.rolePerm
}

// HasAccess check authorization
func (u *User) HasAccess(resource rbac.Resource, action rbac.Action) bool {
	if u.rolePerm == nil {
		return false
	}

	if u.Role == rbac.RoleInternalService {
		u.Role = rbac.RoleAdmin
	}

	return u.rolePerm.HasAccess(resource, action)
}

// IsAdmin check if the user ADMIN
func (u *User) IsAdmin() bool {
	return u.Role == rbac.RoleAdmin
}

type CreateUserInput struct {
	Name                 string `json:"name" validate:"required"`
	Email                string `json:"email" validate:"required,email"`
	Password             string `json:"password" validate:"required,min=6"`
	PasswordConfirmation string `json:"password_confirmation" validate:"required,min=6,eqfield=Password"`
}

// ValidateAndFormat validate and format the phone number
func (c *CreateUserInput) ValidateAndFormat() error {
	if err := validate.Struct(c); err != nil {
		return err
	}

	return nil
}

type UpdateProfileInput struct {
	Name string `json:"name" validate:"required"`
}

func (u *UpdateProfileInput) ValidateAndFormat() error {
	if err := validate.Struct(u); err != nil {
		return err
	}

	return nil
}

// ChangePasswordInput change user password
type ChangePasswordInput struct {
	OldPassword             string `json:"old_password" validate:"required,min=6"`
	NewPassword             string `json:"new_password" validate:"required,min=6"`
	NewPasswordConfirmation string `json:"new_password_confirmation" validate:"required,min=6,eqfield=NewPassword"`
}

// Validate validate user's password & input body
func (c *ChangePasswordInput) Validate() error {
	if c.NewPassword != c.NewPasswordConfirmation {
		return ErrPasswordMismatch
	}

	return validate.Struct(c)
}
