
-- First, disallow nulls
ALTER TABLE users
ALTER COLUMN email SET NOT NULL;

-- Then, add the UNIQUE constraint
ALTER TABLE users
ADD CONSTRAINT email UNIQUE (email);
