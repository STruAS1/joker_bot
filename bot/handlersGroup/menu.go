package handlersgroup

import (
	"SHUTKANULbot/Utilities"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func HandleMenu(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	Joke := Utilities.GetRandomPopularJokeSafe()
	Text := Joke.Text
	var rows [][]tgbotapi.InlineKeyboardButton
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonURL("Оценить шутку", fmt.Sprintf("https://t.me/JOKER8BOT?start=joke_%d", Joke.ID)),
	))
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Ещё шутка", "ViewJoke"),
	))
	if !Joke.AnonymsMode && Joke.Author != "" {
		Text += "\n\n<b><i>Автор:</i></b> @" + fmt.Sprintf("%s", strings.TrimPrefix(Joke.Author, "@"))
	}
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, Text)
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
	msg.ParseMode = "HTML"
	bot.Send(msg)
}
