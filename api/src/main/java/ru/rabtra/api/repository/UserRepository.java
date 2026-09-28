package ru.rabtra.api.repository;

import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;
import ru.rabtra.api.model.User;

@Repository
public interface UserRepository extends JpaRepository<User, Long> {
}