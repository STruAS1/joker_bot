package start

import (
	"SHUTKANULbot/Utilities"
	"SHUTKANULbot/bot/context"
	"SHUTKANULbot/db"
	"SHUTKANULbot/db/models"
	"fmt"
	"log"
	"strconv"
	"strings"

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
			arg := botCtx.Message.CommandArguments()
			if arg != "" {
				args := strings.Split(arg, "_")
				if len(args) == 2 {
					switch args[0] {
					case "joke":
						jokeId, err := strconv.ParseUint(args[1], 10, 0)
						if err != nil {
							return
						}
						HandleJokeViewerReply(botCtx, uint(jokeId), true, user.ID)
						return
					}
				}
			}
			HandleDocs(botCtx)
			return
		} else {
			return
		}
	}
	if botCtx.Message != nil {
		arg := botCtx.Message.CommandArguments()
		if arg != "" {
			args := strings.Split(arg, "_")
			if len(args) == 2 {
				switch args[0] {
				case "joke":
					jokeId, err := strconv.ParseUint(args[1], 10, 0)
					if err != nil {
						return
					}
					HandleJokeViewerReply(botCtx, uint(jokeId), false, user.ID)
					return
				}
			}
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
		tgbotapi.NewInlineKeyboardButtonURL("🚀 Топ угара", "https://t.me/JokerScamCoin"),
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

func HandleJokeViewerReply(botCtx *context.BotContext, jokeID uint, IsFirst bool, userId uint) {
	state := context.GetUserState(botCtx)
	context.UpdateUserLevel(botCtx, 5)
	var rows [][]tgbotapi.InlineKeyboardButton
	var row []tgbotapi.InlineKeyboardButton
	if Utilities.HasUserEvaluatedJoke(userId, jokeID) {
		if IsFirst {
			HandleDocs(botCtx)
			return
		} else {
			msg := tgbotapi.NewMessage(botCtx.UserID, "Вы уже оценивали эту шутку!")
			msg.ParseMode = "HTML"
			msg.DisableWebPagePreview = true
			botCtx.SendMessage(msg)
			return
		}
	}
	Joke := Utilities.GetJokeByID(jokeID)
	for i := range 5 {
		row = append(row, tgbotapi.NewInlineKeyboardButtonData(fmt.Sprint(i+1), fmt.Sprintf("Evolution_%d_%d_%t", i+1, Joke.ID, IsFirst)))
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(row...))
	if !Joke.AnonymsMode && Joke.Author != "" {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonURL("Автор", fmt.Sprintf("https://t.me/%s", strings.TrimPrefix(Joke.Author, "@")))))
	}
	if state.MessageID != 0 {
		msg := tgbotapi.NewEditMessageTextAndMarkup(
			botCtx.UserID,
			state.MessageID,
			Joke.Text,
			tgbotapi.NewInlineKeyboardMarkup(rows...),
		)
		msg.ParseMode = "HTML"
		botCtx.Ctx.BotAPI.Send(msg)
	} else {
		msg := tgbotapi.NewMessage(
			botCtx.UserID,
			Joke.Text,
		)
		msg.ParseMode = "HTML"
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
		botCtx.Ctx.BotAPI.Send(msg)
	}
}
