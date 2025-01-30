package MenuJokes

import (
	"SHUTKANULbot/Utilities"
	"SHUTKANULbot/bot/context"
	"SHUTKANULbot/bot/handlers/start"
	"SHUTKANULbot/db"
	"SHUTKANULbot/db/models"
	"fmt"
	"log"
	"math/rand"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type NewJoke struct {
	Text       string
	ActiveStep uint
}

func NewJokeHandle(botCtx *context.BotContext) {
	state := context.GetUserState(botCtx)
	context.UpdateUserLevel(botCtx, 1)

	var Text string
	if botCtx.Message != nil {
		if botCtx.Message.MessageID != 0 {
			deleteCFG := tgbotapi.DeleteMessageConfig{
				ChatID:    botCtx.UserID,
				MessageID: botCtx.Message.MessageID,
			}
			_, err := botCtx.Ctx.BotAPI.Send(deleteCFG)
			if err != nil {
				log.Printf("Ошибка при удалении сообщения: %v", err)
			}
		}
		Text = botCtx.Message.Text
	}

	Joke, exist := state.Data["NewJoke"].(NewJoke)
	if !exist {
		Joke = NewJoke{}
	}

	switch Joke.ActiveStep {
	case 0:
		if botCtx.CallbackQuery != nil && botCtx.CallbackQuery.Data == "NweJoke" {
			if len(jokeRequests) == 0 {
				log.Println("Ошибка: jokeRequests пуст")
				return
			}
			randomIndex := rand.Intn(len(jokeRequests))
			msgText := fmt.Sprintf("<i><span class=\"tg-spoiler\">%s</span></i>", jokeRequests[randomIndex])

			Joke.ActiveStep++
			var rows [][]tgbotapi.InlineKeyboardButton
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Отмена", "back")))

			if state.MessageID != 0 {
				msg := tgbotapi.NewEditMessageTextAndMarkup(botCtx.UserID, state.MessageID, msgText, tgbotapi.NewInlineKeyboardMarkup(rows...))
				msg.ParseMode = "HTML"
				botCtx.Ctx.BotAPI.Send(msg)
			} else {
				msg := tgbotapi.NewMessage(botCtx.UserID, msgText)
				msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
				msg.ParseMode = "HTML"
				botCtx.SendMessage(msg)
			}
		}
	case 1:
		if botCtx.Message != nil {
			formatetText := Utilities.ApplyFormatting(Text, botCtx.Message.Entities)
			Joke.Text = formatetText

			var rows [][]tgbotapi.InlineKeyboardButton
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Сохранить", "Save")))
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Отмена", "back")))

			if state.MessageID != 0 {
				msg := tgbotapi.NewEditMessageTextAndMarkup(botCtx.UserID, state.MessageID, formatetText, tgbotapi.NewInlineKeyboardMarkup(rows...))
				msg.ParseMode = "HTML"
				botCtx.Ctx.BotAPI.Send(msg)
			} else {
				msg := tgbotapi.NewMessage(botCtx.UserID, formatetText)
				msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
				msg.ParseMode = "HTML"
				botCtx.SendMessage(msg)
			}
			Joke.ActiveStep++
		}
	case 2:
		if botCtx.CallbackQuery != nil && botCtx.CallbackQuery.Data == "Save" {
			var user models.User
			result := db.DB.Where("telegram_id = ?", botCtx.UserID).First(&user)
			if result.Error != nil {
				log.Printf("Ошибка: пользователь %d не найден в базе", botCtx.UserID)
				return
			}

			NewJoke_db := models.Jokes{
				UserID:      user.ID,
				Text:        Joke.Text,
				Evaluations: 0,
				AVGScore:    0,
			}
			db.DB.Create(&NewJoke_db)

			delete(state.Data, "NewJoke")
			start.HandleStartCommand(botCtx)
			return
		}
	}
	state.Data["NewJoke"] = Joke
}

var jokeRequests = [88]string{
	"<b>Давай, не выёбывайся — шути!</b>",
	"<b>Въеби шутку!</b> Или язык в жопе?",
	"<i>Где шутка, блядина?</i> Не тяни хуй!",
	"<b>Смеши, или в хуй дуть будешь?</b>",
	"<b>В ахуе...</b> Ты вообще шутить умеешь?",
	"<i>Шутку давай!</i> А то въебаться можно.",
	"<b>Ну и где твой блядский юмор?</b>",
	"<b>В пизду молчание!</b> Шути, пока можешь.",
	"<b>Выеби народ своим юмором!</b>",
	"<b>Без пизды, жду шутку.</b>",
	"<b>Хули ты в рот набрал?</b> Шути!",
	"<i>В хуй не ставлю, пока не шутишь.</i>",
	"<b>Ты тупой или шутишь плохо?</b>",
	"<b>В рот тебе хуй, если не шутишь!</b>",
	"<i>Босый хуй у микрофона...</i> Шути!",
	"<b>Ахулиард слов, но шутки нет?</b>",
	"<b>Давай, рассмеши народ, не позорься!</b>",
	"<i>Ты как хуй без причиндала.</i> Шути!",
	"<b>Если не шутишь — ты ебанат!</b>",
	"<b>Блядь, где твой юмор?</b> Или ты просто долбоёб?",
	"<b>Архипиздрит! Шути, или нахуй иди.</b>",
	"<b>Смеши, или в хуй дуть будешь?</b>",
	"<i>Ты тут чё, пизду морозишь?</i> Шути!",
	"<b>Где шутка, хуесос?</b> Или в ебло дать?",
	"<b>Ну ты и блядун...</b> Шути давай!",
	"<b>Блядство ебаное!</b> Шути быстрее!",
	"<i>В хуй дышит, но не шутит...</i>",
	"<b>В пизду таких юмористов.</b> Шути уже!",
	"<b>Тебе в рот нассать или шутку скажешь?</b>",
	"<b>Ты долбоёб или юморист?</b> Шути давай!",
	"<b>Пока не смешно...</b> Исправляйся!",
	"<b>Сюда пизду неси, а не молчание.</b>",
	"<b>Ты, бля, цирк забыл?</b> Шути давай!",
	"<b>В рот ебаться или шутить?</b> Выбирай.",
	"<b>Босый хуй у сцены...</b> Шути, братан!",
	"<b>Ебало завали и шути!</b>",
	"<b>Не въёбывайся, а шути!</b>",
	"<b>Ты как хуй на ветру.</b> Шути или уйди!",
	"<b>В ахуе... Ты вообще шутить умеешь?</b>",
	"<i>Ну, блядёшка, жги шутками!</i>",
	"<b>Ты тут для шуток, а не для нытья!</b>",
	"<b>Ты в хуй не всрался без шутки.</b>",
	"<b>Ты долбоёб? Тогда хотя бы шути.</b>",
	"<b>Ты хоть раз смешил народ?</b> Давай проверим.",
	"<b>Давай, ёбнутый, рассмеши нас.</b>",
	"<b>Ну всё, твой хуй на сцене.</b> Шути!",
	"<b>Дай нормальную шутку, не обосрись.</b>",
	"<b>Ну, не пизди, а шути!</b>",
	"<b>Блядь, где твоя сука шутка?</b>",
	"<b>Вся пизда в огне, а ты не шутишь?</b>",
	"<b>Давай, пока не въебали, шути!</b>",
	"<b>Ты тут просто так, хуесос?</b> Или шутишь?",
	"<b>Ты либо шутишь, либо уёбывай!</b>",
	"<b>Не выебывайся, пиши шутку.</b>",
	"<b>Ты долбоёб или стендапер?</b> Проверим.",
	"<b>Блядь, я жду! Шути!</b>",
	"<b>Ну всё, твой хуй на сцене. Шути!</b>",
	"<b>Давай, ёбнутый, рассмеши нас.</b>",
	"<b>Ты хоть раз смешил народ?</b> Давай проверим.",
	"<b>Ну, не пизди, а шути!</b>",
	"<b>Вся пизда в огне, а ты не шутишь?</b>",
	"<b>Ты в хуй не всрался без шутки.</b>",
	"<b>Пиши, блядь, или ливай нахуй.</b>",
	"<b>Я тут не просто так. Давай шутку!</b>",
	"<b>Смотри, долбоёб! Шутка, или в рот нассу.</b>",
	"<b>Не знаешь шуток? Ты уебан.</b>",
	"<b>Ты либо шутишь, либо уёбывай!</b>",
	"<b>Блядь, где твоя сука шутка?</b>",
	"<b>Давай шутку, пока не обоссался.</b>",
	"<b>Вся пизда в огне, а ты не шутишь?</b>",
	"<b>Пока не смешно...</b> Исправляйся!",
	"<b>Ты тут просто так, хуесос?</b> Или шутишь?",
	"<b>Ну ты и блядун...</b> Шути давай!",
	"<b>Не молчи, ты не батя. Шути!</b>",
	"<b>Давай, пока не въебали, шути!</b>",
	"<b>Ты в рот нассать хочешь или шутить?</b>",
	"<b>Шути, или твоё ебало в хуй.</b>",
	"<b>Ты тупой? Тогда хотя бы шути.</b>",
	"<b>Ты долбоёб или стендапер?</b> Проверим.",
	"<b>Ну, блядёшка, жги шутками!</b>",
	"<b>Не въёбывайся, а шути!</b>",
	"<b>Ну всё, твой хуй на сцене.</b> Шути!",
	"<b>Пока не смешно...</b> Исправляйся!",
	"<b>Ты тут просто так, хуесос?</b> Или шутишь?",
	"<b>Ты в ахуе или в хуй дышишь?</b>",
	"<b>Ты здесь? Тогда давай уже шутку!</b>",
	"<b>Всё в ажуре, а хуй на абажуре!</b>",
	"<b>Ты тут чё, бля? Шути давай!</b>",
}
