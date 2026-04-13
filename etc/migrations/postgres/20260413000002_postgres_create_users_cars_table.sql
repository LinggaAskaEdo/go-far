-- +goose Up
-- Create users_cars junction table (many-to-many relationship)
CREATE TABLE public.users_cars (
    user_id    uuid         NOT NULL,
    car_id     uuid         NOT NULL,
    created_at timestamptz  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT users_cars_pkey PRIMARY KEY (user_id, car_id),
    CONSTRAINT fk_uc_user FOREIGN KEY (user_id)
        REFERENCES public.users(id) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT fk_uc_car FOREIGN KEY (car_id)
        REFERENCES public.cars(id) ON DELETE CASCADE ON UPDATE CASCADE
);

-- Create index for reverse lookup (find users by car)
CREATE INDEX idx_users_cars_car ON public.users_cars USING btree (car_id);

-- +goose Down
DROP INDEX IF EXISTS idx_users_cars_car;
DROP TABLE IF EXISTS public.users_cars;
