-- name: CreateCar
INSERT INTO cars (user_id, brand, model, year, color, license_plate, is_available)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, created_at, updated_at;