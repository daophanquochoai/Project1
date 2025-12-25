\connect user_service;

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS users (
                                     id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name          TEXT NOT NULL,
    email         TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role          TEXT NOT NULL CHECK (role IN ('user', 'admin')),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMPTZ
    );

-- Indexes
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);

-- updated_at auto
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_users_updated_at ON users;
CREATE TRIGGER trg_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

INSERT INTO public.users (id,email,password_hash,"role",created_at,updated_at,deleted_at,name) VALUES
                                                                                                   ('2294c66d-79df-4d99-974e-6cac94bc8796'::uuid,'quochoai@example.com','$2a$10$8ynQ6uXQLdsO2nesbV0gueFrrSLvSDX93V9aCUb2TyzHFxaJ1IUuq','admin','2025-12-22 23:06:41.627707+07','2025-12-23 17:45:00.458004+07',NULL,'QuocHoai'),
                                                                                                   ('799f4f24-0b3e-4e7f-88e2-4ebe5b929e69'::uuid,'hongnghi@example.com','$2a$10$Wcr0qb.UU95hQNlvkROo2O3J98pLPqNtyPkdQkEkmXYoErN2C5Noq','user','2025-12-23 22:30:45.530594+07','2025-12-23 22:34:09.171409+07',NULL,'HongNghi');
