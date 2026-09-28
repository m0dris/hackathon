package ru.rabtra.api.repository;

import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;
import ru.rabtra.api.model.Request;
import ru.rabtra.api.model.enums.RequestStatus;

import java.util.List;

@Repository
public interface RequestRepository extends JpaRepository<Request, Long> {
    List<Request> findByUserFlatUserId(Long userId);
    List<Request> findByCompanyId(Long companyId);
    List<Request> findByCompanyIdAndStatus(Long companyId, RequestStatus status);
}