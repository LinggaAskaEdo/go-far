-- name: CreateCar
INSERT INTO cars (user_id, brand, model, year, color, license_plate, is_available)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, created_at, updated_at;

-- name: CreateCarBulk
INSERT INTO cars (user_id, brand, model, year, color, license_plate, is_available, created_at, updated_at)
VALUES {{range $index, $car := .}}
  {{if $index}},{{end}}(${{add $index 1}}, ${{add $index 2}}, ${{add $index 3}}, ${{add $index 4}}, ${{add $index 5}}, ${{add $index 6}}, ${{add $index 7}}, ${{add $index 8}}, ${{add $index 9}})
{{end}};

-- name: FindCarByID
SELECT id, user_id, brand, model, year, color, license_plate, is_available, created_at, updated_at
FROM cars
WHERE id = $1;

-- name: FindCarByIDWithOwner
SELECT 
    c.id, 
    c.user_id, 
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
INNER JOIN users u ON c.user_id = u.id
WHERE c.id = $1;

-- name: FindCarsByUserID
SELECT id, user_id, brand, model, year, color, license_plate, is_available, created_at, updated_at
FROM cars
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: CountCarsByUserID
SELECT COUNT(*)
FROM cars
WHERE user_id = $1;

-- name: UpdateCar
UPDATE cars
SET brand = $1, model = $2, year = $3, color = $4, license_plate = $5, is_available = $6, updated_at = $7
WHERE id = $8
RETURNING updated_at;

-- name: DeleteCar
DELETE FROM cars WHERE id = $1;

-- name: TransferCarOwnership
UPDATE cars
SET user_id = $1, updated_at = $2
WHERE id = $3;

-- name: BulkUpdateCarAvailability
UPDATE cars
SET is_available = $1, updated_at = $2
WHERE id = ANY($3);