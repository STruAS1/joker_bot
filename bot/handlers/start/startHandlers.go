package start

import (
	"SHUTKANULbot/Utilities"
	contextBot "SHUTKANULbot/bot/context"
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func Handle(botCtx *contextBot.BotContext) {
	state := contextBot.GetUserState(botCtx)
	switch state.Level {
	case 1:
		HandleLVL1(botCtx)
	case 2:
		handleLvl2(botCtx)
	case 3:
		handleLvl3(botCtx)
	case 4:
		handleLvl4(botCtx)
	case 5:
		handleLvl5(botCtx)
	}

}

func HandleLVL1(botCtx *contextBot.BotContext) {
	if botCtx.CallbackQuery != nil {
		switch botCtx.CallbackQuery.Data {
		case "Docs":
			HandleDocs(botCtx)
		default:
			state := contextBot.GetUserState(botCtx)
			delete(state.Data, "DocsActiveStep")
			DeleteMessageConfig := tgbotapi.DeleteMessageConfig{
				ChatID:    botCtx.UserID,
				MessageID: state.MessageID,
			}
			botCtx.Ctx.BotAPI.Send(DeleteMessageConfig)
			state.MessageID = 0
			HandleStartCommand(botCtx)
		}
	}
}

func handleLvl2(botCtx *contextBot.BotContext) {

	state := contextBot.GetUserState(botCtx)
	if botCtx.CallbackQuery != nil {
		switch botCtx.CallbackQuery.Data {
		case "back":
			cancel, exist := state.Data["ChancelTransaction"].(context.CancelFunc)
			if exist {
				cancel()
				state.Data["SendErr"] = false
				delete(state.Data, "ChancelTransaction")
			}
			delete(state.Data, "WithdrawActiveStep")
			HandleSettings(botCtx)
		default:
			HandleWithdraw(botCtx)
		}
	} else if botCtx.Message != nil {
		HandleWithdraw(botCtx)
	}
}

func handleLvl4(botCtx *contextBot.BotContext) {
	state := contextBot.GetUserState(botCtx)
	if botCtx.CallbackQuery != nil {
		switch botCtx.CallbackQuery.Data {
		case "back":
			delete(state.Data, "SetAuthorData")
			HandleSettings(botCtx)
		default:
			HandleSetAuthor(botCtx)
		}
	} else if botCtx.Message != nil {
		HandleSetAuthor(botCtx)
	}
}

func handleLvl3(botCtx *contextBot.BotContext) {
	state := contextBot.GetUserState(botCtx)
	if botCtx.CallbackQuery != nil {
		switch botCtx.CallbackQuery.Data {
		case "back":
			cancel, exist := state.Data["ChancelConnect"].(context.CancelFunc)
			if exist {
				cancel()
				state.Data["SendErr"] = false
				delete(state.Data, "ChancelConnect")
			}
			deleteCfg := tgbotapi.DeleteMessageConfig{
				ChatID:    botCtx.UserID,
				MessageID: state.MessageID,
			}
			state.MessageID = 0
			botCtx.Ctx.BotAPI.Send(deleteCfg)
			HandleSettings(botCtx)
		}
	}
}

func handleLvl5(botCtx *contextBot.BotContext) {
	if botCtx.CallbackQuery != nil {
		data := strings.Split(botCtx.CallbackQuery.Data, "_")
		switch data[0] {
		case "Evolution":
			if len(data) == 4 {
				jokeId, _ := strconv.ParseUint(data[2], 10, 0)
				evaluation, _ := strconv.ParseUint(data[1], 10, 0)
				Utilities.AddJokeEvaluation(botCtx.UserID, uint(jokeId), uint(evaluation))
				alert := tgbotapi.NewCallbackWithAlert(botCtx.CallbackQuery.ID, rateMessages[evaluation][rand.Intn(len(rateMessages[evaluation]))])
				alert.ShowAlert = false
				botCtx.Ctx.BotAPI.Request(alert)
				fmt.Print(data[3])
				if data[3] == "true" {
					HandleDocs(botCtx)
				} else {
					HandleStartCommand(botCtx)
				}
			}
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
