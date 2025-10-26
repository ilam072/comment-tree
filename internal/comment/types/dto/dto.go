package dto

type Comment struct {
	Text     string `json:"text" validate:"required"`
	ParentID *int   `json:"parent_id,omitempty"`
	UserID   int    `json:"user_id" validate:"required"`
}
