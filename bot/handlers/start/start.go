package start

import (
	"SHUTKANULbot/bot/context"
	"SHUTKANULbot/db"
	"SHUTKANULbot/db/models"
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func HandleStartCommand(botCtx *context.BotContext) {
	state := context.GetUserState(botCtx)
	context.UpdateUserLevel(botCtx, 0)
	var user models.User
	result := db.DB.Where(&models.User{TelegramID: botCtx.UserID}).First(&user)
	if result.Error != nil {
		if botCtx.Message != nil {
			NewUser := models.User{
				Username:   botCtx.Message.From.UserName,
				TelegramID: botCtx.Message.From.ID,
				FirstName:  botCtx.Message.From.FirstName,
				LastName:   botCtx.Message.From.LastName,
			}
			if err := db.DB.Create(&NewUser).Error; err != nil {
				log.Printf("Ошибка при сохранении пользователя: %v", err)
				return
			}
			db.DB.Where(&models.User{TelegramID: botCtx.UserID}).First(&user)
		} else {
			return
		}
	}
	var TotalJokes int64
	var TotalGetEvaluations int
	var AVGscore float64
	db.DB.Model(&models.Jokes{}).Where(&models.Jokes{UserID: user.ID}).Count(&TotalJokes)
	db.DB.Model(&models.Jokes{}).Where(&models.Jokes{UserID: user.ID}).Select("SUM(evaluations)").Scan(&TotalGetEvaluations)
	db.DB.Model(&models.Jokes{}).Where(&models.Jokes{UserID: user.ID}).Select("AVG(avg_score)").Scan(&AVGscore)
	var MainText string
	MainText = fmt.Sprintf("<b>%s %s</b>", user.FirstName, user.LastName)
	MainText += fmt.Sprintf("\n\n<b>💰Баланс: <i>%d</i></b>", user.Balance)
	MainText += "\n\n📊<b>Статистика: </b>"
	MainText += fmt.Sprintf("\n<blockquote><i><b>Шуток опубликовано: %d\nОценок получено: %d \nСредняя оценка: %f</b></i></blockquote>", TotalJokes, TotalGetEvaluations, AVGscore/100)
	var rows [][]tgbotapi.InlineKeyboardButton
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Смотреть шутки", "ViewJokes"), tgbotapi.NewInlineKeyboardButtonData("Шуткануть", "NweJoke")))
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Мои шутки", "MyJokes")))
	if state.MessageID == 0 {
		msg := tgbotapi.NewMessage(botCtx.UserID, MainText)
		msg.ParseMode = "HTML"
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
		botCtx.SendMessage(msg)
	} else {
		msg := tgbotapi.NewEditMessageTextAndMarkup(botCtx.UserID, state.MessageID, MainText, tgbotapi.NewInlineKeyboardMarkup(rows...))
		msg.ParseMode = "HTML"
		botCtx.Ctx.BotAPI.Send(msg)
	}
}
