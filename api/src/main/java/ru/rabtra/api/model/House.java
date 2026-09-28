package ru.rabtra.api.model;

import jakarta.persistence.*;
import lombok.*;

@Entity
@Table(name = "house")
@Getter
@Setter
@NoArgsConstructor
@AllArgsConstructor
@Builder
public class House {

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @ManyToOne(fetch = FetchType.LAZY)
    @JoinColumn(name = "company_id")
    private User company;

    private String city;

    private String street;

    @Column(name = "house_number")
    private String houseNumber;
}