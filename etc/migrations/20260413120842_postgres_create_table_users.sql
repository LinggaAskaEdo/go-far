-- +goose Up
-- +goose StatementBegin

-- Create user_role enum type
CREATE TYPE user_role AS ENUM ('admin', 'user', 'guest');

-- Create the users table
CREATE TABLE public.users (
    id UUID DEFAULT uuidv7() NOT NULL,
    email VARCHAR(255) NOT NULL,
    password VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    age INT NULL,
    role public.user_role DEFAULT 'user'::user_role NOT NULL,
    is_active BOOLEAN DEFAULT true NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
    CONSTRAINT users_pkey PRIMARY KEY (id),
    CONSTRAINT users_email_key UNIQUE (email),
    CONSTRAINT users_age_check CHECK (age > 0 AND age <= 150)
);

-- Create indexes
CREATE INDEX idx_users_is_active ON public.users USING btree (is_active);
CREATE INDEX idx_users_name ON public.users USING btree (name);
CREATE INDEX idx_users_role ON public.users USING btree (role);

-- Auto-update updated_at trigger
CREATE OR REPLACE FUNCTION trigger_set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create the trigger to automatically update updated_at
CREATE TRIGGER set_users_updated_at
    BEFORE UPDATE ON public.users
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Drop the trigger
DROP TRIGGER IF EXISTS set_users_updated_at ON public.users;

-- Drop the table (cascade will drop indexes automatically)
DROP TABLE IF EXISTS public.users;
DROP TRIGGER IF EXISTS set_users_updated_at ON public.users;
DROP FUNCTION IF EXISTS trigger_set_updated_at();

-- +goose StatementEnd