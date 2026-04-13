-- name: CreateUser
INSERT INTO users (email, password, name, age, role, is_active)
VALUES ({{ .Email }}, {{ .Password }}, {{ .Name }}, {{ .Age }}, {{ .Role }}, true)
RETURNING id, created_at, updated_at;

-- name: FindUserByID
SELECT id, email, name, password, age, role, is_active, created_at, updated_at
FROM users
WHERE id = {{ .ID }};

-- name: FindUserByEmail
SELECT id, email, name, password, age, role, is_active, created_at, updated_at
FROM users
WHERE email = {{ .Email }};

-- name: FindAllUsersBase
SELECT id, email, name, age, role, is_active, created_at, updated_at
FROM users
WHERE 1=1
{{- if .ID }}
    AND id = {{ .ID }}
{{- end }}
{{- if .Name }}
    AND name ILIKE {{ .NamePattern }}
{{- end }}
{{- if .Email }}
    AND email ILIKE {{ .EmailPattern }}
{{- end }}
{{- if gt .MinAge 0 }}
    AND age >= {{ .MinAge }}
{{- end }}
{{- if gt .MaxAge 0 }}
    AND age <= {{ .MaxAge }}
{{- end }}
ORDER BY __SORT_BY__ __SORT_DIR__
LIMIT {{ .Limit }} OFFSET {{ .Offset }};

-- name: CountUsersBase
SELECT COUNT(*)
FROM users
WHERE 1=1
{{- if .ID }}
    AND id = {{ .ID }}
{{- end }}
{{- if .Name }}
    AND name ILIKE {{ .NamePattern }}
{{- end }}
{{- if .Email }}
    AND email ILIKE {{ .EmailPattern }}
{{- end }}
{{- if gt .MinAge 0 }}
    AND age >= {{ .MinAge }}
{{- end }}
{{- if gt .MaxAge 0 }}
    AND age <= {{ .MaxAge }}
{{- end }};

-- name: UpdateUser
UPDATE users
SET email = {{ .Email }}, name = {{ .Name }}, age = {{ .Age }}, role = {{ .Role }}, is_active = {{ .IsActive }}, updated_at = {{ .UpdatedAt }}
WHERE id = {{ .ID }};

-- name: DeleteUser
DELETE FROM users WHERE id = {{ .ID }};

-- name: CheckEmailExists
SELECT COUNT(*) FROM users WHERE email = {{ .Email }} AND id != {{ .ID }};

-- name: BulkInsertUsers
INSERT INTO users (email, name, age, role, created_at, updated_at)
VALUES
{{- range $i, $user := . }}
  {{ if $i }},{{ end }} ({{ $user.Email }}, {{ $user.Name }}, {{ $user.Age }}, {{ $user.Role }}, {{ $user.CreatedAt }}, {{ $user.UpdatedAt }})
{{- end }};
