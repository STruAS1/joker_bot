package models

import "time"

type JokesEvaluations struct {
	ID         uint  `gorm:"primaryKey"`
	UserID     uint  `gorm:"index"`
	User       User  `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	JokeId     uint  `gorm:"column:joke_id;index"`
	Joke       Jokes `gorm:"foreignKey:JokeId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Evaluation uint8
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (JokesEvaluations) TableName() string {
	return "JokesEvaluations"
}
