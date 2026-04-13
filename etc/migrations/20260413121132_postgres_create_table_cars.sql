-- +goose Up
-- +goose StatementBegin

-- Create cars table
CREATE TABLE public.cars (
    id             uuid         DEFAULT uuidv7() NOT NULL,
    brand          varchar(100) NOT NULL,
    model          varchar(100) NOT NULL,
    "year"         int4         NOT NULL,
    color          varchar(50)  NULL,
    license_plate  varchar(20)  NOT NULL,
    is_available   bool         NOT NULL DEFAULT true,
    created_at     timestamptz  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     timestamptz  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT cars_pkey PRIMARY KEY (id),
    CONSTRAINT cars_license_plate_key UNIQUE (license_plate),
    CONSTRAINT cars_year_check CHECK ("year" >= 1900 AND "year" <= 2100)
);

-- Create indexes
CREATE INDEX idx_cars_brand ON public.cars USING btree (brand);
CREATE INDEX idx_cars_is_available ON public.cars USING btree (is_available);
CREATE INDEX idx_cars_license_plate ON public.cars USING btree (license_plate);

-- Auto-update updated_at trigger
CREATE TRIGGER set_cars_updated_at
    BEFORE UPDATE ON public.cars
    FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TRIGGER IF EXISTS set_cars_updated_at ON public.cars;
DROP INDEX IF EXISTS idx_cars_license_plate;
DROP INDEX IF EXISTS idx_cars_is_available;
DROP INDEX IF EXISTS idx_cars_brand;
DROP TABLE IF EXISTS public.cars;

-- +goose StatementEnd