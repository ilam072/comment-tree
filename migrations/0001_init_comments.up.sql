CREATE TABLE IF NOT EXISTS comments (
                                        id SERIAL PRIMARY KEY,
                                        parent_id INT REFERENCES comments(id) ON DELETE CASCADE,
                                        user_id INT NOT NULL,
                                        text TEXT NOT NULL,
                                        created_at TIMESTAMP DEFAULT NOW(),
                                        deleted BOOLEAN DEFAULT FALSE,
                                        document tsvector
);

CREATE INDEX IF NOT EXISTS idx_comments_parent_id ON comments(parent_id);
CREATE INDEX IF NOT EXISTS idx_comments_created_at ON comments(created_at);
CREATE INDEX IF NOT EXISTS idx_comments_user_id ON comments(user_id);

CREATE INDEX IF NOT EXISTS idx_comments_document
    ON comments USING GIN (document);

CREATE OR REPLACE FUNCTION comments_tsvector_trigger() RETURNS trigger AS $$
BEGIN
    NEW.document := to_tsvector('russian', NEW.text);
    RETURN NEW;
END
$$ LANGUAGE plpgsql;

CREATE TRIGGER comments_tsvector_update
    BEFORE INSERT OR UPDATE ON comments
    FOR EACH ROW EXECUTE FUNCTION comments_tsvector_trigger();