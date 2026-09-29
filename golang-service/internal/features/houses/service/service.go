package houses_service

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

func (s *Service) Create(ctx context.Context, companyID int64, h core_domain.House) (core_domain.House, error) {
	h.CompanyId = companyID
	if err := h.Validate(); err != nil {
		return core_domain.House{}, err
	}

	resp, err := s.db.CreateHouse(ctx, &pb.CreateHouseRequestProto{
		CompanyId:   companyID,
		City:        h.City,
		Street:      h.Street,
		HouseNumber: h.HouseNumber,
	})
	if err != nil {
		return core_domain.House{}, core_errors.FromGRPC(err)
	}

	return houseFromProto(resp), nil
}

func (s *Service) My(ctx context.Context, companyID int64) ([]core_domain.House, error) {
	resp, err := s.db.GetMyHouses(ctx, &pb.GetMyHousesRequestProto{CompanyId: companyID})
	if err != nil {
		return nil, core_errors.FromGRPC(err)
	}

	houses := make([]core_domain.House, 0, len(resp.GetHouses()))
	for _, h := range resp.GetHouses() {
		houses = append(houses, houseFromProto(h))
	}

	return houses, nil
}

func (s *Service) SearchByAddress(ctx context.Context, city, street, houseNumber string) (core_domain.House, error) {
	resp, err := s.db.SearchHouseByAddress(ctx, &pb.SearchHouseByAddressRequestProto{
		City:        city,
		Street:      street,
		HouseNumber: houseNumber,
	})
	if err != nil {
		return core_domain.House{}, core_errors.FromGRPC(err)
	}

	return houseFromProto(resp), nil
}

func (s *Service) FlatsByHouse(ctx context.Context, houseID int64) ([]core_domain.Flat, error) {
	resp, err := s.db.GetFlatsByHouse(ctx, &pb.GetFlatsRequestProto{HouseId: houseID})
	if err != nil {
		return nil, core_errors.FromGRPC(err)
	}

	flats := make([]core_domain.Flat, 0, len(resp.GetFlats()))
	for _, f := range resp.GetFlats() {
		flats = append(flats, flatFromProto(f))
	}

	return flats, nil
}

func houseFromProto(h *pb.HouseResponseProto) core_domain.House {
	return core_domain.House{
		Id:          h.GetId(),
		CompanyId:   h.GetCompanyId(),
		City:        h.GetCity(),
		Street:      h.GetStreet(),
		HouseNumber: h.GetHouseNumber(),
	}
}

func flatFromProto(f *pb.FlatResponseProto) core_domain.Flat {
	return core_domain.Flat{
		Id:          f.GetId(),
		City:        f.GetCity(),
		Street:      f.GetStreet(),
		HouseNumber: f.GetHouseNumber(),
		FlatNumber:  int(f.GetFlatNumber()),
	}
}
