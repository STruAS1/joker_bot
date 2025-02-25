package MenuJokes

import (
	"SHUTKANULbot/Utilities"
	"SHUTKANULbot/bot/context"
	"SHUTKANULbot/db"
	"SHUTKANULbot/db/models"
	"fmt"
	"math/rand"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type JokesPages struct {
	CurrentPage uint
	CountOFpage uint
	Pages       [][]Joke
}

type Joke struct {
	ID                 uint
	ShortName          string
	Text               string
	CountOfEvaluations uint
	AVGScore           uint
}

func HandleMyJokes(botCtx *context.BotContext) {
	state := context.GetUserState(botCtx)
	context.UpdateUserLevel(botCtx, 3)

	_JokesPages, exist := state.Data["JokesPages"].(JokesPages)
	if !exist {
		var user models.User
		db.DB.Where(&models.User{TelegramID: botCtx.UserID}).First(&user)

		var jokes []models.Jokes
		db.DB.Where(&models.Jokes{UserID: user.ID}).Order("id DESC").Limit(50).Find(&jokes)
		if len(jokes) == 0 {
			return
		}

		_JokesPages.Pages = make([][]Joke, 0, 10)

		for i, joke := range jokes {
			ShortName := []rune(Utilities.RemoveHTMLTags(joke.Text))
			if len(ShortName) > 20 {
				ShortName = ShortName[:20]
			}
			j := Joke{
				ID:                 joke.ID,
				ShortName:          string(ShortName),
				Text:               joke.Text,
				CountOfEvaluations: joke.Evaluations,
				AVGScore:           joke.AVGScore,
			}

			pageIndex := i / 5
			if pageIndex >= len(_JokesPages.Pages) {
				_JokesPages.Pages = append(_JokesPages.Pages, make([]Joke, 0, 5))
			}

			_JokesPages.Pages[pageIndex] = append(_JokesPages.Pages[pageIndex], j)
		}

		_JokesPages.CountOFpage = uint(len(_JokesPages.Pages))

	}
	if botCtx.CallbackQuery.Data == "page_next" && _JokesPages.CurrentPage+1 < _JokesPages.CountOFpage {
		_JokesPages.CurrentPage++
	}
	if botCtx.CallbackQuery.Data == "page_Prev" && _JokesPages.CurrentPage > 0 {
		_JokesPages.CurrentPage--
	}
	var rows [][]tgbotapi.InlineKeyboardButton
	for i, Joke := range _JokesPages.Pages[_JokesPages.CurrentPage] {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(Joke.ShortName, fmt.Sprintf("Joke_%d_%d", _JokesPages.CurrentPage, i))))
	}
	var showPrev, showNext bool
	if _JokesPages.CountOFpage > 0 {
		showPrev = _JokesPages.CurrentPage > 0
		showNext = _JokesPages.CurrentPage+1 < _JokesPages.CountOFpage
	}
	var row []tgbotapi.InlineKeyboardButton
	if showPrev {
		row = append(row, tgbotapi.NewInlineKeyboardButtonData("◀️", "page_Prev"))
	}
	if showNext {
		row = append(row, tgbotapi.NewInlineKeyboardButtonData("▶️", "page_next"))
	}
	if len(row) != 0 {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(row...))
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("🚀 Главное меню", "back")))
	randomIndex := rand.Intn(len(jokeMenuTitles))
	msg := tgbotapi.NewEditMessageTextAndMarkup(
		botCtx.UserID,
		state.MessageID,
		jokeMenuTitles[randomIndex],
		tgbotapi.NewInlineKeyboardMarkup(rows...),
	)
	msg.ParseMode = "HTML"
	state.Data["JokesPages"] = _JokesPages
	if _, err := botCtx.Ctx.BotAPI.Send(msg); err != nil {
		fmt.Println(err)
	}
}

func HandleMyJokeViewer(botCtx *context.BotContext, page uint8, Index uint8) {
	state := context.GetUserState(botCtx)
	context.UpdateUserLevel(botCtx, 4)
	_JokesPages, exist := state.Data["JokesPages"].(JokesPages)
	if !exist {
		HandleMyJokes(botCtx)
	}
	joke := _JokesPages.Pages[page][Index]
	text := fmt.Sprintf("🃏 <b>Шутка #%d:</b>\n%s \n\n\n✦──────────✦ \n<b>👀Просмотров:</b>  <code>%s</code>\n<b>⭐️Оценка:</b> <code>%s</code>\n✦──────────✦", joke.ID, joke.Text, Utilities.ConvertToFancyString(int(joke.CountOfEvaluations)), Utilities.ConvertToFancyStringFloat(fmt.Sprintf("%f", float64(joke.AVGScore)/20)))
	var rows [][]tgbotapi.InlineKeyboardButton
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("📋 Обратно к списку", "back")))
	text += fmt.Sprintf("\n\n<i><b>Поделись с корешом:</b> \n<code>https://t.me/JOKER8BOT?start=joke_%d</code></i>", joke.ID)
	msg := tgbotapi.NewEditMessageTextAndMarkup(
		botCtx.UserID,
		state.MessageID,
		text,
		tgbotapi.NewInlineKeyboardMarkup(rows...),
	)
	msg.ParseMode = "HTML"
	botCtx.Ctx.BotAPI.Send(msg)
}

var jokeMenuTitles = [29]string{
	"<b>😂 Юмор, от которого батя сказал <i>‘не плохо’</i> и 🏃💨 <span class=\"tg-spoiler\">съебался</span> за хлебом</b>",
	"<b><u>📜 Каталог юмора</u></b>. <i>Без возврата и обмена.</i>",
	"<b>🥊 Шутки, после которых можно огрести <span class=\"tg-spoiler\">пизды</span>, но оно того стоит</b>",
	"<b>🏆 За этот юмор меня либо <i>посадят</i>, либо <u>сделают легендой</u></b>",
	"<b>💀 Охуеть, ты правда сюда зашёл? <i>Ну теперь держись.</i></b>",
	"<b>🐱 Сборник шуток, от которых даже мой кот <u>смотрит на меня с осуждением</u></b>",
	"<b>🧊 Тут даже холодильник <span class=\"tg-spoiler\">охуел</span>, <i>а он видел, как я жру в 3 ночи</i></b>",
	"<b>🤡 После этих шуток <i>либо смеёшься</i>, либо <u><span class=\"tg-spoiler\">уходишь нахуй</span></u></b>",
	"<b>🩸 Коллекция мемов, <i>за которые можно выхватить <span class=\"tg-spoiler\">в ебало</span></i></b>",
	"<b>🔀 Здесь уровень комедии от <u>‘гениально’</u> до <span class=\"tg-spoiler\">‘ну ты и хуесос’</span></b>",
	"<b>📝 Шутки, которые могли бы остаться в <i>черновиках</i>, но мне <span class=\"tg-spoiler\">похуй</span></b>",
	"<b>🥃 Юмор с лёгким привкусом <i><span class=\"tg-spoiler\">‘блядь, нахуя я это читаю?’</span></i></b>",
	"<b>🍺 Комедийный архив, после которого ты точно <u>захочешь <span class=\"tg-spoiler\">нажраться</span></u></b>",
	"<b>⚠️ Если не смеёшься — <i>у тебя либо нет души, либо ты просто <span class=\"tg-spoiler\">долбоёб</span></i></b>",
	"<b>📉 Шутки, за которые меня либо уволят, либо <i>повысят</i></b>",
	"<b>🛒 Юмор, от которого даже кассирша в Пятёрочке <u><span class=\"tg-spoiler\">охуеет</span></u></b>",
	"<b>🏥 Если после этого тебя не заберут санитары — <i>ты, сука, <span class=\"tg-spoiler\">легенда</span></i></b>",
	"<b>👊 Шутки уровня <span class=\"tg-spoiler\">‘ебало сломай, но не смейся’</span></b>",
	"<b>📖 Каталог фраз, за которые мне <i>могут прописать <span class=\"tg-spoiler\">леща</span></i></b>",
	"<b>🎭 Если ты читаешь это дерьмо, то <span class=\"tg-spoiler\">либо у тебя есть вкус, либо ты конченый</span></b>",
	"<b>🎯 Смешно? <i><span class=\"tg-spoiler\">Хуй знает</span></i>, но читать будешь.</b>",
	"<b>🎤 После этих шуток ты либо <i>комик</i>, либо <u><span class=\"tg-spoiler\">хуесос</span></u></b>",
	"<b>💼 Шутки, после которых можно <span class=\"tg-spoiler\">бросить работу</span> и вообще не жалеть</b>",
	"<b>💻 Сборник юмора, от которого даже <i>интернет-тролли</i> <span class=\"tg-spoiler\">охуевают</span></b>",
	"<b>💥 Юмор, после которого тебя <span class=\"tg-spoiler\">либо зауважают</span>, либо <u><span class=\"tg-spoiler\">отпиздят</span></u></b>",
	"<b>🤖 Здесь даже нейросеть зависает, <i>пытаясь понять, что это за <span class=\"tg-spoiler\">хуйня</span></i></b>",
	"<b>🚪 Шутки, за которые батя <u>вернулся</u>, но потом снова <i><span class=\"tg-spoiler\">съебался</span></i></b>",
	"<b>🔮 Если ты это читаешь, значит у тебя <i>либо железные нервы, либо <span class=\"tg-spoiler\">нехуй делать</span></i></b>",
	"<b>🎢 Место, где <u>ржут</u> даже те, кто по жизни <i><span class=\"tg-spoiler\">вечно заебался</span></i></b>",
}
