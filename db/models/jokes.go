package models

import (
	"errors"
	"math"
	"time"

	"gorm.io/gorm"
)

type JokesEvaluations struct {
	ID         uint  `gorm:"primaryKey"`
	UserID     uint  `gorm:"index"`
	User       User  `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	JokeId     uint  `gorm:"column:joke_id;index"`
	Joke       Jokes `gorm:"foreignKey:JokeId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Evaluation uint
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (JokesEvaluations) TableName() string {
	return "JokesEvaluations"
}

type Jokes struct {
	ID             uint   `gorm:"primaryKey"`
	UserID         uint   `gorm:"index"`
	User           User   `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Text           string `gorm:"type:text"`
	Evaluations    uint
	AuthorUserName string
	AnonymsMode    bool
	AVGScore       uint
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (Jokes) TableName() string {
	return "jokes"
}

func (joke *Jokes) AddEvaluation(db *gorm.DB, user *User, evaluation uint) error {
	if evaluation < 1 || evaluation > 5 {
		return errors.New("оценка должна быть от 1 до 5")
	}

	eval := JokesEvaluations{
		UserID:     user.ID,
		JokeId:     joke.ID,
		Evaluation: evaluation,
	}
	if err := db.Create(&eval).Error; err != nil {
		return errors.New("не удалось сохранить оценку")
	}
	var Author User
	db.Where(&User{ID: joke.UserID}).First(&Author)
	Author.AddTokenForEvaluationAuthor(db, uint64(evaluation))
	user.AddTokenForEvaluation(db)
	newEvaluations := joke.Evaluations + 1
	newAvgScore := ((float64(joke.AVGScore) * float64(joke.Evaluations)) + (float64(evaluation) * 20)) / float64(newEvaluations)

	finalScore := uint(math.Round(newAvgScore))
	if finalScore < 1 {
		finalScore = 1
	} else if finalScore > 100 {
		finalScore = 100
	}

	joke.AVGScore = finalScore
	joke.Evaluations = newEvaluations

	return db.Save(joke).Error
}
