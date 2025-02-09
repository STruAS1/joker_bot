package start

import (
	"SHUTKANULbot/Utilities"
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
			HandleDocs(botCtx)
			return
		} else {
			return
		}
	}

	var TotalJokes int64
	var TotalGetEvaluations int
	var AVGscore float64
	var JokesViews int64
	db.DB.Model(&models.Jokes{}).Where(&models.Jokes{UserID: user.ID}).Count(&TotalJokes)
	db.DB.Model(&models.JokesEvaluations{}).Where(&models.JokesEvaluations{UserID: user.ID}).Count(&JokesViews)
	db.DB.Model(&models.Jokes{}).Where(&models.Jokes{UserID: user.ID}).Select("SUM(evaluations)").Scan(&TotalGetEvaluations)
	db.DB.Model(&models.Jokes{}).Where(&models.Jokes{UserID: user.ID}).Select("AVG(avg_score)").Scan(&AVGscore)

	MainText := "<b>🤖 Шутканул? Добро пожаловать!</b>\n\n"
	MainText += "╭━━━━━━━━━🎭\n"
	MainText += fmt.Sprintf("┃  👤 <b>%s %s</b>\n", user.FirstName, user.LastName)
	MainText += fmt.Sprintf("┃  💰 <b>Баланс: <code>%s</code> <a href='https://tonviewer.com/EQBCo9q5lgaBLiTPkZwFZ1dlZ2l5Gg76bT5_eTml0fzZyTLs'>$JOKER</a></b>\n", Utilities.ConvertToFancyStringFloat(fmt.Sprintf("%f", float64((uint64(user.Balance/1_000_000)))/1000)))
	MainText += "┃━━━━━━━━━🎭\n"
	MainText += fmt.Sprintf("┃  ✍️ <b>Шутканул:</b> <code>%s</code>\n", Utilities.ConvertToFancyString(int(TotalJokes)))
	MainText += fmt.Sprintf("┃  👀 <b>Просмотрел шуток:</b> <code>%s</code>\n", Utilities.ConvertToFancyString(int(JokesViews)))
	MainText += fmt.Sprintf("┃  ⭐ <b>Оценок получил:</b> <code>%s</code>\n", Utilities.ConvertToFancyString(TotalGetEvaluations))
	MainText += fmt.Sprintf("┃  📈 <b>Средняя оценка:</b> <code>%.2f</code>\n", AVGscore/20)
	MainText += "╰━━━━━━━━━🎭\n"

	var rows [][]tgbotapi.InlineKeyboardButton
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("📜 Лента шуток", "ViewJokes"),
		tgbotapi.NewInlineKeyboardButtonData("🔥 Шуткануть", "NewJoke"),
	))

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("📂 Мои панчи", "MyJokes"),
		tgbotapi.NewInlineKeyboardButtonData("🛠 Тюнинг", "Settings"),
	))

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("📚 Обучение", "Docs"),
		tgbotapi.NewInlineKeyboardButtonURL("🚀 Топ угара", "https://t.me/YOUR_CHANNEL"),
	))

	if state.MessageID == 0 {
		msg := tgbotapi.NewMessage(botCtx.UserID, MainText)
		msg.ParseMode = "HTML"
		msg.DisableWebPagePreview = true
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
		botCtx.SendMessage(msg)
	} else {
		msg := tgbotapi.NewEditMessageTextAndMarkup(botCtx.UserID, state.MessageID, MainText, tgbotapi.NewInlineKeyboardMarkup(rows...))
		msg.DisableWebPagePreview = true
		msg.ParseMode = "HTML"
		botCtx.Ctx.BotAPI.Send(msg)
	}
}
