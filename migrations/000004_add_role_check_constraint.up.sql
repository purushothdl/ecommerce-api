-- migrations/000004_add_role_check_constraint.up.sql
ALTER TABLE users
ADD CONSTRAINT role_check CHECK (role IN ('user', 'admin'));