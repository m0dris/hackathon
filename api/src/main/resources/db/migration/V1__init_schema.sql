-- jdbc:postgresql://localhost:5432/uk_service


CREATE TABLE users (
    id BIGINT PRIMARY KEY,
    chat_id BIGINT NOT NULL UNIQUE,
    first_name VARCHAR(255) NOT NULL,
    last_name VARCHAR(255) NOT NULL,
    phone_number VARCHAR(255) NOT NULL,
    role VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE house (
    id BIGSERIAL PRIMARY KEY,
    company_id BIGINT REFERENCES users(id) ON DELETE NO ACTION,
    city VARCHAR(255) NOT NULL,
    street VARCHAR(255) NOT NULL,
    house_number VARCHAR(255) NOT NULL,

    CONSTRAINT uq_house_address UNIQUE (city, street, house_number)
);

CREATE TABLE flats (
    id BIGSERIAL PRIMARY KEY,
    house_id BIGINT REFERENCES house(id) ON DELETE NO ACTION,
    flat_number INTEGER NOT NULL,

    CONSTRAINT uq_house_flat UNIQUE (house_id, flat_number)
);

CREATE TABLE user_flats (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    flat_id BIGINT NOT NULL REFERENCES flats(id) ON DELETE CASCADE,

    CONSTRAINT uq_user_flat UNIQUE (user_id, flat_id)
);

CREATE TABLE requests (
    id BIGSERIAL PRIMARY KEY,
    user_flats_id BIGINT NOT NULL REFERENCES user_flats(id) ON DELETE NO ACTION,
    company_id BIGINT NOT NULL REFERENCES users(id) ON DELETE NO ACTION,
    type VARCHAR(255) NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    status VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);


CREATE INDEX idx_house_company_id ON house(company_id);
CREATE INDEX idx_flats_house_id ON flats(house_id);
CREATE INDEX idx_user_flats_user_id ON user_flats(user_id);
CREATE INDEX idx_user_flats_flat_id ON user_flats(flat_id);
CREATE INDEX idx_requests_user_flats_id ON requests(user_flats_id);
CREATE INDEX idx_requests_company_id ON requests(company_id);

CREATE INDEX idx_requests_company_status ON requests(company_id, status);


CREATE OR REPLACE FUNCTION update_modified_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS update_requests_modtime ON requests;

CREATE TRIGGER update_requests_modtime
    BEFORE UPDATE ON requests
    FOR EACH ROW
    EXECUTE FUNCTION update_modified_column();
