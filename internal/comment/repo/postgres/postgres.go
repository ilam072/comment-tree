package postgres

import (
	"comment-tree/internal/comment/types/domain"
	"comment-tree/pkg/errutils"
	"context"
	"fmt"
	"github.com/wb-go/wbf/dbpg"
	"strings"
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
        INSERT INTO comments(text, parent_id, user_id)
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
	const op = "repo.comment.GetCommentsByParent"

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

func (r *CommentRepo) GetComments(ctx context.Context, search string, page, pageSize int, sort string) ([]domain.Comment, error) {
	const op = "repo.comment.GetComments"

	offset := (page - 1) * pageSize

	query := `SELECT id, parent_id, text, user_id, created_at FROM comments`

	var (
		conditions []string
		args       []interface{}
	)

	if search != "" {
		conditions = append(conditions, fmt.Sprintf("document @@ plainto_tsquery('russian', $%d)", len(args)+1))
		args = append(args, search)
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += fmt.Sprintf(" ORDER BY created_at %s", sort)

	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(args)+1, len(args)+2)
	args = append(args, pageSize, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
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

	return comments, rows.Err()
}
