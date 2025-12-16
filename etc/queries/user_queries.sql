-- name: CreateUser
INSERT INTO users (name, email, age)
VALUES ($1, $2, $3)
RETURNING id, created_at, updated_at;

-- name: FindUserByID
SELECT id, name, email, age, created_at, updated_at 
FROM users 
WHERE id = $1;

-- name: FindAllUsersBase
SELECT id, name, email, age, created_at, updated_at 
FROM users
WHERE 1=1
{{if .Name}}
  AND name ILIKE '%' || $name || '%'
{{end}}
{{if .Email}}
  AND email ILIKE '%' || $email || '%'
{{end}}
{{if .MinAge}}
  AND age >= $min_age
{{end}}
{{if .MaxAge}}
  AND age <= $max_age
{{end}}
{{if .SortBy}}
  ORDER BY {{.SortBy}} {{.SortDir}}
{{else}}
  ORDER BY id ASC
{{end}}
LIMIT $limit OFFSET $offset;

-- name: CountUsersBase
SELECT COUNT(*) 
FROM users
WHERE 1=1
{{if .Name}}
  AND name ILIKE '%' || $name || '%'
{{end}}
{{if .Email}}
  AND email ILIKE '%' || $email || '%'
{{end}}
{{if .MinAge}}
  AND age >= $min_age
{{end}}
{{if .MaxAge}}
  AND age <= $max_age
{{end}};

-- name: UpdateUser
UPDATE users
SET name = $1, email = $2, age = $3, updated_at = $4
WHERE id = $5;

-- name: DeleteUser
DELETE FROM users WHERE id = $1;

-- name: CheckEmailExists
SELECT COUNT(*) FROM users WHERE email = $1 AND id != $2;

-- name: BulkInsertUsers
INSERT INTO users (name, email, age, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5);