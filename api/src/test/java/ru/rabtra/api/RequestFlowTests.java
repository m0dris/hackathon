package ru.rabtra.api;

import io.grpc.Status;
import io.grpc.StatusRuntimeException;
import io.grpc.stub.StreamObserver;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;
import ru.rabtra.api.grpc.*;

import java.util.concurrent.atomic.AtomicReference;
import java.util.function.Consumer;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;

@SpringBootTest
class RequestFlowTests {
    @Autowired
    private UkGrpcServiceImpl service;

    @Test
    void residentAndCompanyCompleteRequestWithoutAccessToAnotherFlat() {
        createUser(101, RoleProto.ROLE_COMPANY);
        createUser(102, RoleProto.ROLE_USER);

        HouseResponseProto house = call(observer -> service.createHouse(CreateHouseRequestProto.newBuilder()
                .setCompanyId(101).setCity("Тестовый город").setStreet("Учебная").setHouseNumber("1")
                .build(), observer));
        UserFlatResponseProto flat = call(observer -> service.createMyFlat(CreateMyFlatRequestProto.newBuilder()
                .setRequesterId(102).setHouseId(house.getId()).setFlatNumber(12).build(), observer));

        CreateServiceRequestProto input = CreateServiceRequestProto.newBuilder()
                .setRequesterId(102).setUserFlatsId(flat.getUserFlatId()).setType(RequestTypeProto.TYPE_DIRTY_ENTRANCE)
                .setTitle("Уборка подъезда").setDescription("Тестовое описание проблемы").build();
        ServiceRequestResponseProto request = call(observer -> service.createServiceRequest(input, observer));
        assertThat(request.getStatus()).isEqualTo(RequestStatusProto.STATUS_NEW);
        assertThat(request.getCompanyId()).isEqualTo(101);

        assertThatThrownBy(() -> service.createServiceRequest(input.toBuilder().setRequesterId(999).build(), observer()))
                .isInstanceOfSatisfying(StatusRuntimeException.class,
                        e -> assertThat(e.getStatus().getCode()).isEqualTo(Status.Code.PERMISSION_DENIED));
        assertThatThrownBy(() -> service.updateRequestStatus(UpdateServiceRequestStatusProto.newBuilder()
                        .setCompanyId(999).setRequestId(request.getId()).setStatus(RequestStatusProto.STATUS_IN_PROGRESS)
                        .build(), observer()))
                .isInstanceOfSatisfying(StatusRuntimeException.class,
                        e -> assertThat(e.getStatus().getCode()).isEqualTo(Status.Code.PERMISSION_DENIED));

        advance(request.getId(), RequestStatusProto.STATUS_IN_PROGRESS);
        advance(request.getId(), RequestStatusProto.STATUS_DONE);
        assertThatThrownBy(() -> advance(request.getId(), RequestStatusProto.STATUS_IN_PROGRESS))
                .isInstanceOfSatisfying(StatusRuntimeException.class,
                        e -> assertThat(e.getStatus().getCode()).isEqualTo(Status.Code.FAILED_PRECONDITION));

        ServiceRequestListResponseProto mine = call(observer -> service.getMyRequests(
                GetMyServiceRequestsProto.newBuilder().setRequesterId(102).build(), observer));
        assertThat(mine.getRequestsList()).hasSize(1);
        assertThat(mine.getRequests(0).getStatus()).isEqualTo(RequestStatusProto.STATUS_DONE);
        ServiceRequestListResponseProto filtered = call(observer -> service.getCompanyRequests(
                GetCompanyServiceRequestsProto.newBuilder().setCompanyId(101)
                        .setStatus(RequestStatusProto.STATUS_DONE).build(), observer));
        assertThat(filtered.getRequestsList()).hasSize(1);
    }

    private void createUser(long id, RoleProto role) {
        UserResponseProto ignored = call(observer -> service.createUser(CreateUserRequestProto.newBuilder().setId(id).setChatId(id + 1000)
                .setFirstName("Тест").setLastName("Пользователь").setPhoneNumber("+70000000001")
                .setRole(role).build(), observer));
    }

    private ServiceRequestResponseProto advance(long id, RequestStatusProto status) {
        return call(observer -> service.updateRequestStatus(UpdateServiceRequestStatusProto.newBuilder()
                .setCompanyId(101).setRequestId(id).setStatus(status).build(), observer));
    }

    private static <T> T call(Consumer<StreamObserver<T>> invocation) {
        AtomicReference<T> value = new AtomicReference<>();
        invocation.accept(new StreamObserver<>() {
            public void onNext(T item) { value.set(item); }
            public void onError(Throwable error) { throw new AssertionError(error); }
            public void onCompleted() { }
        });
        return value.get();
    }

    private static <T> StreamObserver<T> observer() {
        return new StreamObserver<>() {
            public void onNext(T item) { }
            public void onError(Throwable error) { throw new AssertionError(error); }
            public void onCompleted() { }
        };
    }
}
