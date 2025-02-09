package start

import (
	contextBot "SHUTKANULbot/bot/context"
	"context"

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
