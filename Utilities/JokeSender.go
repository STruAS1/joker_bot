package Utilities

import (
	"SHUTKANULbot/config"
	"SHUTKANULbot/db"
	"SHUTKANULbot/db/models"
	"fmt"
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func StartDailyJokeSenderMSK(cfg config.Config) {
	loc, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		log.Printf("Ошибка загрузки временной зоны: %v", err)
		return
	}

	botAPI, _ := tgbotapi.NewBotAPI(cfg.Bot.Token)
	for {
		now := time.Now().In(loc)
		nextRun := time.Date(now.Year(), now.Month(), now.Day(), 20, 0, 0, 0, loc)
		if now.After(nextRun) {
			nextRun = nextRun.Add(24 * time.Hour)
		}

		duration := nextRun.Sub(now)
		log.Printf("Следующая рассылка через %v", duration)
		time.Sleep(duration)

		bestJoke, err := GetBestJoke()
		if err != nil {
			log.Printf("Ошибка получения лучшей шутки: %v", err)
			continue
		}
		text := "<b>Лучшая шутка дня:</b>\n\n" + bestJoke.Text
		msg := tgbotapi.NewMessageToChannel("@JokerScamCoin", text)
		msg.ParseMode = "HTML"
		msg.DisableWebPagePreview = true
		msID, _ := botAPI.Send(msg)

		var users []models.User
		if err := db.DB.Find(&users).Error; err != nil {
			log.Printf("Ошибка получения пользователей: %v", err)
			continue
		}

		for _, user := range users {
			go SendJoke(botAPI, user.TelegramID, msID.MessageID)
		}
	}
}

func SendJoke(botAPI *tgbotapi.BotAPI, TelegramID int64, ChanaleMsgID int) {
	msg := tgbotapi.NewMessage(TelegramID, "Новая шутка дня")
	var rows [][]tgbotapi.InlineKeyboardButton
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonURL("Посмотреть", fmt.Sprintf("https://t.me/JokerScamCoin/%d", ChanaleMsgID)),
	))
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
	botAPI.Send(msg)
}
