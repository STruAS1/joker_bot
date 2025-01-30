package MenuJokes

import (
	"SHUTKANULbot/Utilities"
	"SHUTKANULbot/bot/context"
	"SHUTKANULbot/bot/handlers/start"
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

func handleLvl2(botCtx *context.BotContext) {
	if botCtx.CallbackQuery != nil {
		data := strings.Split(botCtx.CallbackQuery.Data, "_")
		switch data[0] {
		case "Evolution":
			if len(data) == 3 {
				jokeId, _ := strconv.ParseUint(data[2], 10, 0)
				evaluation, _ := strconv.ParseUint(data[1], 10, 8)
				Utilities.AddJokeEvaluation(botCtx.UserID, uint(jokeId), uint8(evaluation))
				alert := tgbotapi.NewCallbackWithAlert(botCtx.CallbackQuery.ID, "Вы оценили шутку")
				alert.ShowAlert = false
				botCtx.Ctx.BotAPI.Request(alert)
				HandleJokeViewer(botCtx)
			}
		case "Start":
			start.HandleStartCommand(botCtx)
		}
	}
}
