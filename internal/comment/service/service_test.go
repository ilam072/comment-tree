package service_test

import (
	"comment-tree/internal/comment/mocks"
	"comment-tree/internal/comment/repo"
	"comment-tree/internal/comment/service"
	"comment-tree/internal/comment/types/domain"
	"comment-tree/internal/comment/types/dto"
	"context"
	"errors"
	"reflect"
	"testing"

	"go.uber.org/mock/gomock"
)

func TestComment_SaveComment(t *testing.T) {
	type args struct {
		comment dto.CreateComment
	}

	tests := []struct {
		name      string
		setupMock func(ctrl *gomock.Controller) service.CommentRepo
		args      args
		wantID    int
		wantErr   bool
		errIs     error
	}{
		{
			name: "success without parent",
			setupMock: func(ctrl *gomock.Controller) service.CommentRepo {
				mock := mocks.NewMockCommentRepo(ctrl)

				mock.EXPECT().
					CreateComment(gomock.Any(), domain.Comment{
						ParentID: nil,
						UserID:   1,
						Text:     "text",
					}).
					Return(10, nil)

				return mock
			},
			args: args{
				comment: dto.CreateComment{
					ParentID: nil,
					UserID:   1,
					Text:     "text",
				},
			},
			wantID:  10,
			wantErr: false,
		},
		{
			name: "parent not exists",
			setupMock: func(ctrl *gomock.Controller) service.CommentRepo {
				mock := mocks.NewMockCommentRepo(ctrl)

				parentID := 5
				mock.EXPECT().
					Exists(gomock.Any(), parentID).
					Return(false, nil)

				return mock
			},
			args: func() args {
				parentID := 5
				return args{
					comment: dto.CreateComment{
						ParentID: &parentID,
						UserID:   1,
						Text:     "text",
					},
				}
			}(),
			wantID:  0,
			wantErr: true,
			errIs:   service.ErrParentNotFound,
		},
		{
			name: "exists returns error",
			setupMock: func(ctrl *gomock.Controller) service.CommentRepo {
				mock := mocks.NewMockCommentRepo(ctrl)

				parentID := 5
				mock.EXPECT().
					Exists(gomock.Any(), parentID).
					Return(false, errors.New("db error"))

				return mock
			},
			args: func() args {
				parentID := 5
				return args{
					comment: dto.CreateComment{
						ParentID: &parentID,
						UserID:   1,
						Text:     "text",
					},
				}
			}(),
			wantID:  0,
			wantErr: true,
		},
		{
			name: "create returns error",
			setupMock: func(ctrl *gomock.Controller) service.CommentRepo {
				mock := mocks.NewMockCommentRepo(ctrl)

				mock.EXPECT().
					CreateComment(gomock.Any(), gomock.Any()).
					Return(0, errors.New("insert error"))

				return mock
			},
			args: args{
				comment: dto.CreateComment{
					ParentID: nil,
					UserID:   1,
					Text:     "text",
				},
			},
			wantID:  0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc := service.New(tt.setupMock(ctrl))

			gotID, err := svc.SaveComment(context.Background(), tt.args.comment)

			if (err != nil) != tt.wantErr {
				t.Fatalf("unexpected error state: %v", err)
			}
			if tt.errIs != nil && !errors.Is(err, tt.errIs) {
				t.Fatalf("expected error %v, got %v", tt.errIs, err)
			}
			if gotID != tt.wantID {
				t.Fatalf("expected id %d, got %d", tt.wantID, gotID)
			}
		})
	}
}

func TestComment_GetCommentsByParent(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(ctrl *gomock.Controller) service.CommentRepo
		want      dto.Comments
		wantErr   bool
	}{
		{
			name: "success",
			setupMock: func(ctrl *gomock.Controller) service.CommentRepo {
				mock := mocks.NewMockCommentRepo(ctrl)

				mock.EXPECT().
					GetCommentsByParent(gomock.Any(), 1).
					Return([]domain.Comment{
						{ID: 1, Text: "a", UserID: 2},
					}, nil)

				return mock
			},
			want: dto.Comments{
				Comments: []dto.GetComment{
					{ID: 1, Text: "a", UserID: 2},
				},
			},
			wantErr: false,
		},
		{
			name: "repo error",
			setupMock: func(ctrl *gomock.Controller) service.CommentRepo {
				mock := mocks.NewMockCommentRepo(ctrl)

				mock.EXPECT().
					GetCommentsByParent(gomock.Any(), 1).
					Return(nil, errors.New("db error"))

				return mock
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc := service.New(tt.setupMock(ctrl))

			got, err := svc.GetCommentsByParent(context.Background(), 1)

			if (err != nil) != tt.wantErr {
				t.Fatalf("unexpected error: %v", err)
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("expected %+v, got %+v", tt.want, got)
			}
		})
	}
}

func TestComment_GetComments(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(ctrl *gomock.Controller) service.CommentRepo
		want      dto.Comments
		wantErr   bool
	}{
		{
			name: "success",
			setupMock: func(ctrl *gomock.Controller) service.CommentRepo {
				mock := mocks.NewMockCommentRepo(ctrl)

				mock.EXPECT().
					GetComments(gomock.Any(), "search", 1, 10, "asc").
					Return([]domain.Comment{
						{ID: 1, Text: "a", UserID: 2},
					}, nil)

				return mock
			},
			want: dto.Comments{
				Comments: []dto.GetComment{
					{ID: 1, Text: "a", UserID: 2},
				},
			},
			wantErr: false,
		},
		{
			name: "repo error",
			setupMock: func(ctrl *gomock.Controller) service.CommentRepo {
				mock := mocks.NewMockCommentRepo(ctrl)

				mock.EXPECT().
					GetComments(gomock.Any(), "search", 1, 10, "asc").
					Return(nil, errors.New("db error"))

				return mock
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc := service.New(tt.setupMock(ctrl))

			got, err := svc.GetComments(context.Background(), "search", 1, 10, "asc")

			if (err != nil) != tt.wantErr {
				t.Fatalf("unexpected error: %v", err)
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("expected %+v, got %+v", tt.want, got)
			}
		})
	}
}

func TestComment_DeleteComment(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(ctrl *gomock.Controller) service.CommentRepo
		wantErr   bool
		errIs     error
	}{
		{
			name: "success",
			setupMock: func(ctrl *gomock.Controller) service.CommentRepo {
				mock := mocks.NewMockCommentRepo(ctrl)

				mock.EXPECT().
					DeleteComment(gomock.Any(), 1).
					Return(nil)

				return mock
			},
			wantErr: false,
		},
		{
			name: "not found",
			setupMock: func(ctrl *gomock.Controller) service.CommentRepo {
				mock := mocks.NewMockCommentRepo(ctrl)

				mock.EXPECT().
					DeleteComment(gomock.Any(), 1).
					Return(repo.ErrCommentNotFound)

				return mock
			},
			wantErr: true,
			errIs:   service.ErrCommentNotFound,
		},
		{
			name: "other error",
			setupMock: func(ctrl *gomock.Controller) service.CommentRepo {
				mock := mocks.NewMockCommentRepo(ctrl)

				mock.EXPECT().
					DeleteComment(gomock.Any(), 1).
					Return(errors.New("db error"))

				return mock
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc := service.New(tt.setupMock(ctrl))

			err := svc.DeleteComment(context.Background(), 1)

			if (err != nil) != tt.wantErr {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.errIs != nil && !errors.Is(err, tt.errIs) {
				t.Fatalf("expected error %v, got %v", tt.errIs, err)
			}
		})
	}
}
