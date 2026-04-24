-- name: CreateCar
INSERT INTO cars (brand, model, year, color, license_plate, is_available)
VALUES ({{ arg .Brand }}, {{ arg .Model }}, {{ arg .Year }}, {{ arg .Color }}, {{ arg .LicensePlate }}, {{ arg .IsAvailable }})
RETURNING id, brand, model, year, color, license_plate, is_available, created_at, updated_at;

-- name: CreateCarBulk
INSERT INTO cars (brand, model, year, color, license_plate, is_available, created_at, updated_at)
VALUES
{{ range $i, $car := . }}
  {{ if $i }},{{ end }} ({{ arg $car.Brand }}, {{ arg $car.Model }}, {{ arg $car.Year }}, {{ arg $car.Color }}, {{ arg $car.LicensePlate }}, {{ arg $car.IsAvailable }}, {{ arg $car.CreatedAt }}, {{ arg $car.UpdatedAt }})
{{ end }};

-- name: AssignCarToUser
INSERT INTO users_cars (user_id, car_id)
VALUES ({{ arg .UserID }}, {{ arg .CarID }});

-- name: AssignCarToUserBulk
INSERT INTO users_cars (user_id, car_id)
VALUES
{{ range $i, $uc := . }}
  {{ if $i }},{{ end }} ({{ arg $uc.UserID }}, {{ arg $uc.CarID }})
{{ end }};

-- name: FindCarByID
SELECT id, brand, model, year, color, license_plate, is_available, created_at, updated_at
FROM cars
WHERE id = {{ arg .ID }};

-- name: FindCarByIDWithOwner
SELECT
    c.id,
    c.brand,
    c.model,
    c.year,
    c.color,
    c.license_plate,
    c.is_available,
    c.created_at,
    c.updated_at,
    u.name AS owner_name,
    u.email AS owner_email
FROM cars c
INNER JOIN users_cars uc ON c.id = uc.car_id
INNER JOIN users u ON uc.user_id = u.id
WHERE c.id = {{ arg .ID }};

-- name: FindCarsByUserID
SELECT c.id, c.brand, c.model, c.year, c.color, c.license_plate, c.is_available, c.created_at, c.updated_at
FROM cars c
INNER JOIN users_cars uc ON c.id = uc.car_id
WHERE uc.user_id = {{ arg .UserID }}
ORDER BY c.created_at DESC;

-- name: CountCarsByUserID
SELECT COUNT(*)
FROM cars c
INNER JOIN users_cars uc ON c.id = uc.car_id
WHERE uc.user_id = {{ arg .UserID }};

-- name: UpdateCar
UPDATE cars
SET brand = {{ arg .Brand }}, model = {{ arg .Model }}, year = {{ arg .Year }}, color = {{ arg .Color }}, license_plate = {{ arg .LicensePlate }}, is_available = {{ arg .IsAvailable }}, updated_at = {{ arg .UpdatedAt }}
WHERE id = {{ arg .ID }}
RETURNING updated_at;

-- name: DeleteCar
DELETE FROM cars WHERE id = {{ arg .ID }};

-- name: TransferCarOwnership
UPDATE users_cars
SET user_id = {{ arg .NewUserID }}
WHERE car_id = {{ arg .CarID }};

-- name: BulkUpdateCarAvailability
UPDATE cars
SET is_available = {{ arg .IsAvailable }}, updated_at = {{ arg .UpdatedAt }}
WHERE id = ANY(ARRAY[{{ range $i, $id := .CarIDs }}{{ if $i }},{{ end }}{{ $id }}{{ end }}]);

-- name: CheckCarOwnership
SELECT COUNT(*)
FROM users_cars
WHERE car_id = {{ arg .CarID }} AND user_id = {{ arg .UserID }};

-- name: CheckCarsOwnership
SELECT car_id
FROM users_cars
WHERE car_id = ANY(ARRAY[{{ range $i, $id := .CarIDs }}{{ if $i }},{{ end }}{{ arg $id }}{{ end }}]) AND user_id = {{ arg .UserID }};