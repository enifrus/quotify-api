package models

import "time"

type QuoteCreateResponse struct {
	ID          uint      `json:"id"`
	Content     string    `json:"content"`
	AuthorID    uint      `json:"author_id"`
	AuthorName  string    `json:"author_username"`
	IsAnonymous bool      `json:"isanonymous"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type QuoteSearchResponse struct {
	ID          uint      `json:"id"`
	Content     string    `json:"content"`
	AuthorID    uint      `json:"author_id"`
	AuthorName  string    `json:"authorName"`
	IsAnonymous bool      `json:"is_anonymous"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	VoteCount   int       `json:"vote_points"`
}
