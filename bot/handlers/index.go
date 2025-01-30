package handlers

import (
	// "SHUTKANULbot/bot/handlers/MenuJokes"
	"SHUTKANULbot/bot/context"
	"SHUTKANULbot/bot/handlers/MenuJokes"
	"SHUTKANULbot/bot/handlers/start"
	"strings"
)

var nameHandlers = map[string]func(*context.BotContext){
	"start":    start.Handle,
	"JokeMenu": MenuJokes.Handle,
	// "ads":          ads.Handle,
	// "profile":      profile.Handle,
	// "Verification": verification.Handle,
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
			// case "adsMenu":
			// 	context.UpdateUserName(userId, ctx, "ads")
			// 	ads.HandleMenu(update, ctx)
			case "StartMenu":
				context.UpdateUserName(botCtx, "start")
				start.HandleStartCommand(botCtx)
			case "NweJoke":
				context.UpdateUserName(botCtx, "JokeMenu")
				MenuJokes.NewJokeHandle(botCtx)
			case "ViewJokes":
				context.UpdateUserName(botCtx, "JokeMenu")
				MenuJokes.HandleJokeViewer(botCtx)
				// case "AddAds":
				// 	ads.HandleSelectADS(update, ctx)
				// case "AdsHistory":
				// 	ads.HandleSelectADSHistory(update, ctx)
				// case "profile":
				// 	context.UpdateUserName(userId, ctx, "profile")
				// 	profile.HandleProfile(update, ctx)
				// case "+balance":
				// 	profile.HandleSelectPaymentMetod(update, ctx)
				// case "Docs":
				// 	start.HandleDocs(update, ctx)
				// case "Transfer":
				// 	profile.HandleDoPayment(update, ctx)
				// case "Verification":
				// 	if len(strings.Split(update.CallbackQuery.Data, "_")) == 2 && strings.Split(update.CallbackQuery.Data, "_")[1] == strconv.Itoa(state.MessageID) {
				// 		context.UpdateUserName(userId, ctx, "Verification")
				// 		verification.HandleVerification(update, ctx)
				// 	}
			}

		}
	}

}
