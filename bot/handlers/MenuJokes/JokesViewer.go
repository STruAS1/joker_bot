package MenuJokes

import (
	"SHUTKANULbot/Utilities"
	"SHUTKANULbot/bot/context"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func HandleJokeViewer(botCtx *context.BotContext) {
	state := context.GetUserState(botCtx)
	context.UpdateUserLevel(botCtx, 2)
	var rows [][]tgbotapi.InlineKeyboardButton
	var row []tgbotapi.InlineKeyboardButton
	Joke, err := Utilities.GetNextJoke(botCtx.UserID)
	if err != nil {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("В главное меню", "Start")))
		if state.MessageID != 0 {
			msg := tgbotapi.NewEditMessageTextAndMarkup(
				botCtx.UserID,
				state.MessageID,
				"Шуток больше нет",
				tgbotapi.NewInlineKeyboardMarkup(rows...),
			)
			msg.ParseMode = "HTML"
			botCtx.Ctx.BotAPI.Send(msg)
		} else {
			msg := tgbotapi.NewMessage(
				botCtx.UserID,
				"Шуток больше нет",
			)
			msg.ParseMode = "HTML"
			msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
			botCtx.Ctx.BotAPI.Send(msg)
		}
	}
	for i := range 5 {
		row = append(row, tgbotapi.NewInlineKeyboardButtonData(fmt.Sprint(i+1), fmt.Sprintf("Evolution_%d_%d", i+1, Joke.ID)))
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(row...))
	Text := fmt.Sprintf("<b>#%d\n\n</b>", Joke.ID) + Joke.Text
	if !Joke.AnonymsMode && Joke.Author != "" {
		Text += "\n\n<b><i>Автор:</i></b> @" + strings.TrimPrefix(Joke.Author, "@")
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("В главное меню", "Start")))

	if state.MessageID != 0 {
		msg := tgbotapi.NewEditMessageTextAndMarkup(
			botCtx.UserID,
			state.MessageID,
			Text,
			tgbotapi.NewInlineKeyboardMarkup(rows...),
		)
		msg.ParseMode = "HTML"
		botCtx.Ctx.BotAPI.Send(msg)
	} else {
		msg := tgbotapi.NewMessage(
			botCtx.UserID,
			Text,
		)
		msg.ParseMode = "HTML"
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
		botCtx.Ctx.BotAPI.Send(msg)
	}
}
