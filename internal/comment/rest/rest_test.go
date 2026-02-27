package rest_test

import (
	"bytes"
	"comment-tree/internal/comment/mocks"
	"comment-tree/internal/comment/rest"
	"comment-tree/internal/comment/service"
	"comment-tree/internal/comment/types/dto"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/wb-go/wbf/ginext"
	"go.uber.org/mock/gomock"
)

func setupContext(method, url string, body []byte) (*ginext.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := httptest.NewRequest(method, url, bytes.NewReader(body))
	c.Request = req

	return c, w
}

func TestCommentHandler_CreateComment(t *testing.T) {
	tests := []struct {
		name       string
		body       any
		setupMocks func(ctrl *gomock.Controller) (*mocks.MockComment, *mocks.MockValidator)
		wantStatus int
	}{
		{
			name: "invalid json",
			body: `invalid-json`,
			setupMocks: func(ctrl *gomock.Controller) (*mocks.MockComment, *mocks.MockValidator) {
				return mocks.NewMockComment(ctrl), mocks.NewMockValidator(ctrl)
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "validation error",
			body: dto.CreateComment{UserID: 1, Text: "text"},
			setupMocks: func(ctrl *gomock.Controller) (*mocks.MockComment, *mocks.MockValidator) {
				comment := mocks.NewMockComment(ctrl)
				validator := mocks.NewMockValidator(ctrl)

				validator.EXPECT().
					Validate(gomock.Any()).
					Return(errors.New("validation failed"))

				return comment, validator
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "parent not found",
			body: dto.CreateComment{UserID: 1, Text: "text"},
			setupMocks: func(ctrl *gomock.Controller) (*mocks.MockComment, *mocks.MockValidator) {
				comment := mocks.NewMockComment(ctrl)
				validator := mocks.NewMockValidator(ctrl)

				validator.EXPECT().
					Validate(gomock.Any()).
					Return(nil)

				comment.EXPECT().
					SaveComment(gomock.Any(), gomock.Any()).
					Return(0, service.ErrParentNotFound)

				return comment, validator
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "internal error",
			body: dto.CreateComment{UserID: 1, Text: "text"},
			setupMocks: func(ctrl *gomock.Controller) (*mocks.MockComment, *mocks.MockValidator) {
				comment := mocks.NewMockComment(ctrl)
				validator := mocks.NewMockValidator(ctrl)

				validator.EXPECT().
					Validate(gomock.Any()).
					Return(nil)

				comment.EXPECT().
					SaveComment(gomock.Any(), gomock.Any()).
					Return(0, errors.New("db error"))

				return comment, validator
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "success",
			body: dto.CreateComment{UserID: 1, Text: "text"},
			setupMocks: func(ctrl *gomock.Controller) (*mocks.MockComment, *mocks.MockValidator) {
				comment := mocks.NewMockComment(ctrl)
				validator := mocks.NewMockValidator(ctrl)

				validator.EXPECT().
					Validate(gomock.Any()).
					Return(nil)

				comment.EXPECT().
					SaveComment(gomock.Any(), gomock.Any()).
					Return(100, nil)

				return comment, validator
			},
			wantStatus: http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			commentMock, validatorMock := tt.setupMocks(ctrl)
			handler := rest.NewCommentHandler(commentMock, validatorMock)

			var bodyBytes []byte
			switch v := tt.body.(type) {
			case string:
				bodyBytes = []byte(v)
			default:
				bodyBytes, _ = json.Marshal(v)
			}

			ctx, w := setupContext(http.MethodPost, "/comments", bodyBytes)

			handler.CreateComment(ctx)

			if w.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}

func TestCommentHandler_GetCommentTree(t *testing.T) {
	tests := []struct {
		name       string
		idParam    string
		setupMocks func(ctrl *gomock.Controller) *mocks.MockComment
		wantStatus int
	}{
		{
			name:    "invalid id",
			idParam: "abc",
			setupMocks: func(ctrl *gomock.Controller) *mocks.MockComment {
				return mocks.NewMockComment(ctrl)
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "internal error",
			idParam: "1",
			setupMocks: func(ctrl *gomock.Controller) *mocks.MockComment {
				mock := mocks.NewMockComment(ctrl)
				mock.EXPECT().
					GetCommentsByParent(gomock.Any(), 1).
					Return(dto.Comments{}, errors.New("db error"))
				return mock
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:    "success",
			idParam: "1",
			setupMocks: func(ctrl *gomock.Controller) *mocks.MockComment {
				mock := mocks.NewMockComment(ctrl)
				mock.EXPECT().
					GetCommentsByParent(gomock.Any(), 1).
					Return(dto.Comments{}, nil)
				return mock
			},
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			commentMock := tt.setupMocks(ctrl)
			handler := rest.NewCommentHandler(commentMock, nil)

			ctx, w := setupContext(http.MethodGet, "/comments/"+tt.idParam, nil)
			ctx.Params = gin.Params{{Key: "id", Value: tt.idParam}}

			handler.GetCommentTree(ctx)

			if w.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}

func TestCommentHandler_GetComments(t *testing.T) {
	tests := []struct {
		name       string
		query      string
		setupMocks func(ctrl *gomock.Controller) *mocks.MockComment
		wantStatus int
	}{
		{
			name:  "success with defaults",
			query: "",
			setupMocks: func(ctrl *gomock.Controller) *mocks.MockComment {
				mock := mocks.NewMockComment(ctrl)
				mock.EXPECT().
					GetComments(gomock.Any(), "", 1, 10, "ASC").
					Return(dto.Comments{}, nil)
				return mock
			},
			wantStatus: http.StatusOK,
		},
		{
			name:  "invalid sort fallback",
			query: "?sort=invalid&page=-1&page_size=0",
			setupMocks: func(ctrl *gomock.Controller) *mocks.MockComment {
				mock := mocks.NewMockComment(ctrl)
				mock.EXPECT().
					GetComments(gomock.Any(), "", 1, 10, "ASC").
					Return(dto.Comments{}, nil)
				return mock
			},
			wantStatus: http.StatusOK,
		},
		{
			name:  "internal error",
			query: "",
			setupMocks: func(ctrl *gomock.Controller) *mocks.MockComment {
				mock := mocks.NewMockComment(ctrl)
				mock.EXPECT().
					GetComments(gomock.Any(), "", 1, 10, "ASC").
					Return(dto.Comments{}, errors.New("db error"))
				return mock
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			commentMock := tt.setupMocks(ctrl)
			handler := rest.NewCommentHandler(commentMock, nil)

			ctx, w := setupContext(http.MethodGet, "/comments"+tt.query, nil)

			handler.GetComments(ctx)

			if w.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}

func TestCommentHandler_DeleteComment(t *testing.T) {
	tests := []struct {
		name       string
		idParam    string
		setupMocks func(ctrl *gomock.Controller) *mocks.MockComment
		wantStatus int
	}{
		{
			name:    "invalid id",
			idParam: "abc",
			setupMocks: func(ctrl *gomock.Controller) *mocks.MockComment {
				return mocks.NewMockComment(ctrl)
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "not found",
			idParam: "10",
			setupMocks: func(ctrl *gomock.Controller) *mocks.MockComment {
				mock := mocks.NewMockComment(ctrl)
				mock.EXPECT().
					DeleteComment(gomock.Any(), 10).
					Return(service.ErrCommentNotFound)
				return mock
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:    "internal error",
			idParam: "10",
			setupMocks: func(ctrl *gomock.Controller) *mocks.MockComment {
				mock := mocks.NewMockComment(ctrl)
				mock.EXPECT().
					DeleteComment(gomock.Any(), 10).
					Return(errors.New("db error"))
				return mock
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:    "success",
			idParam: "10",
			setupMocks: func(ctrl *gomock.Controller) *mocks.MockComment {
				mock := mocks.NewMockComment(ctrl)
				mock.EXPECT().
					DeleteComment(gomock.Any(), 10).
					Return(nil)
				return mock
			},
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			commentMock := tt.setupMocks(ctrl)
			handler := rest.NewCommentHandler(commentMock, nil)

			ctx, w := setupContext(http.MethodDelete, "/comments/"+tt.idParam, nil)
			ctx.Params = gin.Params{{Key: "id", Value: tt.idParam}}

			handler.DeleteComment(ctx)

			if w.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}
