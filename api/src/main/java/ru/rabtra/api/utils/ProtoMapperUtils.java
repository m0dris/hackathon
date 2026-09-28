package ru.rabtra.api.utils;

import com.google.protobuf.Timestamp;
import ru.rabtra.api.grpc.RequestStatusProto;
import ru.rabtra.api.grpc.RequestTypeProto;
import ru.rabtra.api.grpc.RoleProto;
import ru.rabtra.api.model.enums.RequestStatus;
import ru.rabtra.api.model.enums.RequestType;
import ru.rabtra.api.model.enums.Role;

import java.time.Instant;
import java.time.LocalDateTime;
import java.time.ZoneOffset;

public class ProtoMapperUtils {

    public static Timestamp toProtoTimestamp(LocalDateTime localDateTime) {
        if (localDateTime == null) return Timestamp.getDefaultInstance();
        Instant instant = localDateTime.toInstant(ZoneOffset.UTC);
        return Timestamp.newBuilder()
                .setSeconds(instant.getEpochSecond())
                .setNanos(instant.getNano())
                .build();
    }

    public static Role toJavaRole(RoleProto proto) {
        if (proto == RoleProto.ROLE_UNSPECIFIED || proto == RoleProto.UNRECOGNIZED) return null;
        return Role.valueOf(proto.name().replace("ROLE_", ""));
    }

    public static RoleProto toProtoRole(Role role) {
        if (role == null) return RoleProto.ROLE_UNSPECIFIED;
        return RoleProto.valueOf("ROLE_" + role.name());
    }

    public static RequestType toJavaType(RequestTypeProto proto) {
        if (proto == RequestTypeProto.TYPE_UNSPECIFIED || proto == RequestTypeProto.UNRECOGNIZED) return null;
        return RequestType.valueOf(proto.name().replace("TYPE_", ""));
    }

    public static RequestTypeProto toProtoType(RequestType type) {
        if (type == null) return RequestTypeProto.TYPE_UNSPECIFIED;
        return RequestTypeProto.valueOf("TYPE_" + type.name());
    }

    public static RequestStatus toJavaStatus(RequestStatusProto proto) {
        if (proto == RequestStatusProto.STATUS_UNSPECIFIED || proto == RequestStatusProto.UNRECOGNIZED) return null;
        return RequestStatus.valueOf(proto.name().replace("STATUS_", ""));
    }

    public static RequestStatusProto toProtoStatus(RequestStatus status) {
        if (status == null) return RequestStatusProto.STATUS_UNSPECIFIED;
        return RequestStatusProto.valueOf("STATUS_" + status.name());
    }
}