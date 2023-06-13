CREATE TABLE IF NOT EXISTS "links" (
    id integer PRIMARY KEY,
    created_at datetime,
    updated_at datetime,
    deleted_at datetime,
    url text,
    title text,
    description text,
    saved_at datetime,
    read_at datetime
);

CREATE INDEX idx_links_url ON links (url);
CREATE INDEX idx_links_deleted_at ON links (deleted_at);
