package postgres

import (
	"comment-tree/internal/comment/types/domain"
	"comment-tree/pkg/errutils"
	"context"
	"github.com/wb-go/wbf/dbpg"
)

type CommentRepo struct {
	db *dbpg.DB
}

func New(db *dbpg.DB) *CommentRepo {
	return &CommentRepo{db: db}
}

func (r *CommentRepo) CreateComment(ctx context.Context, comment domain.Comment) (int, error) {
	const op = "repo.comment.Create"

	query := `
        INSERT INTO comments (text, parent_id, user_id)
        VALUES ($1, $2, $3)
        RETURNING id;
    `
	var id int
	if err := r.db.QueryRowContext(ctx, query, comment.Text, comment.ParentID, comment.UserID).Scan(&id); err != nil {
		return 0, errutils.Wrap(op, err)
	}
	return id, nil
}

func (r *CommentRepo) Exists(ctx context.Context, id int) (bool, error) {
	const op = "repo.comment.Exists"

	query := `SELECT EXISTS(SELECT 1 FROM comments WHERE id = $1)`

	var exists bool
	if err := r.db.QueryRowContext(
		ctx,
		query,
		id,
	).Scan(&exists); err != nil {
		return false, errutils.Wrap(op, err)
	}
	return exists, nil
}

func (r *CommentRepo) GetCommentsByParent(ctx context.Context, parentID int) ([]domain.Comment, error) {
	const op = "repo.comment.GetAllNested"

	query := `WITH RECURSIVE tree AS (
            SELECT id, parent_id, text, user_id, created_at
            FROM comments
            WHERE id = $1
            UNION ALL
            SELECT c.id, c.parent_id, c.text, c.user_id, c.created_at
            FROM comments c
            JOIN tree t ON c.parent_id = t.id
        )
        SELECT id, parent_id, text, user_id, created_at FROM tree ORDER BY created_at;`

	rows, err := r.db.QueryContext(ctx, query, parentID)
	if err != nil {
		return nil, errutils.Wrap(op, err)
	}
	defer rows.Close()

	var comments []domain.Comment
	for rows.Next() {
		var c domain.Comment
		if err := rows.Scan(&c.ID, &c.ParentID, &c.Text, &c.UserID, &c.CreatedAt); err != nil {
			return nil, errutils.Wrap(op, err)
		}
		comments = append(comments, c)
	}

	return comments, nil
}
