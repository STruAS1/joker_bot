package InlineQuery

import (
	"SHUTKANULbot/Utilities"
	"SHUTKANULbot/db"
	"SHUTKANULbot/db/models"
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func HandleInlineQuery(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	if update.InlineQuery == nil {
		return
	}
	query := update.InlineQuery.Query
	var user models.User
	if err := db.DB.Where(&models.User{TelegramID: update.InlineQuery.From.ID}).First(&user).Error; err != nil {
		return
	}

	var results []interface{}

	if query == "" {
		randomJoke := Utilities.GetRandomPopularJokeSafe()
		text := randomJoke.Text
		if !randomJoke.AnonymsMode && randomJoke.Author != "" {
			text += "\n\n<b><i>Автор:</i></b> @" + strings.TrimPrefix(randomJoke.Author, "@")
		}

		randomArticle := tgbotapi.InlineQueryResultArticle{
			Type:  "article",
			ID:    "random",
			Title: "Рандомная шутка",
			InputMessageContent: tgbotapi.InputTextMessageContent{
				Text:      text,
				ParseMode: tgbotapi.ModeHTML,
			},
			ReplyMarkup: &tgbotapi.InlineKeyboardMarkup{
				InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
					{
						tgbotapi.NewInlineKeyboardButtonURL("Оценить шутку", fmt.Sprintf("https://t.me/JOKER8BOT?start=joke_%d", randomJoke.ID)),
					},
				},
			},
		}
		results = append(results, randomArticle)
	}

	if query != "" {
		id, err := strconv.ParseUint(query, 10, 0)
		if err == nil {
			var joke models.Jokes
			if err := db.DB.Where(&models.Jokes{ID: uint(id)}).First(&joke).Error; err == nil {
				ShortName := []rune(Utilities.RemoveHTMLTags(joke.Text))
				if len(ShortName) > 20 {
					ShortName = ShortName[:20]
				}
				jokeText := joke.Text
				if !joke.AnonymsMode && joke.AuthorUserName != "" {
					jokeText += "\n\n<b><i>Автор:</i></b> @" + strings.TrimPrefix(joke.AuthorUserName, "@")
				}
				article := tgbotapi.InlineQueryResultArticle{
					Type:  "article",
					ID:    strconv.Itoa(int(joke.ID)),
					Title: fmt.Sprintf("%s #%d", string(ShortName), joke.ID),
					InputMessageContent: tgbotapi.InputTextMessageContent{
						Text:      jokeText,
						ParseMode: tgbotapi.ModeHTML,
					},
					ReplyMarkup: &tgbotapi.InlineKeyboardMarkup{
						InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
							{
								tgbotapi.NewInlineKeyboardButtonURL("Оценить шутку", fmt.Sprintf("https://t.me/JOKER8BOT?start=joke_%d", joke.ID)),
							},
						},
					},
				}
				results = append(results, article)
			}
		}
	} else {
		var jokes []models.Jokes
		db.DB.Where(&models.Jokes{UserID: user.ID}).Order("id DESC").Limit(50).Find(&jokes)
		for _, joke := range jokes {
			ShortName := []rune(Utilities.RemoveHTMLTags(joke.Text))
			if len(ShortName) > 20 {
				ShortName = ShortName[:20]
			}
			jokeText := joke.Text
			if !joke.AnonymsMode && joke.AuthorUserName != "" {
				jokeText += "\n\n<b><i>Автор:</i></b> @" + strings.TrimPrefix(joke.AuthorUserName, "@")
			}
			article := tgbotapi.InlineQueryResultArticle{
				Type:  "article",
				ID:    strconv.Itoa(int(joke.ID)),
				Title: fmt.Sprintf("%s #%d", string(ShortName), joke.ID),
				InputMessageContent: tgbotapi.InputTextMessageContent{
					Text:      jokeText,
					ParseMode: tgbotapi.ModeHTML,
				},
				ReplyMarkup: &tgbotapi.InlineKeyboardMarkup{
					InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
						{
							tgbotapi.NewInlineKeyboardButtonURL("Оценить шутку", fmt.Sprintf("https://t.me/JOKER8BOT?start=joke_%d", joke.ID)),
						},
					},
				},
			}
			results = append(results, article)
		}
	}

	inlineConfig := tgbotapi.InlineConfig{
		InlineQueryID: update.InlineQuery.ID,
		Results:       results,
		IsPersonal:    true,
		CacheTime:     0,
	}
	bot.Send(inlineConfig)
}
