package domain

import (
	"fmt"
	"time"

	core_errors "github.com/m0dris/hackathon/internal/core/errors"
)

type RequestType string

const (
	RequestTypeDirtyEntrance  RequestType = "DIRTY_ENTRANCE"
	RequestTypeElevatorBroken RequestType = "ELEVATOR_BROKEN"
	RequestTypeNoisyNeighbors RequestType = "NOISY_NEIGHBORS"
	RequestTypePlumbing       RequestType = "PLUMBING"
	RequestTypeElectricity    RequestType = "ELECTRICITY"
	RequestTypeHeating        RequestType = "HEATING"
	RequestTypeIntercom       RequestType = "INTERCOM"
	RequestTypeParking        RequestType = "PARKING"
	RequestTypeOther          RequestType = "OTHER"
)

var RequestTypeLabels = []struct {
	Type  RequestType
	Label string
}{
	{RequestTypeDirtyEntrance, "🧹 Грязный подъезд"},
	{RequestTypeElevatorBroken, "🛗 Лифт не работает"},
	{RequestTypeNoisyNeighbors, "🔊 Шумные соседи"},
	{RequestTypePlumbing, "🚰 Сантехника"},
	{RequestTypeElectricity, "⚡ Электричество"},
	{RequestTypeHeating, "🔥 Отопление"},
	{RequestTypeIntercom, "🔔 Домофон"},
	{RequestTypeParking, "🚗 Парковка"},
	{RequestTypeOther, "❓ Другое"},
}

func (t RequestType) Valid() bool {
	for _, l := range RequestTypeLabels {
		if l.Type == t {
			return true
		}
	}

	return false
}

func (t RequestType) Label() string {
	for _, l := range RequestTypeLabels {
		if l.Type == t {
			return l.Label
		}
	}

	return string(t)
}

type RequestStatus string

const (
	RequestStatusNew        RequestStatus = "NEW"
	RequestStatusInProgress RequestStatus = "IN_PROGRESS"
	RequestStatusDone       RequestStatus = "DONE"
)

func (s RequestStatus) Label() string {
	switch s {
	case RequestStatusNew:
		return "🆕 Новая"
	case RequestStatusInProgress:
		return "🔧 В работе"
	case RequestStatusDone:
		return "✅ Выполнена"
	default:
		return string(s)
	}
}

func (s RequestStatus) NextStatus() RequestStatus {
	switch s {
	case RequestStatusNew:
		return RequestStatusInProgress
	case RequestStatusInProgress:
		return RequestStatusDone
	default:
		return ""
	}
}

func (s RequestStatus) NextActionLabel() string {
	switch s {
	case RequestStatusNew:
		return "▶️ Взять в работу"
	case RequestStatusInProgress:
		return "✅ Завершить"
	default:
		return ""
	}
}

type ServiceRequest struct {
	Id          int64
	UserFlatsId int64
	CompanyId   int64
	Type        RequestType
	Title       string
	Description string
	Status      RequestStatus
	CreatedAt   time.Time
	UpdatedAt   *time.Time

	City        string
	Street      string
	HouseNumber string
	FlatNumber  int

	ApplicantFirstName string
	ApplicantLastName  string
	ApplicantPhone     string
}

func (r *ServiceRequest) Address() string {
	return fmt.Sprintf("%s, %s, %s, кв. %d", r.City, r.Street, r.HouseNumber, r.FlatNumber)
}

func (r *ServiceRequest) ApplicantName() string {
	return fmt.Sprintf("%s %s", r.ApplicantFirstName, r.ApplicantLastName)
}

func (r *ServiceRequest) Validate() error {
	if !r.Type.Valid() {
		return fmt.Errorf("invalid `Type`: %s: %w", r.Type, core_errors.ErrInvalidArgument)
	}

	if l := len([]rune(r.Title)); l < 3 || l > 255 {
		return fmt.Errorf("invalid `Title` len: %d: %w", l, core_errors.ErrInvalidArgument)
	}

	if l := len([]rune(r.Description)); l < 3 || l > 2000 {
		return fmt.Errorf("invalid `Description` len: %d: %w", l, core_errors.ErrInvalidArgument)
	}

	return nil
}
