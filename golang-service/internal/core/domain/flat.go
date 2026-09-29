package domain

import (
	"fmt"

	core_errors "github.com/m0dris/hackathon/internal/core/errors"
)

type Flat struct {
	Id          int64
	City        string
	Street      string
	HouseNumber string
	FlatNumber  int
}

func (f *Flat) Address() string {
	return fmt.Sprintf("%s, %s, %s, кв. %d", f.City, f.Street, f.HouseNumber, f.FlatNumber)
}

type UserFlat struct {
	UserFlatId int64
	Flat       Flat
}

func ValidateFlatNumber(flatNumber int) error {
	if flatNumber < 1 || flatNumber > 9999 {
		return fmt.Errorf("invalid `FlatNumber`: %d: %w", flatNumber, core_errors.ErrInvalidArgument)
	}

	return nil
}
