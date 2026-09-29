package domain

import (
	"fmt"
	"time"

	core_errors "github.com/m0dris/hackathon/internal/core/errors"
)

type Role string

const (
	RoleUser    Role = "USER"
	RoleCompany Role = "COMPANY"
)

func (r Role) Valid() bool {
	return r == RoleUser || r == RoleCompany
}

type User struct {
	Id          int64
	ChatId      int64
	FirstName   string
	LastName    string
	PhoneNumber string
	Role        Role
	CreatedAt   time.Time
}

func (u *User) IsCompany() bool {
	return u.Role == RoleCompany
}

func (u *User) Validate() error {
	if !u.Role.Valid() {
		return fmt.Errorf("invalid `Role`: %s: %w", u.Role, core_errors.ErrInvalidArgument)
	}

	if l := len([]rune(u.FirstName)); l < 1 || l > 100 {
		return fmt.Errorf("invalid `FirstName` len: %d: %w", l, core_errors.ErrInvalidArgument)
	}

	if l := len([]rune(u.LastName)); l < 1 || l > 100 {
		return fmt.Errorf("invalid `LastName` len: %d: %w", l, core_errors.ErrInvalidArgument)
	}

	if l := len([]rune(u.PhoneNumber)); l < 5 || l > 20 {
		return fmt.Errorf("invalid `PhoneNumber` len: %d: %w", l, core_errors.ErrInvalidArgument)
	}

	return nil
}
