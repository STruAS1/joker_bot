package handlers

import (
	// "SHUTKANULbot/bot/handlers/MenuJokes"
	TonConnectCallback "SHUTKANULbot/TonConnectCallBack"
	"SHUTKANULbot/Utilities"
	"SHUTKANULbot/bot/context"
	"SHUTKANULbot/bot/handlers/MenuJokes"
	"SHUTKANULbot/bot/handlers/start"
	"SHUTKANULbot/db"
	"SHUTKANULbot/db/models"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var nameHandlers = map[string]func(*context.BotContext){
	"start":    start.Handle,
	"JokeMenu": MenuJokes.Handle,
}

func HandleUpdate(botCtx *context.BotContext) {
	state := context.GetUserState(botCtx)
	if botCtx.Message != nil {
		switch botCtx.Message.Command() {
		case "start":
			state.MessageID = 0
			context.ClearAllUserData(botCtx)
			start.HandleStartCommand(botCtx)
			return
		}
	}
	if state.Level != 0 {
		if handler, exists := nameHandlers[state.Name]; exists {
			handler(botCtx)
		} else {
			start.Handle(botCtx)
		}
	} else {
		if botCtx.CallbackQuery != nil {
			switch strings.Split(botCtx.CallbackQuery.Data, "_")[0] {
			case "StartMenu":
				context.UpdateUserName(botCtx, "start")
				start.HandleStartCommand(botCtx)
			case "Docs":
				context.UpdateUserName(botCtx, "start")
				start.HandleDocs(botCtx)
			case "ConnectWallet":
				context.UpdateUserName(botCtx, "start")
				start.HandleTonConnect(botCtx)
			case "Settings":
				context.UpdateUserName(botCtx, "start")
				start.HandleSettings(botCtx)
			case "Withdraw":
				context.UpdateUserName(botCtx, "start")
				start.HandleWithdraw(botCtx)
			case "DisconnectWallet":
				context.UpdateUserName(botCtx, "start")
				TonConnectCallback.Disconnect(botCtx.UserID)
				start.HandleSettings(botCtx)
			case "SetAuthor":
				context.UpdateUserName(botCtx, "start")
				start.HandleSetAuthor(botCtx)
			case "SetAnonymsMode":
				var user models.User
				db.DB.Where(&models.User{TelegramID: botCtx.UserID}).First(&user)
				user.SetAnonymsMode(db.DB)
				context.UpdateUserName(botCtx, "start")
				start.HandleSettings(botCtx)
			case "NewJoke":
				time, IsCooldow := Utilities.GetRemainingCooldown(uint(botCtx.UserID))
				if IsCooldow {
					alert := tgbotapi.NewCallbackWithAlert(botCtx.CallbackQuery.ID, "Вы сможете шуткануть через "+time)
					alert.ShowAlert = false
					botCtx.Ctx.BotAPI.Request(alert)
					return
				}
				context.UpdateUserName(botCtx, "JokeMenu")
				MenuJokes.NewJokeHandle(botCtx)
			case "ViewJokes":
				context.UpdateUserName(botCtx, "JokeMenu")
				MenuJokes.HandleJokeViewer(botCtx)
			case "MyJokes":
				context.UpdateUserName(botCtx, "JokeMenu")
				MenuJokes.HandleMyJokes(botCtx)

			default:
				state.MessageID = 0
				context.ClearAllUserData(botCtx)
				start.HandleStartCommand(botCtx)
				return
			}
		}
		if botCtx.Message != nil {
			state.MessageID = 0
			context.ClearAllUserData(botCtx)
			start.HandleStartCommand(botCtx)
			return
		}
	}

}
