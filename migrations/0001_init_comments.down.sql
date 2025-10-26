DROP TRIGGER IF EXISTS comments_tsvector_update ON comments;
DROP FUNCTION IF EXISTS comments_tsvector_trigger;
DROP INDEX IF EXISTS idx_comments_document;
DROP INDEX IF EXISTS idx_comments_parent_id;
DROP INDEX IF EXISTS idx_comments_created_at;
DROP INDEX IF EXISTS idx_comments_user_id;
DROP TABLE IF EXISTS comments CASCADE;
