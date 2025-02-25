package Utilities

import (
	"SHUTKANULbot/db"
	"SHUTKANULbot/db/models"
	"fmt"
	"log"
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
	ID          uint   `gorm:"column:id"`
	Text        string `gorm:"column:text"`
	Author      string `gorm:"column:author_user_name"`
	AnonymsMode bool   `gorm:"column:anonyms_mode"`
}

func GetFifteenJokes(UserId uint) []Joke {
	var evaluatedJokeIDs []uint
	db.DB.Model(&models.JokesEvaluations{}).
		Where(`"user_id" = ? AND "created_at" >= ?`, UserId, time.Now().Add(-24*time.Hour)).
		Pluck(`"joke_id"`, &evaluatedJokeIDs)

	evaluatedJokeIDs = removeDuplicates(evaluatedJokeIDs)

	if len(evaluatedJokeIDs) == 0 {
		evaluatedJokeIDs = append(evaluatedJokeIDs, 0)
	}

	var jokes []Joke

	db.DB.Raw(`
    WITH ranked_jokes AS (
        SELECT DISTINCT ON (j.id) 
            j.id, 
            j.text, 
            j.author_user_name, 
            j.anonyms_mode,
            1 as category,
            ROW_NUMBER() OVER (
                PARTITION BY 1 
                ORDER BY (j.avg_score * j.evaluations) / (POWER(DATE_PART('day', NOW() - j.created_at) + 1, 1.2)) DESC
            ) as row_num
        FROM jokes j
        WHERE j.evaluations < 10 AND j.id NOT IN (?)
        
        UNION ALL

        SELECT DISTINCT ON (j.id) 
            j.id, 
            j.text, 
            j.author_user_name, 
            j.anonyms_mode,
            2 as category,
            ROW_NUMBER() OVER (
                PARTITION BY 2 
                ORDER BY ABS(
                    COALESCE(
                        (SELECT AVG(e.evaluation) FROM "JokesEvaluations" e WHERE e.joke_id = j.id GROUP BY e.joke_id) - 
                        (SELECT AVG(e.evaluation) FROM "JokesEvaluations" e WHERE e.joke_id = j.id GROUP BY e.joke_id),
                    0)
                ) DESC
            ) as row_num
        FROM jokes j
        WHERE j.evaluations BETWEEN 10 AND 50 AND j.id NOT IN (?)

        UNION ALL

        SELECT DISTINCT ON (j.id) 
            j.id, 
            j.text, 
            j.author_user_name, 
            j.anonyms_mode,
            3 as category,
            ROW_NUMBER() OVER (
                PARTITION BY 3 
                ORDER BY (100 - COALESCE(STDDEV(e.evaluation) / NULLIF(AVG(e.evaluation), 0) * 100, 0)) DESC
            ) as row_num
        FROM jokes j
        JOIN "JokesEvaluations" e ON e.joke_id = j.id
        WHERE j.evaluations BETWEEN 50 AND 100 AND j.id NOT IN (?)
        GROUP BY j.id, j.text, j.author_user_name, j.anonyms_mode

        UNION ALL

        SELECT DISTINCT ON (j.id) 
            j.id, 
            j.text, 
            j.author_user_name, 
            j.anonyms_mode,
            4 as category,
            ROW_NUMBER() OVER (
                PARTITION BY 4 
                ORDER BY j.avg_score * (1 - (DATE_PART('day', NOW() - j.created_at) / 30.0)) DESC
            ) as row_num
        FROM jokes j
        WHERE j.evaluations > 100 AND j.id NOT IN (?)
    )
    SELECT id, text, author_user_name, anonyms_mode 
    FROM ranked_jokes 
    ORDER BY category, row_num 
    LIMIT 15;
`, evaluatedJokeIDs, evaluatedJokeIDs, evaluatedJokeIDs, evaluatedJokeIDs).Scan(&jokes)

	if len(jokes) < 15 {
		var extraJokes []Joke
		db.DB.Raw(`
        SELECT DISTINCT ON (j.id) 
            j.id, 
            j.text, 
            j.author_user_name, 
            j.anonyms_mode
        FROM jokes j
        WHERE j.id NOT IN (?)
        ORDER BY j.id, j.avg_score DESC, j.evaluations DESC
        LIMIT ?`, evaluatedJokeIDs, 15-len(jokes)).Scan(&extraJokes)

		jokes = append(jokes, removeDuplicateJokes(extraJokes)...)
	}

	if len(jokes) < 15 {
		var latestJokes []Joke
		db.DB.Raw(`
        SELECT DISTINCT ON (j.id) 
            j.id, 
            j.text, 
            j.author_user_name, 
            j.anonyms_mode
        FROM jokes j
        WHERE j.id NOT IN (?)
        ORDER BY j.id, j.created_at DESC
        LIMIT ?`, evaluatedJokeIDs, 15-len(jokes)).Scan(&latestJokes)

		jokes = append(jokes, removeDuplicateJokes(latestJokes)...)
	}
	return removeDuplicateJokes(jokes)
}

func removeDuplicates(ids []uint) []uint {
	seen := make(map[uint]bool)
	result := []uint{}
	for _, id := range ids {
		if !seen[id] {
			seen[id] = true
			result = append(result, id)
		}
	}
	return result
}

func removeDuplicateJokes(jokes []Joke) []Joke {
	seen := make(map[uint]bool)
	result := []Joke{}
	for _, joke := range jokes {
		if !seen[joke.ID] {
			seen[joke.ID] = true
			result = append(result, joke)
		}
	}
	return result
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
func AddJokeEvaluation(tgId int64, jokeId uint, evaluation uint) error {
	if evaluation < 1 || evaluation > 5 {
		return fmt.Errorf("оценка должна быть от 1 до 5")
	}

	mutex.Lock()
	defer mutex.Unlock()

	var user models.User
	if err := db.DB.Where("telegram_id = ?", tgId).First(&user).Error; err != nil {
		return fmt.Errorf("пользователь с tgId %d не найден", tgId)
	}

	var joke models.Jokes
	if err := db.DB.First(&joke, jokeId).Error; err != nil {
		return fmt.Errorf("шутка с ID %d не найдена", jokeId)
	}
	joke.AddEvaluation(db.DB, &user, evaluation)
	log.Printf("Оценка обновлена: joke_id=%d, new_avg_score=%d, total_evaluations=%d", jokeId, joke.AVGScore, joke.Evaluations)
	return nil
}

func GetRandomPopularJokeSafe() Joke {
	var joke Joke

	popularQuery := `
		WITH popular AS (
			SELECT id, text, author_user_name, anonyms_mode
			FROM jokes
			WHERE evaluations >= 10 AND created_at >= NOW() - INTERVAL '7 days'
			ORDER BY (avg_score * evaluations) DESC
			LIMIT 100
		)
		SELECT id, text, author_user_name, anonyms_mode
		FROM popular
		ORDER BY RANDOM()
		LIMIT 1;
	`

	err := db.DB.Raw(popularQuery).Scan(&joke).Error
	if err != nil || joke.ID == 0 {
		fallbackQuery := `
			SELECT id, text, author_user_name, anonyms_mode
			FROM jokes
			ORDER BY RANDOM()
			LIMIT 1;
		`
		errFallback := db.DB.Raw(fallbackQuery).Scan(&joke).Error
		if errFallback != nil || joke.ID == 0 {
			joke = Joke{
				ID:          0,
				Text:        "Шуток не найдено, попробуй позже.",
				Author:      "",
				AnonymsMode: true,
			}
		}
	}
	return joke
}
func GetJokeByID(jokeID uint) Joke {
	var joke Joke
	err := db.DB.Raw(`
		SELECT id, text, author_user_name, anonyms_mode
		FROM jokes
		WHERE id = ?
		LIMIT 1
	`, jokeID).Scan(&joke).Error

	if err != nil || joke.ID == 0 {
		return Joke{
			ID:          jokeID,
			Text:        "Шутка не найдена, попробуй позже.",
			Author:      "system",
			AnonymsMode: false,
		}
	}

	return joke
}

func HasUserEvaluatedJoke(userID uint, jokeID uint) bool {
	var count int64
	err := db.DB.
		Model(&models.JokesEvaluations{}).
		Where(&models.JokesEvaluations{UserID: userID, JokeId: jokeID}).
		Count(&count).Error

	if err != nil {
		return false
	}
	return count > 0
}
func GetRemainingCooldown(userIDTG uint) (string, bool) {
	var lastJokeTime time.Time
	var user models.User
	db.DB.Where("telegram_id = ?", userIDTG).First(&user)

	err := db.DB.Raw(`
		SELECT created_at
		FROM jokes
		WHERE user_id = ?
		ORDER BY created_at DESC
		LIMIT 1
	`, user.ID).Scan(&lastJokeTime).Error
	if err != nil || lastJokeTime.IsZero() {
		fmt.Print(lastJokeTime)
		log.Printf("Ошибка при получении последней шутки для пользователя %d: %v", user.ID, err)
		return "", false
	}

	cooldown := 12 * time.Hour
	elapsed := time.Since(lastJokeTime)
	if elapsed < cooldown {
		remaining := cooldown - elapsed
		hours := int(remaining.Hours())
		minutes := int(remaining.Minutes()) % 60
		return fmt.Sprintf("%dч %dм", hours, minutes), true
	}
	return "", false
}
func GetBestJoke() (Joke, error) {
	var bestJoke Joke
	query := `
		SELECT id, text, author_user_name, anonyms_mode
		FROM jokes
		WHERE created_at >= NOW() - INTERVAL '1 day'
		ORDER BY (avg_score * evaluations) / (POWER(DATE_PART('day', NOW() - created_at) + 1, 1.2)) DESC
		LIMIT 1;
	`
	err := db.DB.Raw(query).Scan(&bestJoke).Error
	if err != nil || bestJoke.ID == 0 {
		return Joke{}, fmt.Errorf("не удалось найти лучшую шутку")
	}
	return bestJoke, nil
}

func init() {
	go startCacheCleaner()
}
