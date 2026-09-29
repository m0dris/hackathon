package requests_service

import (
	"context"

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

func (s *Service) Create(ctx context.Context, requesterID int64, r core_domain.ServiceRequest) (core_domain.ServiceRequest, error) {
	if err := r.Validate(); err != nil {
		return core_domain.ServiceRequest{}, err
	}

	resp, err := s.db.CreateServiceRequest(ctx, &pb.CreateServiceRequestProto{
		RequesterId: requesterID,
		UserFlatsId: r.UserFlatsId,
		Type:        requestTypeToProto(r.Type),
		Title:       r.Title,
		Description: r.Description,
	})
	if err != nil {
		return core_domain.ServiceRequest{}, core_errors.FromGRPC(err)
	}

	return requestFromProto(resp), nil
}

func (s *Service) My(ctx context.Context, requesterID int64) ([]core_domain.ServiceRequest, error) {
	resp, err := s.db.GetMyRequests(ctx, &pb.GetMyServiceRequestsProto{RequesterId: requesterID})
	if err != nil {
		return nil, core_errors.FromGRPC(err)
	}

	return requestsFromProto(resp.GetRequests()), nil
}

func (s *Service) Company(ctx context.Context, companyID int64, status *core_domain.RequestStatus) ([]core_domain.ServiceRequest, error) {
	req := &pb.GetCompanyServiceRequestsProto{CompanyId: companyID}
	if status != nil {
		protoStatus := requestStatusToProto(*status)
		req.Status = &protoStatus
	}

	resp, err := s.db.GetCompanyRequests(ctx, req)
	if err != nil {
		return nil, core_errors.FromGRPC(err)
	}

	return requestsFromProto(resp.GetRequests()), nil
}

func (s *Service) UpdateStatus(ctx context.Context, companyID, requestID int64, status core_domain.RequestStatus) (core_domain.ServiceRequest, error) {
	resp, err := s.db.UpdateRequestStatus(ctx, &pb.UpdateServiceRequestStatusProto{
		CompanyId: companyID,
		RequestId: requestID,
		Status:    requestStatusToProto(status),
	})
	if err != nil {
		return core_domain.ServiceRequest{}, core_errors.FromGRPC(err)
	}

	return requestFromProto(resp), nil
}

func requestsFromProto(in []*pb.ServiceRequestResponseProto) []core_domain.ServiceRequest {
	out := make([]core_domain.ServiceRequest, 0, len(in))
	for _, r := range in {
		out = append(out, requestFromProto(r))
	}

	return out
}

func requestFromProto(r *pb.ServiceRequestResponseProto) core_domain.ServiceRequest {
	return core_domain.ServiceRequest{
		Id:          r.GetId(),
		UserFlatsId: r.GetUserFlatsId(),
		CompanyId:   r.GetCompanyId(),
		Type:        requestTypeFromProto(r.GetType()),
		Title:       r.GetTitle(),
		Description: r.GetDescription(),
		Status:      requestStatusFromProto(r.GetStatus()),
		CreatedAt:   core_grpc_pbtime.ToTime(r.GetCreatedAt()),
		UpdatedAt:   core_grpc_pbtime.ToTimePtr(r.GetUpdatedAt()),

		City:        r.GetCity(),
		Street:      r.GetStreet(),
		HouseNumber: r.GetHouseNumber(),
		FlatNumber:  int(r.GetFlatNumber()),

		ApplicantFirstName: r.GetApplicantFirstName(),
		ApplicantLastName:  r.GetApplicantLastName(),
		ApplicantPhone:     r.GetApplicantPhone(),
	}
}

var requestTypeToProtoMap = map[core_domain.RequestType]pb.RequestTypeProto{
	core_domain.RequestTypeDirtyEntrance:  pb.RequestTypeProto_TYPE_DIRTY_ENTRANCE,
	core_domain.RequestTypeElevatorBroken: pb.RequestTypeProto_TYPE_ELEVATOR_BROKEN,
	core_domain.RequestTypeNoisyNeighbors: pb.RequestTypeProto_TYPE_NOISY_NEIGHBORS,
	core_domain.RequestTypePlumbing:       pb.RequestTypeProto_TYPE_PLUMBING,
	core_domain.RequestTypeElectricity:    pb.RequestTypeProto_TYPE_ELECTRICITY,
	core_domain.RequestTypeHeating:        pb.RequestTypeProto_TYPE_HEATING,
	core_domain.RequestTypeIntercom:       pb.RequestTypeProto_TYPE_INTERCOM,
	core_domain.RequestTypeParking:        pb.RequestTypeProto_TYPE_PARKING,
	core_domain.RequestTypeOther:          pb.RequestTypeProto_TYPE_OTHER,
}

func requestTypeToProto(t core_domain.RequestType) pb.RequestTypeProto {
	return requestTypeToProtoMap[t]
}

func requestTypeFromProto(t pb.RequestTypeProto) core_domain.RequestType {
	for k, v := range requestTypeToProtoMap {
		if v == t {
			return k
		}
	}

	return core_domain.RequestTypeOther
}

func requestStatusToProto(s core_domain.RequestStatus) pb.RequestStatusProto {
	switch s {
	case core_domain.RequestStatusInProgress:
		return pb.RequestStatusProto_STATUS_IN_PROGRESS
	case core_domain.RequestStatusDone:
		return pb.RequestStatusProto_STATUS_DONE
	default:
		return pb.RequestStatusProto_STATUS_NEW
	}
}

func requestStatusFromProto(s pb.RequestStatusProto) core_domain.RequestStatus {
	switch s {
	case pb.RequestStatusProto_STATUS_IN_PROGRESS:
		return core_domain.RequestStatusInProgress
	case pb.RequestStatusProto_STATUS_DONE:
		return core_domain.RequestStatusDone
	default:
		return core_domain.RequestStatusNew
	}
}
