package service

import (
	"comment-tree/internal/comment/types/domain"
	"comment-tree/internal/comment/types/dto"
	"comment-tree/pkg/errutils"
	"context"
	"errors"
)

type CommentRepo interface {
	CreateComment(ctx context.Context, comment domain.Comment) (int, error)
	Exists(ctx context.Context, id int) (bool, error)
}

type Comment struct {
	repo CommentRepo
}

var (
	ErrParentNotFound = errors.New("parent comment not found")
)

func New(repo CommentRepo) *Comment {
	return &Comment{repo: repo}
}

func (c *Comment) SaveComment(ctx context.Context, comment dto.Comment) (int, error) {
	const op = "service.comment.Save"

	parentID := comment.ParentID
	if parentID != nil {
		exists, err := c.repo.Exists(ctx, *parentID)
		if err != nil {
			return 0, errutils.Wrap(op, err)
		}
		if !exists {
			return 0, ErrParentNotFound
		}
	}

	domainComment := domain.Comment{
		ParentID: comment.ParentID,
		UserID:   comment.UserID,
		Text:     comment.Text,
	}

	id, err := c.repo.CreateComment(ctx, domainComment)
	if err != nil {
		return 0, errutils.Wrap(op, err)
	}

	return id, nil
}
