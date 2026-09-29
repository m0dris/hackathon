package users_service

import (
	"context"
	"errors"

	core_domain "github.com/m0dris/hackathon/internal/core/domain"
	core_errors "github.com/m0dris/hackathon/internal/core/errors"
	core_grpc_pbtime "github.com/m0dris/hackathon/internal/core/grpc/pbtime"
	pb "github.com/m0dris/hackathon/internal/grpc/pb"
)

type Service struct {
	db pb.UkServiceClient
}

func New(db pb.UkServiceClient) *Service {
	return &Service{db: db}
}

func (s *Service) Register(ctx context.Context, u core_domain.User) (core_domain.User, error) {
	if err := u.Validate(); err != nil {
		return core_domain.User{}, err
	}

	resp, err := s.db.CreateUser(ctx, &pb.CreateUserRequestProto{
		Id:          u.Id,
		ChatId:      u.ChatId,
		FirstName:   u.FirstName,
		LastName:    u.LastName,
		PhoneNumber: u.PhoneNumber,
		Role:        roleToProto(u.Role),
	})
	if err != nil {
		return core_domain.User{}, core_errors.FromGRPC(err)
	}

	return userFromProto(resp), nil
}

func (s *Service) Me(ctx context.Context, maxUserID int64) (core_domain.User, bool, error) {
	resp, err := s.db.GetUserMe(ctx, &pb.GetUserMeRequestProto{RequesterId: maxUserID})
	if err != nil {
		wrapped := core_errors.FromGRPC(err)
		if errors.Is(wrapped, core_errors.ErrNotFound) {
			return core_domain.User{}, false, nil
		}

		return core_domain.User{}, false, wrapped
	}

	return userFromProto(resp), true, nil
}

func roleToProto(r core_domain.Role) pb.RoleProto {
	if r == core_domain.RoleCompany {
		return pb.RoleProto_ROLE_COMPANY
	}

	return pb.RoleProto_ROLE_USER
}

func roleFromProto(r pb.RoleProto) core_domain.Role {
	if r == pb.RoleProto_ROLE_COMPANY {
		return core_domain.RoleCompany
	}

	return core_domain.RoleUser
}

func userFromProto(u *pb.UserResponseProto) core_domain.User {
	return core_domain.User{
		Id:          u.GetId(),
		ChatId:      u.GetChatId(),
		FirstName:   u.GetFirstName(),
		LastName:    u.GetLastName(),
		PhoneNumber: u.GetPhoneNumber(),
		Role:        roleFromProto(u.GetRole()),
		CreatedAt:   core_grpc_pbtime.ToTime(u.GetCreatedAt()),
	}
}
