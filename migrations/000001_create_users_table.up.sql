CREATE TABLE IF NOT EXISTS users (
    id bigserial PRIMARY KEY,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    name text NOT NULL,
    email text UNIQUE NOT NULL,
    password_hash text NOT NULL, -- We store a hash, not the password
    role text NOT NULL DEFAULT 'user', -- For admin/user roles
    version integer NOT NULL DEFAULT 1
);