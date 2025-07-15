package models

import "time"

type Vote struct {
	ID          uint `gorm:"primaryKey"`
	QuoteID     uint
	Quote       Quote
	VoterID     uint
	Voter       User
	IsAnonymous bool      `gorm:"default:false"`
	VoteDate    time.Time `gorm:"type:date"`
}
