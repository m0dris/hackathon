package ru.rabtra.api.model;

import jakarta.persistence.*;
import lombok.*;

@Entity
@Table(name = "user_flats")
@Getter
@Setter
@NoArgsConstructor
@AllArgsConstructor
@Builder
public class UserFlat {

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @ManyToOne(fetch = FetchType.LAZY)
    @JoinColumn(name = "user_id")
    private User user;

    @ManyToOne(fetch = FetchType.LAZY)
    @JoinColumn(name = "flat_id")
    private Flat flat;
}
