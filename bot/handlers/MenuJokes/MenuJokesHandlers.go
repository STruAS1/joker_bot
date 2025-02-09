package MenuJokes

import (
	"SHUTKANULbot/Utilities"
	"SHUTKANULbot/bot/context"
	"SHUTKANULbot/bot/handlers/start"
	"math/rand"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func Handle(botCtx *context.BotContext) {
	state := context.GetUserState(botCtx)
	switch state.Level {
	case 1:
		handleLvl1(botCtx)
	case 2:
		handleLvl2(botCtx)
	case 3:
		handleLvl3(botCtx)
	case 4:
		handleLvl4(botCtx)
	}

}

func handleLvl1(botCtx *context.BotContext) {
	state := context.GetUserState(botCtx)
	if botCtx.CallbackQuery != nil {
		switch botCtx.CallbackQuery.Data {
		case "back":
			delete(state.Data, "NewJoke")
			start.HandleStartCommand(botCtx)
		default:
			NewJokeHandle(botCtx)
		}
	} else if botCtx.Message != nil {
		NewJokeHandle(botCtx)
	}
}

func handleLvl3(botCtx *context.BotContext) {
	state := context.GetUserState(botCtx)
	if botCtx.CallbackQuery != nil {
		data := strings.Split(botCtx.CallbackQuery.Data, "_")
		switch data[0] {
		case "back":
			delete(state.Data, "JokesPages")
			start.HandleStartCommand(botCtx)
		case "Joke":
			jokeId, _ := strconv.ParseUint(data[2], 10, 8)
			pageId, _ := strconv.ParseUint(data[1], 10, 8)
			HandleMyJokeViewer(botCtx, uint8(pageId), uint8(jokeId))
		default:
			HandleMyJokes(botCtx)
		}
	}
}

func handleLvl4(botCtx *context.BotContext) {
	if botCtx.CallbackQuery != nil && botCtx.CallbackQuery.Data == "back" {
		HandleMyJokes(botCtx)
	}
}

func handleLvl2(botCtx *context.BotContext) {
	if botCtx.CallbackQuery != nil {
		data := strings.Split(botCtx.CallbackQuery.Data, "_")
		switch data[0] {
		case "Evolution":
			if len(data) == 3 {
				jokeId, _ := strconv.ParseUint(data[2], 10, 0)
				evaluation, _ := strconv.ParseUint(data[1], 10, 0)
				Utilities.AddJokeEvaluation(botCtx.UserID, uint(jokeId), uint(evaluation))
				alert := tgbotapi.NewCallbackWithAlert(botCtx.CallbackQuery.ID, rateMessages[evaluation][rand.Intn(len(rateMessages[evaluation]))])
				alert.ShowAlert = false
				botCtx.Ctx.BotAPI.Request(alert)
				HandleJokeViewer(botCtx)
			}
		case "Start":
			start.HandleStartCommand(botCtx)
		}
	}
}

var rateMessages = map[uint64][]string{
	1: {
		"Ну и кринж…",
		"👎 Это было больно.",
		"Шутка хуже, чем твой батя ушёл за хлебом.",
		"Минус уши, минус настроение.",
		"Это точно не смешно.",
		"Слабовато…",
		"Шутка упала лицом в грязь.",
		"Автор, ты это серьёзно?",
		"Брат, так не шутят.",
		"Дно пробито.",
	},
	2: {
		"Ну... могло быть и хуже.",
		"🙄 Не фонтан, но хоть не кринж.",
		"Шутка на троечку… а нет, на двоечку.",
		"Слегка улыбнулся, но так, из вежливости.",
		"Ну да… ну да…",
		"Почти получилось, старайся.",
		"Бро, ты можешь лучше!",
		"На грани фола, но ещё держится.",
		"Не самый худший панч, но далеко не лучший.",
		"Было бы круто, но не зашло.",
	},
	3: {
		"Средненько, но норм.",
		"👌 Нормально, но без огонька.",
		"Ну, бывало и лучше.",
		"Это, конечно, не гениально, но терпимо.",
		"Шутка стабильно средняя.",
		"Что-то в этом есть.",
		"Чуть больше харизмы — и будет топ.",
		"Не плохо, но и не хорошо.",
		"Чистый нейтрал, не обидел, но и не удивил.",
		"Если бы шутки были пельменями, это был бы ОК пельмень.",
	},
	4: {
		"🔥 Хорош! Чувствуется вайб.",
		"Мне нравится твой стиль!",
		"Это почти топ, но чутка не дотянул.",
		"Уже близко к хайпу!",
		"Крепкий панч, держи пятюню ✋",
		"Круто, но можно было ещё чуть жирнее.",
		"Ты на верном пути!",
		"Шутка с претензией на классику.",
		"Это уже уровень, респект!",
		"Достойно, но можно ещё сильнее!",
	},
	5: {
		"💀 Я УМЕР!",
		"😂 Это лучшее, что я слышал!",
		"КТО-ТО ВЫЗЫВАЛ УБИЙЦУ СМЕХА?",
		"🔥 Это разъёб! В топ срочно!",
		"Ты — гений юмора.",
		"Это шедевр, залетай в топ!",
		"Я орнул так, что соседи вызвали полицию.",
		"Дайте этому человеку микрофон, срочно!",
		"💯 Достойно статуса легенды!",
		"Это панчлайн, который войдёт в историю!",
	},
}
