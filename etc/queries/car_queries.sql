-- name: CreateCar
INSERT INTO cars (brand, model, year, color, license_plate, is_available)
VALUES ({{ .Brand }}, {{ .Model }}, {{ .Year }}, {{ .Color }}, {{ .LicensePlate }}, {{ .IsAvailable }})
RETURNING id, created_at, updated_at;

-- name: CreateCarBulk
INSERT INTO cars (brand, model, year, color, license_plate, is_available, created_at, updated_at)
VALUES
{{- range $i, $car := . }}
  {{ if $i }},{{ end }} ({{ $car.Brand }}, {{ $car.Model }}, {{ $car.Year }}, {{ $car.Color }}, {{ $car.LicensePlate }}, {{ $car.IsAvailable }}, {{ $car.CreatedAt }}, {{ $car.UpdatedAt }})
{{- end }};

-- name: AssignCarToUser
INSERT INTO users_cars (user_id, car_id)
VALUES ({{ .UserID }}, {{ .CarID }});

-- name: AssignCarToUserBulk
INSERT INTO users_cars (user_id, car_id)
VALUES
{{- range $i, $uc := . }}
  {{ if $i }},{{ end }} ({{ $uc.UserID }}, {{ $uc.CarID }})
{{- end }};

-- name: FindCarByID
SELECT id, brand, model, year, color, license_plate, is_available, created_at, updated_at
FROM cars
WHERE id = {{ .ID }};

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
WHERE c.id = {{ .ID }};

-- name: FindCarsByUserID
SELECT c.id, c.brand, c.model, c.year, c.color, c.license_plate, c.is_available, c.created_at, c.updated_at
FROM cars c
INNER JOIN users_cars uc ON c.id = uc.car_id
WHERE uc.user_id = {{ .UserID }}
ORDER BY c.created_at DESC;

-- name: CountCarsByUserID
SELECT COUNT(*)
FROM cars c
INNER JOIN users_cars uc ON c.id = uc.car_id
WHERE uc.user_id = {{ .UserID }};

-- name: UpdateCar
UPDATE cars
SET brand = {{ .Brand }}, model = {{ .Model }}, year = {{ .Year }}, color = {{ .Color }}, license_plate = {{ .LicensePlate }}, is_available = {{ .IsAvailable }}, updated_at = {{ .UpdatedAt }}
WHERE id = {{ .ID }}
RETURNING updated_at;

-- name: DeleteCar
DELETE FROM cars WHERE id = {{ .ID }};

-- name: TransferCarOwnership
UPDATE users_cars
SET user_id = {{ .NewUserID }}
WHERE car_id = {{ .CarID }};

-- name: BulkUpdateCarAvailability
UPDATE cars
SET is_available = {{ .IsAvailable }}, updated_at = {{ .UpdatedAt }}
WHERE id = ANY(ARRAY[{{ range $i, $id := .CarIDs }}{{ if $i }},{{ end }}{{ $id }}{{ end }}]);
