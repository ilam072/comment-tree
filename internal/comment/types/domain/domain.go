package domain

import "time"

type Comment struct {
	ID        int
	ParentID  *int
	UserID    int
	Text      string
	CreatedAt time.Time
	Deleted   bool
}
