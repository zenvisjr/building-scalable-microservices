CREATE TABLE IF NOT EXISTS accounts (
  id CHAR(27) PRIMARY KEY,
  name VARCHAR(24) NOT NULL,
  email VARCHAR(30) NOT NULL UNIQUE,
  password_hash TEXT NOT NULL,
  role TEXT NOT NULL DEFAULT 'user' CHECK (role IN ('user', 'admin')),
  token_version INT DEFAULT 1,
  is_active BOOLEAN DEFAULT TRUE
);