CREATE TABLE oauth_accounts (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,  -- foreign key
    provider TEXT NOT NULL,                 
    provider_id TEXT NOT NULL,              
    access_token TEXT,
    refresh_token TEXT,
    token_expiry TIMESTAMPTZ,

    UNIQUE (provider, provider_id),         
    UNIQUE (user_id, provider)            
);
