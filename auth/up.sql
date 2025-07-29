CREATE TABLE IF NOT EXISTS refresh_tokens (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    refresh_token TEXT NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL

    -- cant have oreign key to table in another database as everyone is independent in a docker so cant communicate
    -- FOREIGN KEY (user_id) REFERENCES accounts(id) ON DELETE CASCADE
);

-- Create index to speed up queries by user_id
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id);