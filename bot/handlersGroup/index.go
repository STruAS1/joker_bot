package handlersgroup

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

func HandleUpdate(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	if update.Message != nil && update.Message.IsCommand() && update.Message.Command() == "start" {
		HandleMenu(bot, update)
		return
	}
}
