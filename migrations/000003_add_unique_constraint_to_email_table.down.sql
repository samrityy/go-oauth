-- Remove UNIQUE constraint
ALTER TABLE users
DROP CONSTRAINT email;

-- Allow nulls again
ALTER TABLE users
ALTER COLUMN email DROP NOT NULL;
