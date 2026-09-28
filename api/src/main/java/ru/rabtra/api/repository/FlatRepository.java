package ru.rabtra.api.repository;

import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;
import ru.rabtra.api.model.Flat;

import java.util.List;
import java.util.Optional;

@Repository
public interface FlatRepository extends JpaRepository<Flat, Long> {
    List<Flat> findByHouseId(Long houseId);

    Optional<Flat> findByHouseIdAndFlatNumber(Long houseId, Integer flatNumber);
}
