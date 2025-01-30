package Utilities

import (
	"SHUTKANULbot/db"
	"SHUTKANULbot/db/models"
	"fmt"
	"log"
	"math"
	"sync"
	"time"
)

type cacheEntry struct {
	jokes      []Joke
	lastUpdate time.Time
}

var (
	jokesCache = make(map[uint]cacheEntry)
	mutex      sync.Mutex
)

type Joke struct {
	ID     uint   `gorm:"column:id"`
	Text   string `gorm:"column:text"`
	Author string `gorm:"column:author_user_name"`
}

func GetFifteenJokes(UserId uint) []Joke {
	var evaluatedJokeIDs []uint
	db.DB.Model(&models.JokesEvaluations{}).
		Where(`"user_id" = ? AND "created_at" >= ?`, UserId, time.Now().Add(-24*time.Hour)).
		Pluck(`"joke_id"`, &evaluatedJokeIDs)
	if len(evaluatedJokeIDs) == 0 {
		evaluatedJokeIDs = append(evaluatedJokeIDs, 0)
	}
	var jokes []Joke

	db.DB.Raw(`
    WITH ranked_jokes AS (
        SELECT j.id, j.text, u.author_user_name,
            1 as category,
            ROW_NUMBER() OVER (PARTITION BY 1 ORDER BY (j.avg_score * j.evaluations) / (POWER(DATE_PART('day', NOW() - j.created_at) + 1, 1.2)) DESC) as row_num
        FROM jokes j
        JOIN "Users" u ON j.user_id = u.id
        WHERE j.evaluations < 10 AND j.id NOT IN (?)
        
        UNION ALL

        SELECT j.id, j.text, u.author_user_name,
            2 as category,
            ROW_NUMBER() OVER (PARTITION BY 2 ORDER BY ABS(
                COALESCE(
                    (SELECT AVG(e.evaluation) FROM "JokesEvaluations" e WHERE e.joke_id = j.id GROUP BY e.joke_id) - 
                    (SELECT AVG(e.evaluation) FROM "JokesEvaluations" e WHERE e.joke_id = j.id GROUP BY e.joke_id),
                0)
            ) DESC) as row_num
        FROM jokes j
        JOIN "Users" u ON j.user_id = u.id
        WHERE j.evaluations BETWEEN 10 AND 50 AND j.id NOT IN (?)

        UNION ALL

        SELECT j.id, j.text, u.author_user_name,
            3 as category,
            ROW_NUMBER() OVER (PARTITION BY 3 ORDER BY (100 - COALESCE(STDDEV(e.evaluation) / NULLIF(AVG(e.evaluation), 0) * 100, 0)) DESC) as row_num
        FROM jokes j
        JOIN "Users" u ON j.user_id = u.id
        JOIN "JokesEvaluations" e ON e.joke_id = j.id
        WHERE j.evaluations BETWEEN 50 AND 100 AND j.id NOT IN (?)
        GROUP BY j.id, j.text, u.author_user_name

        UNION ALL

        SELECT j.id, j.text, u.author_user_name,
            4 as category,
            ROW_NUMBER() OVER (PARTITION BY 4 ORDER BY j.avg_score * (1 - (DATE_PART('day', NOW() - j.created_at) / 30.0)) DESC) as row_num
        FROM jokes j
        JOIN "Users" u ON j.user_id = u.id
        WHERE j.evaluations > 100 AND j.id NOT IN (?)
    )
    SELECT id, text, author_user_name 
    FROM ranked_jokes 
    ORDER BY category, row_num 
    LIMIT 15;
`, evaluatedJokeIDs, evaluatedJokeIDs, evaluatedJokeIDs, evaluatedJokeIDs).Scan(&jokes)

	if len(jokes) < 15 {
		var extraJokes []Joke
		db.DB.Table("jokes").
			Select(`"id", "text", (SELECT "author_user_name" FROM "Users" WHERE "Users"."id" = "jokes"."user_id") AS "author_user_name"`).
			Where(`"id" NOT IN (?)`, evaluatedJokeIDs).
			Order(`"avg_score" DESC, "evaluations" DESC`).
			Limit(15 - len(jokes)).
			Find(&extraJokes)
		jokes = append(jokes, extraJokes...)
	}

	if len(jokes) < 15 {
		var latestJokes []Joke
		db.DB.Table("jokes").
			Select(`"id", "text", (SELECT "author_user_name" FROM "Users" WHERE "Users"."id" = "jokes"."user_id") AS "author_user_name"`).
			Where(`"id" NOT IN (?)`, evaluatedJokeIDs).
			Order(`"created_at" DESC`).
			Limit(15 - len(jokes)).
			Find(&latestJokes)
		jokes = append(jokes, latestJokes...)
	}

	return jokes
}

func startCacheCleaner() {
	ticker := time.NewTicker(10 * time.Minute)
	for {
		<-ticker.C
		mutex.Lock()
		for userID, entry := range jokesCache {
			if time.Since(entry.lastUpdate) > time.Hour {
				delete(jokesCache, userID)
			}
		}
		mutex.Unlock()
	}
}

func GetNextJoke(userTGId int64) (Joke, error) {
	var userID uint

	err := db.DB.Table("Users").
		Select("id").
		Where(`"telegram_id" = ?`, userTGId).
		Scan(&userID).Error

	if err != nil || userID == 0 {
		return Joke{}, err
	}

	mutex.Lock()
	defer mutex.Unlock()

	if entry, exists := jokesCache[userID]; exists {
		if time.Since(entry.lastUpdate) > time.Hour {
			delete(jokesCache, userID)
		}
	}

	if entry, exists := jokesCache[userID]; !exists || len(entry.jokes) == 0 {
		newJokes := GetFifteenJokes(userID)
		if len(newJokes) == 0 {
			return Joke{}, fmt.Errorf("нету шуток")
		}
		jokesCache[userID] = cacheEntry{
			jokes:      newJokes,
			lastUpdate: time.Now(),
		}
	}
	nextJoke := jokesCache[userID].jokes[0]
	jokesCache[userID] = cacheEntry{
		jokes:      jokesCache[userID].jokes[1:],
		lastUpdate: jokesCache[userID].lastUpdate,
	}

	return nextJoke, nil
}
func AddJokeEvaluation(tgId int64, jokeId uint, evaluation uint8) error {
	if evaluation < 1 || evaluation > 5 {
		return fmt.Errorf("оценка должна быть от 1 до 5")
	}

	mutex.Lock()
	defer mutex.Unlock()

	var user models.User
	if err := db.DB.Where("telegram_id = ?", tgId).First(&user).Error; err != nil {
		return fmt.Errorf("пользователь с tgId %d не найден", tgId)
	}

	newEvaluation := models.JokesEvaluations{
		UserID:     user.ID,
		JokeId:     jokeId,
		Evaluation: evaluation,
	}
	if err := db.DB.Create(&newEvaluation).Error; err != nil {
		return fmt.Errorf("ошибка при создании оценки: %v", err)
	}

	var joke models.Jokes
	if err := db.DB.First(&joke, jokeId).Error; err != nil {
		return fmt.Errorf("шутка с ID %d не найдена", jokeId)
	}

	newEvaluations := joke.Evaluations + 1
	newAvgScore := ((float64(joke.AVGScore) * float64(joke.Evaluations)) + (float64(evaluation) * 20)) / float64(newEvaluations)

	finalScore := uint8(math.Round(newAvgScore))
	if finalScore < 1 {
		finalScore = 1
	} else if finalScore > 100 {
		finalScore = 100
	}

	joke.AVGScore = finalScore
	joke.Evaluations = newEvaluations

	if err := db.DB.Save(&joke).Error; err != nil {
		return fmt.Errorf("ошибка при обновлении данных шутки: %v", err)
	}

	log.Printf("Оценка обновлена: joke_id=%d, new_avg_score=%d, total_evaluations=%d", jokeId, joke.AVGScore, joke.Evaluations)
	return nil
}

func init() {
	go startCacheCleaner()
}
