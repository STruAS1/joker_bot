package MenuJokes

import (
	"SHUTKANULbot/Utilities"
	"SHUTKANULbot/bot/context"
	"SHUTKANULbot/bot/handlers/start"
	"SHUTKANULbot/db"
	"SHUTKANULbot/db/models"
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
		Text = botCtx.Message.Text
	}

	Joke, exist := state.Data["NewJoke"].(NewJoke)
	if !exist {
		Joke = NewJoke{}
	}

	switch Joke.ActiveStep {
	case 0:
		if botCtx.CallbackQuery != nil && botCtx.CallbackQuery.Data == "NewJoke" {
			if len(jokeRequests) == 0 {
				log.Println("Ошибка: jokeRequests пуст")
				return
			}
			randomIndex := rand.Intn(len(jokeRequests))
			msgText := jokeRequests[randomIndex]

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

			msg := tgbotapi.NewMessage(botCtx.UserID, formatetText)
			msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
			msg.ParseMode = "HTML"
			botCtx.SendMessage(msg)

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
				UserID:         user.ID,
				Text:           Joke.Text,
				Evaluations:    0,
				AVGScore:       0,
				AuthorUserName: user.AuthorUserName,
				AnonymsMode:    user.AnonymsMode,
			}
			db.DB.Create(&NewJoke_db)
			delete(state.Data, "NewJoke")
			start.HandleStartCommand(botCtx)
			return
		}
	}
	state.Data["NewJoke"] = Joke
}

var jokeRequests = [50]string{
	"<b>🎤 Давай, не выёбывайся — шути!</b>",
	"<b>🔥 Въеби шутку!</b> Или <span class=\"tg-spoiler\">язык в жопе?</span>",
	"<i>😂 Где шутка,</i> <span class=\"tg-spoiler\">блядина?</span> Не тяни <span class=\"tg-spoiler\">хуй!</span>",
	"<b>🤡 Смеши, или в <span class=\"tg-spoiler\">хуй</span> дуть будешь?</b>",
	"<b>🤨 В ахуе...</b> Ты вообще шутить умеешь?",
	"<i>⚡ Шутку давай!</i> А то <span class=\"tg-spoiler\">въебаться можно.</span>",
	"<b>🎭 Ну и где твой</b> <span class=\"tg-spoiler\">блядский юмор?</span>",
	"<b>🚀 В <span class=\"tg-spoiler\">пизду</span> молчание!</b> Шути, пока можешь.",
	"<b>🎪 <span class=\"tg-spoiler\">Выеби</span> народ своим юмором!</b>",
	"<b>🕰️ Без <span class=\"tg-spoiler\">пизды</span>, жду шутку.</b>",
	"<b>😶 <span class=\"tg-spoiler\">Хули</span> ты в рот набрал?</b> Шути!",
	"<i>👀 В <span class=\"tg-spoiler\">хуй</span> не ставлю, пока не шутишь.</i>",
	"<b>🧠 Ты тупой или шутишь плохо?</b>",
	"<b>🍆 В рот тебе <span class=\"tg-spoiler\">хуй</span>, если не шутишь!</b>",
	"<i>🎙️ <span class=\"tg-spoiler\">Босый хуй</span> у микрофона...</i> Шути!",
	"<b>🧐 <span class=\"tg-spoiler\">Ахулиард</span> слов, но шутки нет?</b>",
	"<b>🤔 Давай, рассмеши народ, не позорься!</b>",
	"<i>🥶 Ты как <span class=\"tg-spoiler\">хуй</span> без причиндала.</i> Шути!",
	"<b>⚠️ Если не шутишь — ты <span class=\"tg-spoiler\">ебанат!</span></b>",
	"<b>📢 <span class=\"tg-spoiler\">Блядь</span>, где твой юмор?</b> Или ты просто <span class=\"tg-spoiler\">долбоёб?</span>",
	"<b>🛑 <span class=\"tg-spoiler\">Архипиздрит!</span> Шути, или <span class=\"tg-spoiler\">нахуй</span> иди.</b>",
	"<b>🎬 Смеши, или в <span class=\"tg-spoiler\">хуй</span> дуть будешь?</b>",
	"<i>🥶 Ты тут чё, <span class=\"tg-spoiler\">пизду</span> морозишь?</i> Шути!",
	"<b>🧨 Где шутка,</b> <span class=\"tg-spoiler\">хуесос?</span> Или в <span class=\"tg-spoiler\">ебло</span> дать?",
	"<b>🥴 Ну ты и <span class=\"tg-spoiler\">блядун...</span></b> Шути давай!",
	"<b>😡 <span class=\"tg-spoiler\">Блядство ебаное!</span></b> Шути быстрее!",
	"<i>🥶 В <span class=\"tg-spoiler\">хуй</span> дышит, но не шутит...</i>",
	"<b>💩 В <span class=\"tg-spoiler\">пизду</span> таких юмористов.</b> Шути уже!",
	"<b>🚽 Тебе в рот <span class=\"tg-spoiler\">нассать</span> или шутку скажешь?</b>",
	"<b>🎭 Ты <span class=\"tg-spoiler\">долбоёб</span> или юморист?</b> Шути давай!",
	"<b>📉 Пока не смешно...</b> Исправляйся!",
	"<b>🚪 Сюда <span class=\"tg-spoiler\">пизду</span> неси, а не молчание.</b>",
	"<b>🎪 Ты, <span class=\"tg-spoiler\">бля,</span> цирк забыл?</b> Шути давай!",
	"<b>🎲 В <span class=\"tg-spoiler\">рот ебаться</span> или шутить?</b> Выбирай.",
	"<b>📣 <span class=\"tg-spoiler\">Босый хуй</span> у сцены...</b> Шути, братан!",
	"<b>🖕 <span class=\"tg-spoiler\">Ебало</span> завали и шути!</b>",
	"<b>🌪️ Ты как <span class=\"tg-spoiler\">хуй</span> на ветру.</b> Шути или уйди!",
	"<b>🔥 Вся <span class=\"tg-spoiler\">пизда</span> в огне, а ты не шутишь?</b>",
	"<b>📢 Ты в <span class=\"tg-spoiler\">хуй</span> не всрался без шутки.</b>",
	"<b>📝 Пиши, <span class=\"tg-spoiler\">блядь</span>, или ливай <span class=\"tg-spoiler\">нахуй.</span></b>",
	"<b>📌 Я тут не просто так. Давай шутку!</b>",
	"<b>👀 Смотри, <span class=\"tg-spoiler\">долбоёб!</span> Шутка, или в <span class=\"tg-spoiler\">рот нассу.</span></b>",
	"<b>🎭 Не знаешь шуток? Ты <span class=\"tg-spoiler\">уебан.</span></b>",
	"<b>🛑 Ты либо шутишь, либо <span class=\"tg-spoiler\">уёбывай!</span></b>",
	"<b>🔊 <span class=\"tg-spoiler\">Блядь</span>, где твоя <span class=\"tg-spoiler\">сука шутка?</span></b>",
	"<b>🌊 Давай шутку, пока не <span class=\"tg-spoiler\">обоссался.</span></b>",
	"<b>🔥 Вся <span class=\"tg-spoiler\">пизда</span> в огне, а ты не шутишь?</b>",
	"<b>📉 Пока не смешно...</b> Исправляйся!",
	"<b>🫠 Ты тут просто так, <span class=\"tg-spoiler\">хуесос?</span> Или шутишь?</b>",
	"<b>🚀 Не молчи, ты не <span class=\"tg-spoiler\">батя.</span> Шути!</b>",
}
