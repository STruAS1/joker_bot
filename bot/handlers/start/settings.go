package start

import (
	TonConnectCallback "SHUTKANULbot/TonConnectCallBack"
	"SHUTKANULbot/Utilities"
	contextBot "SHUTKANULbot/bot/context"
	"SHUTKANULbot/db"
	"SHUTKANULbot/db/models"
	"bytes"
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/skip2/go-qrcode"
)

func HandleSettings(botCtx *contextBot.BotContext) {
	// Обновляем уровень пользователя
	contextBot.UpdateUserLevel(botCtx, 0)
	state := contextBot.GetUserState(botCtx)
	var rows [][]tgbotapi.InlineKeyboardButton

	// Проверяем состояние кошелька через TONconnect
	connected, wallet := TonConnectCallback.IsUserConnected(botCtx.UserID)
	var walletStatus string
	if connected {
		walletStatus = fmt.Sprintf("<b>• Кошелёк:</b>\n <code>%s</code>", wallet)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Отключить кошелёк", "DisconnectWallet"),
		))
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Вывести $JOKER", "Withdraw"),
		))
	} else {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Подключить кошелёк", "ConnectWallet"),
		))
	}

	// Получаем данные пользователя из БД
	var user models.User
	db.DB.Where(&models.User{TelegramID: botCtx.UserID}).First(&user)

	// Кнопка для анонимного режима с галочкой или крестиком
	var anonymsSuffix string
	if user.AnonymsMode {
		anonymsSuffix = "✅"
	} else {
		anonymsSuffix = "❌"
	}

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Указать автора", "SetAuthor"),
		tgbotapi.NewInlineKeyboardButtonData("Анонимный режим "+anonymsSuffix, "SetAnonymsMode"),
	))
	// Кнопка возврата в главное меню
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("« Назад", "StartMenu"),
	))

	text := "<b>⚙️ Настройки</b>\n\n"
	if connected {
		text += walletStatus + "\n\n"
	}
	if user.AuthorUserName != "" {
		text += fmt.Sprintf("• <b>Автор:</b> <i>@%s</i>\n\n", strings.TrimPrefix(user.AuthorUserName, "@"))
	} else {
		text += "• <b>Автор:</b> <i>не указан</i>\n\n"
	}
	var anonText string
	if user.AnonymsMode {
		anonText = "✅\n"
	} else {
		anonText = "❌\n"
	}
	text += "• <b>Анонимный режим:</b> " + anonText + "\n"

	if state.MessageID == 0 {
		msg := tgbotapi.NewMessage(botCtx.UserID, text)
		msg.ParseMode = "HTML"
		msg.DisableWebPagePreview = true
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
		if _, err := botCtx.SendMessage(msg); err != nil {
			fmt.Print(err)
		}
	} else {
		msg := tgbotapi.NewEditMessageTextAndMarkup(botCtx.UserID, state.MessageID, text, tgbotapi.NewInlineKeyboardMarkup(rows...))
		msg.ParseMode = "HTML"
		msg.DisableWebPagePreview = true
		if _, err := botCtx.Ctx.BotAPI.Send(msg); err != nil {
			fmt.Print(err)
		}
	}
}

func HandleTonConnect(botCtx *contextBot.BotContext) {
	contextBot.UpdateUserLevel(botCtx, 3)
	state := contextBot.GetUserState(botCtx)
	deleteCfg := tgbotapi.DeleteMessageConfig{
		ChatID:    botCtx.UserID,
		MessageID: state.MessageID,
	}
	botCtx.Ctx.BotAPI.Send(deleteCfg)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	state.Data["ChancelConnect"] = cancel
	Wallets, deeplink, _ := TonConnectCallback.GenerateWalletLinks(botCtx, ctx)
	var qrBuffer bytes.Buffer
	qrCode, err := qrcode.New(deeplink, qrcode.Medium)
	if err != nil {
		log.Println("Ошибка генерации QR-кода:", err)
		return
	}
	err = qrCode.Write(256, &qrBuffer)
	if err != nil {
		log.Println("Ошибка кодирования QR-кода:", err)
		return
	}
	var rows [][]tgbotapi.InlineKeyboardButton
	for Wallet, link := range Wallets {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL(Wallet, link),
		))
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("« Назад", "back"),
	))
	var Text string = "Подключи свой TON-кошелёк, нажав кнопку ниже: "
	photo := tgbotapi.NewPhoto(botCtx.UserID, tgbotapi.FileBytes{
		Name:  "qrcode.png",
		Bytes: qrBuffer.Bytes(),
	})
	photo.Caption = Text
	photo.ParseMode = "HTML"
	photo.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
	msge, _ := botCtx.Ctx.BotAPI.Send(photo)
	state.MessageID = msge.MessageID
}

type WithdrawDataType struct {
	ActiveStep uint
	Amount     uint64
}

func HandleWithdraw(botCtx *contextBot.BotContext) {
	var msgText string
	if botCtx.Message != nil {
		msgText = botCtx.Message.Text
	}
	contextBot.UpdateUserLevel(botCtx, 2)
	connected, _ := TonConnectCallback.IsUserConnected(botCtx.UserID)
	if !connected {
		HandleTonConnect(botCtx)
		return
	}
	var user models.User
	db.DB.Where(&models.User{TelegramID: botCtx.UserID}).First(&user)
	state := contextBot.GetUserState(botCtx)
	WithdrawData, exist := state.Data["WithdrawActiveStep"].(WithdrawDataType)
	if !exist {
		WithdrawData = WithdrawDataType{}
	}
	var rows [][]tgbotapi.InlineKeyboardButton
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Отмена", "back"),
	))
	switch WithdrawData.ActiveStep {
	case 0:
		if botCtx.CallbackQuery.Data == "Withdraw" {
			var Text string = fmt.Sprintf("💰 <b>Баланс: <code>%s</code> $JOKER</b>\n", fmt.Sprintf("%f", float64((uint64(user.Balance/1_000_000)))/1000))
			Text += "Введите сумму вывода"
			if user.Balance == 0 {
				if state.MessageID == 0 {
					msg := tgbotapi.NewMessage(botCtx.UserID, "Эй, бро, у тебя на счету 0. Сначала заработай!")
					msg.ParseMode = "HTML"
					msg.DisableWebPagePreview = true
					msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
					botCtx.SendMessage(msg)
				} else {
					msg := tgbotapi.NewEditMessageTextAndMarkup(botCtx.UserID, state.MessageID, "Эй, бро, у тебя на счету 0. Сначала заработай!", tgbotapi.NewInlineKeyboardMarkup(rows...))
					msg.DisableWebPagePreview = true
					msg.ParseMode = "HTML"
					botCtx.Ctx.BotAPI.Send(msg)
				}
				return
			}
			if state.MessageID == 0 {
				msg := tgbotapi.NewMessage(botCtx.UserID, Text)
				msg.ParseMode = "HTML"
				msg.DisableWebPagePreview = true
				msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
				botCtx.SendMessage(msg)
			} else {
				msg := tgbotapi.NewEditMessageTextAndMarkup(botCtx.UserID, state.MessageID, Text, tgbotapi.NewInlineKeyboardMarkup(rows...))
				msg.DisableWebPagePreview = true
				msg.ParseMode = "HTML"
				botCtx.Ctx.BotAPI.Send(msg)
			}
			WithdrawData.ActiveStep++
		}
	case 1:
		if botCtx.Message != nil {
			amount_float, err := strconv.ParseFloat(msgText, 64)
			if err != nil {
				msg := tgbotapi.NewMessage(botCtx.UserID, "Введите корректную сумму вывода")
				msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
				msg.DisableWebPagePreview = true
				msg.ParseMode = "HTML"
				if _, err := botCtx.SendMessage(msg); err != nil {
					fmt.Print(err)
				}
				return
			}
			AmountUint64 := uint64(amount_float * 1_000_000_000)
			fmt.Print(AmountUint64, '\n')
			fmt.Print(user.Balance, '\n')
			if AmountUint64 > user.Balance {
				msg := tgbotapi.NewMessage(botCtx.UserID, "Недостаточно средств")
				msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
				msg.DisableWebPagePreview = true
				msg.ParseMode = "HTML"
				if _, err := botCtx.SendMessage(msg); err != nil {
					fmt.Print(err)
				}
				return
			}
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			state.Data["ChancelTransaction"] = cancel
			WithdrawData.Amount = AmountUint64
			msg := tgbotapi.NewMessage(botCtx.UserID, "Подтвердите транзакцию в своём кошельке")
			msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
			msg.DisableWebPagePreview = true
			msg.ParseMode = "HTML"
			if _, err := botCtx.SendMessage(msg); err != nil {
				fmt.Print(err)
			}
			go TonConnectCallback.SendTokensViaTonConnect(botCtx, AmountUint64, ctx)
		}
	}
	state.Data["WithdrawActiveStep"] = WithdrawData
}

type SetAuthorDataType struct {
	ActiveStep uint
	Author     string
}

func HandleSetAuthor(botCtx *contextBot.BotContext) {
	state := contextBot.GetUserState(botCtx)
	var msgText string
	contextBot.UpdateUserLevel(botCtx, 4)
	if botCtx.Message != nil {
		msgText = botCtx.Message.Text
	}
	SetAuthorData, exist := state.Data["SetAuthorData"].(SetAuthorDataType)
	if !exist {
		SetAuthorData = SetAuthorDataType{}
	}
	var rows [][]tgbotapi.InlineKeyboardButton
	switch SetAuthorData.ActiveStep {
	case 0:
		if botCtx.CallbackQuery != nil && botCtx.CallbackQuery.Data == "SetAuthor" {
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Отмена", "back"),
			))
			var Text string = "Отправь любой юзернейм канала, пользователя или бота"
			if state.MessageID == 0 {
				msg := tgbotapi.NewMessage(botCtx.UserID, Text)
				msg.ParseMode = "HTML"
				msg.DisableWebPagePreview = true
				msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
				botCtx.SendMessage(msg)
			} else {
				msg := tgbotapi.NewEditMessageTextAndMarkup(botCtx.UserID, state.MessageID, Text, tgbotapi.NewInlineKeyboardMarkup(rows...))
				msg.DisableWebPagePreview = true
				msg.ParseMode = "HTML"
				botCtx.Ctx.BotAPI.Send(msg)
			}
			SetAuthorData.ActiveStep++
		}
	case 1:
		if botCtx.Message != nil {
			valid, _ := Utilities.IsUsernameValid(msgText)
			if !valid {
				rows = append(rows, tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("Отмена", "back"),
				))
				msg := tgbotapi.NewMessage(botCtx.UserID, "Отправь, пожалуйста, реальный юзернейм")
				msg.ParseMode = "HTML"
				msg.DisableWebPagePreview = true
				msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
				botCtx.SendMessage(msg)
				state.Data["SetAuthorData"] = SetAuthorData
				return
			}
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Сохранить", "Save"),
			))
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Отмена", "back"),
			))
			SetAuthorData.ActiveStep++
			SetAuthorData.Author = msgText

			msg := tgbotapi.NewMessage(botCtx.UserID, fmt.Sprintf("Теперь под каждой шуткой будет написан ваш юзернейм: \n\n@%s", strings.TrimPrefix(msgText, "@")))
			msg.ParseMode = "HTML"
			msg.DisableWebPagePreview = true
			msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
			botCtx.SendMessage(msg)
		}
	case 2:
		if botCtx.CallbackQuery != nil && botCtx.CallbackQuery.Data == "Save" {
			var user models.User
			db.DB.Where(&models.User{TelegramID: botCtx.UserID}).First(&user)
			user.AuthorUserName = SetAuthorData.Author
			db.DB.Save(user)
			delete(state.Data, "SetAuthorData")
			HandleSettings(botCtx)
			return
		}
	}
	state.Data["SetAuthorData"] = SetAuthorData
}
