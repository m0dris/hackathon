package ru.rabtra.api.grpc;

import io.grpc.Status;
import io.grpc.stub.StreamObserver;
import lombok.RequiredArgsConstructor;
import net.devh.boot.grpc.server.service.GrpcService;
import org.springframework.transaction.annotation.Transactional;
import ru.rabtra.api.model.*;
import ru.rabtra.api.model.enums.RequestStatus;
import ru.rabtra.api.repository.*;
import ru.rabtra.api.utils.ProtoMapperUtils;

import java.util.List;

@GrpcService
@RequiredArgsConstructor
public class UkGrpcServiceImpl extends UkServiceGrpc.UkServiceImplBase {

    private final UserRepository userRepository;
    private final HouseRepository houseRepository;
    private final FlatRepository flatRepository;
    private final UserFlatRepository userFlatRepository;
    private final RequestRepository requestRepository;

    @Override
    @Transactional
    public void createUser(CreateUserRequestProto request, StreamObserver<UserResponseProto> responseObserver) {
        User user = User.builder()
                .id(request.getId())
                .chatId(request.getChatId())
                .firstName(request.getFirstName())
                .lastName(request.getLastName())
                .phoneNumber(request.getPhoneNumber())
                .role(ProtoMapperUtils.toJavaRole(request.getRole()))
                .build();

        user = userRepository.saveAndFlush(user);

        UserResponseProto response = buildUserResponse(user);
        responseObserver.onNext(response);
        responseObserver.onCompleted();
    }

    @Override
    @Transactional(readOnly = true)
    public void getUserMe(GetUserMeRequestProto request, StreamObserver<UserResponseProto> responseObserver) {
        User user = userRepository.findById(request.getRequesterId())
                .orElseThrow(() -> Status.NOT_FOUND.withDescription("User not found").asRuntimeException());

        responseObserver.onNext(buildUserResponse(user));
        responseObserver.onCompleted();
    }

    @Override
    @Transactional
    public void createHouse(CreateHouseRequestProto request, StreamObserver<HouseResponseProto> responseObserver) {
        User company = userRepository.findById(request.getCompanyId())
                .orElseThrow(() -> Status.NOT_FOUND.withDescription("Company (User) not found").asRuntimeException());

        House house = House.builder()
                .company(company)
                .city(request.getCity())
                .street(request.getStreet())
                .houseNumber(request.getHouseNumber())
                .build();

        house = houseRepository.save(house);

        responseObserver.onNext(buildHouseResponse(house));
        responseObserver.onCompleted();
    }

    @Override
    @Transactional(readOnly = true)
    public void getMyHouses(GetMyHousesRequestProto request, StreamObserver<HouseListResponseProto> responseObserver) {
        List<House> houses = houseRepository.findByCompanyId(request.getCompanyId());

        HouseListResponseProto.Builder listBuilder = HouseListResponseProto.newBuilder();
        houses.forEach(h -> listBuilder.addHouses(buildHouseResponse(h)));

        responseObserver.onNext(listBuilder.build());
        responseObserver.onCompleted();
    }

    @Override
    @Transactional(readOnly = true)
    public void searchHouseByAddress(SearchHouseByAddressRequestProto request, StreamObserver<HouseResponseProto> responseObserver) {
        House house = houseRepository.findByCityIgnoreCaseAndStreetIgnoreCaseAndHouseNumberIgnoreCase(
                request.getCity(), request.getStreet(), request.getHouseNumber()
        ).orElseThrow(() -> Status.NOT_FOUND.withDescription("Дом по указанному адресу не найден").asRuntimeException());

        responseObserver.onNext(buildHouseResponse(house));
        responseObserver.onCompleted();
    }

    @Override
    @Transactional(readOnly = true)
    public void getFlatsByHouse(GetFlatsRequestProto request, StreamObserver<FlatListResponseProto> responseObserver) {
        List<Flat> flats = flatRepository.findByHouseId(request.getHouseId());

        FlatListResponseProto.Builder listBuilder = FlatListResponseProto.newBuilder();
        flats.forEach(f -> listBuilder.addFlats(buildFlatResponse(f)));

        responseObserver.onNext(listBuilder.build());
        responseObserver.onCompleted();
    }

    @Override
    @Transactional
    public void createMyFlat(CreateMyFlatRequestProto request, StreamObserver<UserFlatResponseProto> responseObserver) {
        User user = userRepository.findById(request.getRequesterId())
                .orElseThrow(() -> Status.NOT_FOUND.withDescription("User not found").asRuntimeException());

        House house = houseRepository.findById(request.getHouseId())
                .orElseThrow(() -> Status.NOT_FOUND.withDescription("House not found").asRuntimeException());

        Flat flat = flatRepository.findByHouseIdAndFlatNumber(house.getId(), request.getFlatNumber())
                .orElseGet(() -> {
                    Flat newFlat = Flat.builder().house(house).flatNumber(request.getFlatNumber()).build();
                    return flatRepository.save(newFlat);
                });

        UserFlat userFlat = UserFlat.builder()
                .user(user)
                .flat(flat)
                .build();
        userFlat = userFlatRepository.save(userFlat);

        UserFlatResponseProto response = UserFlatResponseProto.newBuilder()
                .setUserFlatId(userFlat.getId())
                .setFlat(buildFlatResponse(flat))
                .build();

        responseObserver.onNext(response);
        responseObserver.onCompleted();
    }

    @Override
    @Transactional(readOnly = true)
    public void getMyFlats(GetMyFlatsRequestProto request, StreamObserver<UserFlatListResponseProto> responseObserver) {
        List<UserFlat> userFlats = userFlatRepository.findByUserId(request.getRequesterId());

        UserFlatListResponseProto.Builder listBuilder = UserFlatListResponseProto.newBuilder();
        userFlats.forEach(uf -> listBuilder.addMyFlats(
                UserFlatResponseProto.newBuilder()
                        .setUserFlatId(uf.getId())
                        .setFlat(buildFlatResponse(uf.getFlat()))
                        .build()
        ));

        responseObserver.onNext(listBuilder.build());
        responseObserver.onCompleted();
    }

    @Override
    @Transactional
    public void createServiceRequest(CreateServiceRequestProto request, StreamObserver<ServiceRequestResponseProto> responseObserver) {
        UserFlat userFlat = userFlatRepository.findById(request.getUserFlatsId())
                .orElseThrow(() -> Status.NOT_FOUND.withDescription("UserFlat not found").asRuntimeException());

        User company = userFlat.getFlat().getHouse().getCompany();

        Request serviceReq = Request.builder()
                .userFlat(userFlat)
                .company(company)
                .type(ProtoMapperUtils.toJavaType(request.getType()))
                .title(request.getTitle())
                .description(request.getDescription())
                .status(RequestStatus.NEW)
                .build();

        serviceReq = requestRepository.save(serviceReq);

        responseObserver.onNext(buildServiceRequestResponse(serviceReq));
        responseObserver.onCompleted();
    }

    @Override
    @Transactional(readOnly = true)
    public void getMyRequests(GetMyServiceRequestsProto request, StreamObserver<ServiceRequestListResponseProto> responseObserver) {
        List<Request> requests = requestRepository.findByUserFlatUserId(request.getRequesterId());

        ServiceRequestListResponseProto.Builder listBuilder = ServiceRequestListResponseProto.newBuilder();
        requests.forEach(r -> listBuilder.addRequests(buildServiceRequestResponse(r)));

        responseObserver.onNext(listBuilder.build());
        responseObserver.onCompleted();
    }

    @Override
    @Transactional(readOnly = true)
    public void getCompanyRequests(GetCompanyServiceRequestsProto request, StreamObserver<ServiceRequestListResponseProto> responseObserver) {
        List<Request> requests;

        if (request.hasStatus()) {
            RequestStatus status = ProtoMapperUtils.toJavaStatus(request.getStatus());
            requests = requestRepository.findByCompanyIdAndStatus(request.getCompanyId(), status);
        } else {
            requests = requestRepository.findByCompanyId(request.getCompanyId());
        }

        ServiceRequestListResponseProto.Builder listBuilder = ServiceRequestListResponseProto.newBuilder();
        requests.forEach(r -> listBuilder.addRequests(buildServiceRequestResponse(r)));

        responseObserver.onNext(listBuilder.build());
        responseObserver.onCompleted();
    }

    @Override
    @Transactional
    public void updateRequestStatus(UpdateServiceRequestStatusProto request, StreamObserver<ServiceRequestResponseProto> responseObserver) {
        Request serviceReq = requestRepository.findById(request.getRequestId())
                .orElseThrow(() -> Status.NOT_FOUND.withDescription("Request not found").asRuntimeException());

        if (!serviceReq.getCompany().getId().equals(request.getCompanyId())) {
            throw Status.PERMISSION_DENIED.withDescription("Not your request").asRuntimeException();
        }

        serviceReq.setStatus(ProtoMapperUtils.toJavaStatus(request.getStatus()));
        serviceReq = requestRepository.save(serviceReq);

        responseObserver.onNext(buildServiceRequestResponse(serviceReq));
        responseObserver.onCompleted();
    }


    private UserResponseProto buildUserResponse(User user) {
        return UserResponseProto.newBuilder()
                .setId(user.getId())
                .setChatId(user.getChatId())
                .setFirstName(user.getFirstName())
                .setLastName(user.getLastName())
                .setPhoneNumber(user.getPhoneNumber())
                .setRole(ProtoMapperUtils.toProtoRole(user.getRole()))
                .setCreatedAt(ProtoMapperUtils.toProtoTimestamp(user.getCreatedAt()))
                .build();
    }

    private HouseResponseProto buildHouseResponse(House house) {
        return HouseResponseProto.newBuilder()
                .setId(house.getId())
                .setCompanyId(house.getCompany().getId())
                .setCity(house.getCity())
                .setStreet(house.getStreet())
                .setHouseNumber(house.getHouseNumber())
                .build();
    }

    private FlatResponseProto buildFlatResponse(Flat flat) {
        House house = flat.getHouse();
        return FlatResponseProto.newBuilder()
                .setId(flat.getId())
                .setCity(house.getCity())
                .setStreet(house.getStreet())
                .setHouseNumber(house.getHouseNumber())
                .setFlatNumber(flat.getFlatNumber())
                .build();
    }

    private ServiceRequestResponseProto buildServiceRequestResponse(Request req) {
        UserFlat uf = req.getUserFlat();
        Flat f = uf.getFlat();
        House h = f.getHouse();
        User applicant = uf.getUser();

        return ServiceRequestResponseProto.newBuilder()
                .setId(req.getId())
                .setUserFlatsId(uf.getId())
                .setCompanyId(req.getCompany().getId())
                .setType(ProtoMapperUtils.toProtoType(req.getType()))
                .setTitle(req.getTitle())
                .setDescription(req.getDescription())
                .setStatus(ProtoMapperUtils.toProtoStatus(req.getStatus()))
                .setCreatedAt(ProtoMapperUtils.toProtoTimestamp(req.getCreatedAt()))
                .setUpdatedAt(ProtoMapperUtils.toProtoTimestamp(req.getUpdatedAt()))
                .setCity(h.getCity())
                .setStreet(h.getStreet())
                .setHouseNumber(h.getHouseNumber())
                .setFlatNumber(f.getFlatNumber())
                .setApplicantFirstName(applicant.getFirstName())
                .setApplicantLastName(applicant.getLastName())
                .setApplicantPhone(applicant.getPhoneNumber())
                .build();
    }
}