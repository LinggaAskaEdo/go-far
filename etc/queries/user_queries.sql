-- name: CreateUser
INSERT INTO users (email, password, name, age, role, is_active)
VALUES ({{ .Email }}, {{ .Password }}, {{ .Name }}, {{ .Age }}, {{ .Role }}, true)
RETURNING id, created_at, updated_at;

-- name: FindUserByID
SELECT id, email, name, password, age, role, is_active, created_at, updated_at
FROM users
WHERE id = {{ arg .ID }};

-- name: FindUserByEmail
SELECT id, email, name, password, age, role, is_active, created_at, updated_at
FROM users
WHERE email = {{ arg .Email }};

-- name: FindAllUsersBase
SELECT id, email, name, age, role, is_active, created_at, updated_at
FROM users
WHERE 1=1
{{ if .ID }}
    AND id = {{ arg .ID }}
{{ end }}
{{ if .Name }}
    AND name ILIKE {{ arg .NamePattern }}
{{ end }}
{{ if .Email }}
    AND email ILIKE {{ arg .EmailPattern }}
{{ end }}
{{ if gt .MinAge 0 }}
    AND age >= {{ arg .MinAge }}
{{ end }}
{{ if gt .MaxAge 0 }}
    AND age <= {{ arg .MaxAge }}
{{ end }}
{{ if .SortBy }}
    ORDER BY {{ raw .SortBy }} {{ raw .SortDir }}
{{ end }}
LIMIT {{ arg .Limit }} OFFSET {{ arg .Offset }};

-- name: CountUsersBase
SELECT COUNT(*)
FROM users
WHERE 1=1
{{ if .ID }}
    AND id = {{ arg .ID }}
{{ end }}
{{ if .Name }}
    AND name ILIKE {{ arg .NamePattern }}
{{ end }}
{{ if .Email }}
    AND email ILIKE {{ arg .EmailPattern }}
{{ end }}
{{ if gt .MinAge 0 }}
    AND age >= {{ arg .MinAge }}
{{ end }}
{{ if gt .MaxAge 0 }}
    AND age <= {{ arg .MaxAge }}
{{ end }};

-- name: UpdateUser
UPDATE users
SET email = {{ arg .Email }}, name = {{ arg .Name }}, age = {{ arg .Age }}, role = {{ arg .Role }}, is_active = {{ arg .IsActive }}, updated_at = {{ arg .UpdatedAt }}
WHERE id = {{ arg .ID }};

-- name: DeleteUser
DELETE FROM users WHERE id = {{ arg .ID }};

-- name: CheckEmailExists
SELECT COUNT(*) FROM users WHERE email = {{ arg .Email }} AND id != {{ arg .ID }};

-- name: BulkInsertUsers
INSERT INTO users (email, name, age, role)
VALUES
{{ range $i, $user := .Users }}
  {{ if $i }},{{ end }}
  ({{ arg $user.Email }}, {{ arg $user.Name }}, {{ arg $user.Age }}, {{ arg $user.Role }})
{{ end }}

-- name: FindUsersBaseV2
SELECT id, email, name, age, role, is_active, created_at, updated_at
FROM users;
