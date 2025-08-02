
-- First, disallow nulls
ALTER TABLE users
ALTER COLUMN email SET NOT NULL;

-- Then, add the UNIQUE constraint
ALTER TABLE users
ADD CONSTRAINT users_email_key UNIQUE (password);
