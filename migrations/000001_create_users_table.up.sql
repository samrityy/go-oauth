CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    provider TEXT NOT NULL,
    provider_id TEXT NOT NULL UNIQUE,
    login TEXT,
    name TEXT,
    email TEXT,
    avatar_url TEXT,
    access_token TEXT,
    refresh_token TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
