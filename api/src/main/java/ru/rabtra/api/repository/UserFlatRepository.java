package ru.rabtra.api.repository;

import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;
import ru.rabtra.api.model.UserFlat;

import java.util.List;

@Repository
public interface UserFlatRepository extends JpaRepository<UserFlat, Long> {
    List<UserFlat> findByUserId(Long userId);
}