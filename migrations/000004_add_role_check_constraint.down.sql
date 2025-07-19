-- migrations/000004_add_role_check_constraint.down.sql
ALTER TABLE users
DROP CONSTRAINT IF EXISTS role_check;