package start

import (
	"SHUTKANULbot/Utilities"
	"SHUTKANULbot/bot/context"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Doc struct {
	Text      string
	PhotoName string
}

var Docs []Doc = []Doc{
	{"<b>JokerBot</b> — это <b>Telegram-бот</b>, где пользователи соревнуются в юморе, публикуют шутки, получают оценки и <i>зарабатывают токены!</i> 🎭💰🎁", "firstImage.jpg"},
	{("<b>🎭 Публикуй шутки — получай оценки!</b>" +
		"\n<blockquote><i>Каждый день у тебя есть 2 попытки, чтобы рассмешить других! Лучшие шутки попадают в топ!</i></blockquote>" +
		"\n\n<b>🏆 Попади в топ и продвигай себя!</b>" +
		"\n<blockquote><i>- Шутка с самым высоким рейтингом дня публикуется в главном канале." +
		"\n- В твоей шутке может быть указана ссылка на твой Telegram-канал или аккаунт. Это отличный способ набрать подписчиков и привлечь внимание к твоим профилям!</i></blockquote>" +
		"\n\n<b>💬 Добавляй бота в группы</b>" +
		"\n<blockquote><i>- Ты можешь добавить этого бота в группы, и он будет взаимодействовать с участниками.</i></blockquote>"),
		"SecondImage.jpg"},
	{"\n\n<b>🪙Давай по фактам:</b>" +
		"\n<i><blockquote>- Ты пишешь шутки, их оценивают — и что? Получаешь токены, которые нихрена не стоят. 💸" +
		"\n- Пригласил друга? Да, тебе капнет 1% с его токенов, но смысл, если это бесполезные цифры? 🫠" +
		"\n- Оцениваешь шутки, участвуешь в розыгрышах — и всё ради чего? Ради воздуха. 💨</blockquote></i>" +
		"\n\n<b>🃏 Ого, можно делать NFT!</b>" +
		"\n<i><blockquote>- Да кому нахрен нужны твои NFT-шутки? 😂 Куда ты их денешь? Кто их купит? Никто.</blockquote></i>" +
		"\n\n<b>💰 Но можно вывести токены!</b>" +
		"\n<i><blockquote>- Ага, выводи. Только курс – говно, комиссии сожрут половину, а спроса нет. В итоге ни денег, ни времени, ни смысла. 🤡</blockquote></i>", "LastImage.jpg"},
}

func HandleDocs(botCtx *context.BotContext) {
	state := context.GetUserState(botCtx)
	context.UpdateUserLevel(botCtx, 1)
	ActiveStep, exist := state.Data["DocsActiveStep"].(int)
	if !exist {
		ActiveStep = 0
	}
	var rows [][]tgbotapi.InlineKeyboardButton
	var callback string
	var button string
	if len(Docs) == ActiveStep+1 {
		callback = "StartMenu"
		button = "Главное меню"
		delete(state.Data, "DocsActiveStep")
	} else {
		callback = "Docs"
		button = "Дальше"
		state.Data["DocsActiveStep"] = ActiveStep + 1
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(button, callback)))
	Doc := Docs[ActiveStep]
	if state.MessageID != 0 {
		DeleteMessageConfig := tgbotapi.DeleteMessageConfig{
			ChatID:    botCtx.UserID,
			MessageID: state.MessageID,
		}
		botCtx.Ctx.BotAPI.Send(DeleteMessageConfig)
	}
	msg := tgbotapi.NewPhoto(botCtx.UserID, tgbotapi.FileID(Utilities.GetPhotoId(botCtx, Doc.PhotoName)))
	msg.Caption = Doc.Text
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
	msg.ParseMode = "HTML"
	photoMSG, err := botCtx.Ctx.BotAPI.Send(msg)
	if err != nil {
		fmt.Print(err)
	}
	state.MessageID = photoMSG.MessageID
}
