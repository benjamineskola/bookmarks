CREATE TABLE IF NOT EXISTS "users" (
    id integer PRIMARY KEY,
    created_at datetime,
    updated_at datetime,
    deleted_at datetime,
    email string UNIQUE NOT NULL,
    password string NOT NULL
);

CREATE UNIQUE INDEX idx_users_email ON users (email);
