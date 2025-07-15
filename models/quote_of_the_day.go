package models

import "time"

type QuoteOfTheDay struct {
	ID        uint `gorm:"primaryKey"`
	QuoteID   uint
	Quote     Quote
	Date      time.Time `gorm:"type:date;unique"`
	CreatedAt time.Time
}
