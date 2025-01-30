package models

import "time"

type Jokes struct {
	ID          uint   `gorm:"primaryKey"`
	UserID      uint   `gorm:"index"`
	User        User   `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Text        string `gorm:"type:text"`
	Evaluations uint64
	AVGScore    uint8
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (Jokes) TableName() string {
	return "jokes"
}
