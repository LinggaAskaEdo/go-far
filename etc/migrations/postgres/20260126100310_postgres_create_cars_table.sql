-- +goose Up
-- +goose StatementBegin
CREATE TABLE cars (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id UUID NOT NULL,
  brand VARCHAR(100) NOT NULL,
  model VARCHAR(100) NOT NULL,
  year INTEGER NOT NULL CHECK (year >= 1900 AND year <= 2100),
  color VARCHAR(50),
  license_plate VARCHAR(20) UNIQUE NOT NULL,
  is_available BOOLEAN DEFAULT true,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT fk_user FOREIGN KEY (user_id) 
    REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_cars_user_id ON cars(user_id);
CREATE INDEX idx_cars_brand ON cars(brand);
CREATE INDEX idx_cars_license_plate ON cars(license_plate);
CREATE INDEX idx_cars_is_available ON cars(is_available);

CREATE TRIGGER update_cars_updated_at 
BEFORE UPDATE ON cars
FOR EACH ROW 
EXECUTE FUNCTION update_updated_at_column();

-- 50 car limit per user
CREATE OR REPLACE FUNCTION check_user_car_limit()
RETURNS TRIGGER AS $$
DECLARE
    car_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO car_count 
    FROM cars 
    WHERE user_id = NEW.user_id;
    
    IF car_count >= 50 THEN
        RAISE EXCEPTION 'User cannot have more than 50 cars';
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE 'plpgsql';

CREATE TRIGGER check_car_limit
BEFORE INSERT ON cars
FOR EACH ROW
EXECUTE FUNCTION check_user_car_limit();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS cars CASCADE;
-- +goose StatementEnd
