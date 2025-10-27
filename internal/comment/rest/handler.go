package rest

import (
	"comment-tree/internal/comment/service"
	"comment-tree/internal/comment/types/dto"
	"comment-tree/internal/response"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/zlog"
	"net/http"
	"strconv"
	"strings"
)

type Comment interface {
	SaveComment(ctx context.Context, comment dto.CreateComment) (int, error)
	GetCommentsByParent(ctx context.Context, parentID int) (dto.Comments, error)
	GetComments(ctx context.Context, search string, page, pageSize int, sort string) (dto.Comments, error)
	DeleteComment(ctx context.Context, id int) error
}

type Validator interface {
	Validate(i interface{}) error
}

type CommentHandler struct {
	comment   Comment
	validator Validator
}

func NewCommentHandler(comment Comment, validator Validator) *CommentHandler {
	return &CommentHandler{comment: comment, validator: validator}
}

func (h *CommentHandler) CreateComment(c *ginext.Context) {
	var comment dto.CreateComment
	if err := json.NewDecoder(c.Request.Body).Decode(&comment); err != nil {
		zlog.Logger.Error().Err(err).Msg("failed to decode")
		response.Error("invalid request body").WriteJSON(c, http.StatusBadRequest)
		return
	}

	if err := h.validator.Validate(comment); err != nil {
		response.Error(fmt.Sprintf("validation error: %s", err.Error())).WriteJSON(c, http.StatusBadRequest)
		return
	}

	commentID, err := h.comment.SaveComment(c.Request.Context(), comment)
	if err != nil {
		if errors.Is(err, service.ErrParentNotFound) {
			response.Error("parent with such id not found").WriteJSON(c, http.StatusNotFound)
			return
		}
		zlog.Logger.Error().Err(err).Interface("comment", comment).Msg("failed to save comment")
		response.Error("internal server error, try again later").WriteJSON(c, http.StatusInternalServerError)
		return
	}

	response.Success(commentID).WriteJSON(c, http.StatusCreated)
}

func (h *CommentHandler) GetCommentTree(c *ginext.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Error("invalid id, must be integer").WriteJSON(c, http.StatusBadRequest)
		return
	}

	comments, err := h.comment.GetCommentsByParent(c.Request.Context(), id)
	if err != nil {
		zlog.Logger.Error().Err(err).Int("id", id).Msg("failed to get comments by parent")
		response.Error("internal server error, try again later").WriteJSON(c, http.StatusInternalServerError)
		return
	}

	response.Raw(c, http.StatusOK, comments)
}

func (h *CommentHandler) GetComments(c *ginext.Context) {
	search := c.Query("search")

	sort := strings.ToUpper(c.DefaultQuery("sort", "ASC"))
	if sort != "ASC" && sort != "DESC" {
		sort = "ASC"
	}

	// page
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page <= 0 {
		page = 1
	}

	// pageSize
	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if err != nil || pageSize <= 0 {
		pageSize = 10
	}

	comments, err := h.comment.GetComments(c.Request.Context(), search, page, pageSize, sort)
	if err != nil {
		zlog.Logger.Error().Err(err).Msg("failed to get comments")
		response.Error("internal server error, try again later").WriteJSON(c, http.StatusInternalServerError)
		return
	}

	response.Raw(c, http.StatusOK, comments)
}

func (h *CommentHandler) DeleteComment(c *ginext.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Error("invalid id, must be integer").WriteJSON(c, http.StatusBadRequest)
		return
	}

	if err := h.comment.DeleteComment(c.Request.Context(), id); err != nil {
		if errors.Is(err, service.ErrCommentNotFound) {
			response.Error("comment with such id not found").WriteJSON(c, http.StatusNotFound)
			return
		}
		zlog.Logger.Error().Err(err).Int("id", id).Msg("failed to delete comment")
		response.Error("internal server error, try again later").WriteJSON(c, http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusOK)
}
