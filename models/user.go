package models

import "time"

type User struct {
	ID        uint      `gorm:"primaryKey"`
	Username  string    `gorm:"unique;not null" json:"username"`
	Password  string    `gorm:"not null"`
	Quotes    []Quote   `gorm:"foreignKey:AuthorID"`
	Votes     []Vote    `gorm:"foreignKey:VoterID"`
	CreatedAt time.Time `json:"created_at"`
}
