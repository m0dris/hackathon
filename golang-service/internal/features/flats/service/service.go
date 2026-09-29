package flats_service

import (
	"context"

	core_domain "github.com/m0dris/hackathon/internal/core/domain"
	core_errors "github.com/m0dris/hackathon/internal/core/errors"
	pb "github.com/m0dris/hackathon/internal/grpc/pb"
)

type Service struct {
	db pb.UkServiceClient
}

func New(db pb.UkServiceClient) *Service {
	return &Service{db: db}
}

func (s *Service) AddMy(ctx context.Context, requesterID int64, houseID int64, flatNumber int) (core_domain.UserFlat, error) {
	if err := core_domain.ValidateFlatNumber(flatNumber); err != nil {
		return core_domain.UserFlat{}, err
	}

	resp, err := s.db.CreateMyFlat(ctx, &pb.CreateMyFlatRequestProto{
		RequesterId: requesterID,
		HouseId:     houseID,
		FlatNumber:  int32(flatNumber),
	})
	if err != nil {
		return core_domain.UserFlat{}, core_errors.FromGRPC(err)
	}

	return userFlatFromProto(resp), nil
}

func (s *Service) My(ctx context.Context, requesterID int64) ([]core_domain.UserFlat, error) {
	resp, err := s.db.GetMyFlats(ctx, &pb.GetMyFlatsRequestProto{RequesterId: requesterID})
	if err != nil {
		return nil, core_errors.FromGRPC(err)
	}

	flats := make([]core_domain.UserFlat, 0, len(resp.GetMyFlats()))
	for _, f := range resp.GetMyFlats() {
		flats = append(flats, userFlatFromProto(f))
	}

	return flats, nil
}

func userFlatFromProto(uf *pb.UserFlatResponseProto) core_domain.UserFlat {
	f := uf.GetFlat()

	return core_domain.UserFlat{
		UserFlatId: uf.GetUserFlatId(),
		Flat: core_domain.Flat{
			Id:          f.GetId(),
			City:        f.GetCity(),
			Street:      f.GetStreet(),
			HouseNumber: f.GetHouseNumber(),
			FlatNumber:  int(f.GetFlatNumber()),
		},
	}
}
