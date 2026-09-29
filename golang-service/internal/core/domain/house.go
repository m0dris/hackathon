package domain

import (
	"fmt"

	core_errors "github.com/m0dris/hackathon/internal/core/errors"
)

type House struct {
	Id          int64
	CompanyId   int64
	City        string
	Street      string
	HouseNumber string
}

func (h *House) Address() string {
	return fmt.Sprintf("%s, %s, %s", h.City, h.Street, h.HouseNumber)
}

func (h *House) Validate() error {
	if l := len([]rune(h.City)); l < 2 || l > 100 {
		return fmt.Errorf("invalid `City` len: %d: %w", l, core_errors.ErrInvalidArgument)
	}

	if l := len([]rune(h.Street)); l < 2 || l > 100 {
		return fmt.Errorf("invalid `Street` len: %d: %w", l, core_errors.ErrInvalidArgument)
	}

	if l := len([]rune(h.HouseNumber)); l < 1 || l > 20 {
		return fmt.Errorf("invalid `HouseNumber` len: %d: %w", l, core_errors.ErrInvalidArgument)
	}

	return nil
}
