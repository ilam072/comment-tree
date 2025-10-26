package dto

type Comment struct {
	ID       int    `json:"id"`
	Text     string `json:"text" validate:"required"`
	ParentID *int   `json:"parent_id,omitempty"`
	UserID   int    `json:"user_id" validate:"required"`
}

type Comments struct {
	Comments []Comment `json:"comments"`
}
