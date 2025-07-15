package models

import "time"

type Quote struct {
	ID          uint   `gorm:"primaryKey"`
	Content     string `gorm:"not null"`
	AuthorID    uint
	Author      User
	IsAnonymous bool `gorm:"default:false" json:"is_anonymous"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Votes       []Vote
}
