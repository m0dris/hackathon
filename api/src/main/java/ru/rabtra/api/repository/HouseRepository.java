package ru.rabtra.api.repository;

import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;
import ru.rabtra.api.model.House;

import java.util.List;
import java.util.Optional;

@Repository
public interface HouseRepository extends JpaRepository<House, Long> {
    List<House> findByCompanyId(Long companyId);

    Optional<House> findByCityIgnoreCaseAndStreetIgnoreCaseAndHouseNumberIgnoreCase(
            String city, String street, String houseNumber);
}