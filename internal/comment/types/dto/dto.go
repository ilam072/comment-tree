package dto

type CreateComment struct {
	Text     string `json:"text" validate:"required"`
	ParentID *int   `json:"parent_id,omitempty"`
	UserID   int    `json:"user_id" validate:"required"`
}

type GetComment struct {
	ID       int    `json:"id"`
	Text     string `json:"text" validate:"required"`
	ParentID *int   `json:"parent_id,omitempty"`
	UserID   int    `json:"user_id" validate:"required"`
}

type Comments struct {
	Comments []GetComment `json:"comments"`
}
