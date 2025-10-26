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
